// Package config — profils nommés (ISC-377), config non-secrète dans
// ~/.config/terranova/config.json, secret dans le trousseau système avec
// repli fichier 0600 (ISC-376). Le secret n'est JAMAIS écrit dans le dépôt
// courant ni dans config.json (ISC-370).
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const DefaultBaseURL = "https://app.semisto.org/api/v1"

type Profile struct {
	BaseURL string `json:"base_url,omitempty"`
	Hub     string `json:"hub,omitempty"` // hub par défaut (X-Hub-Id), persistable (ISC-375)
}

type Config struct {
	DefaultProfile string              `json:"default_profile,omitempty"`
	Profiles       map[string]*Profile `json:"profiles,omitempty"`
	dir            string
}

func Dir() string {
	if d := os.Getenv("TERRANOVA_CONFIG_DIR"); d != "" {
		return d
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(base, "terranova")
}

func Load() (*Config, error) {
	c := &Config{Profiles: map[string]*Profile{}, dir: Dir()}
	raw, err := os.ReadFile(filepath.Join(c.dir, "config.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return c, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(raw, c); err != nil {
		return nil, fmt.Errorf("config.json illisible : %w", err)
	}
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	return c, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, "config.json"), append(raw, '\n'), 0o600)
}

// ResolveProfile — priorité : --profile > TERRANOVA_PROFILE > default_profile > "default".
func (c *Config) ResolveProfile(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("TERRANOVA_PROFILE"); env != "" {
		return env
	}
	if c.DefaultProfile != "" {
		return c.DefaultProfile
	}
	return "default"
}

func (c *Config) Profile(name string) *Profile {
	if p := c.Profiles[name]; p != nil {
		return p
	}
	p := &Profile{}
	c.Profiles[name] = p
	return p
}

func (c *Config) BaseURL(profile string) string {
	if p := c.Profiles[profile]; p != nil && p.BaseURL != "" {
		return strings.TrimRight(p.BaseURL, "/")
	}
	if env := os.Getenv("TERRANOVA_BASE_URL"); env != "" {
		return strings.TrimRight(env, "/")
	}
	return DefaultBaseURL
}

// ── Secret — trousseau système, repli fichier 0600 ─────────────────────────

const keychainService = "terranova-cli"

// SetToken range le jeton. TERRANOVA_NO_KEYCHAIN=1 force le repli fichier
// (utile aux tests et aux machines sans trousseau).
func SetToken(profile, token string) error {
	if useKeychain() {
		if err := keychainSet(profile, token); err == nil {
			return nil
		}
	}
	return fileSet(profile, token)
}

// Token lit le jeton : env (scripts, CI) > trousseau > fichier.
func Token(profile string) (string, error) {
	if t := os.Getenv("TERRANOVA_API_TOKEN"); t != "" {
		return t, nil
	}
	if useKeychain() {
		if t, err := keychainGet(profile); err == nil && t != "" {
			return t, nil
		}
	}
	return fileGet(profile)
}

func DeleteToken(profile string) error {
	if useKeychain() {
		_ = keychainDelete(profile)
	}
	return fileDelete(profile)
}

func useKeychain() bool {
	return runtime.GOOS == "darwin" && os.Getenv("TERRANOVA_NO_KEYCHAIN") == ""
}

func keychainSet(profile, token string) error {
	return exec.Command("security", "add-generic-password", "-U",
		"-s", keychainService, "-a", profile, "-w", token).Run()
}

func keychainGet(profile string) (string, error) {
	out, err := exec.Command("security", "find-generic-password",
		"-s", keychainService, "-a", profile, "-w").Output()
	return strings.TrimSpace(string(out)), err
}

func keychainDelete(profile string) error {
	return exec.Command("security", "delete-generic-password",
		"-s", keychainService, "-a", profile).Run()
}

func credentialsPath() string { return filepath.Join(Dir(), "credentials.json") }

func readCredentials() (map[string]string, error) {
	creds := map[string]string{}
	raw, err := os.ReadFile(credentialsPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return creds, nil
		}
		return nil, err
	}
	return creds, json.Unmarshal(raw, &creds)
}

func writeCredentials(creds map[string]string) error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(credentialsPath(), append(raw, '\n'), 0o600)
}

func fileSet(profile, token string) error {
	creds, err := readCredentials()
	if err != nil {
		return err
	}
	creds[profile] = token
	return writeCredentials(creds)
}

func fileGet(profile string) (string, error) {
	creds, err := readCredentials()
	if err != nil {
		return "", err
	}
	if creds[profile] == "" {
		return "", fmt.Errorf("aucun jeton pour le profil %q — lance `terranova auth login`", profile)
	}
	return creds[profile], nil
}

func fileDelete(profile string) error {
	creds, err := readCredentials()
	if err != nil {
		return err
	}
	delete(creds, profile)
	return writeCredentials(creds)
}
