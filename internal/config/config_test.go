package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testDir(t *testing.T) {
	t.Helper()
	t.Setenv("TERRANOVA_CONFIG_DIR", t.TempDir())
	t.Setenv("TERRANOVA_NO_KEYCHAIN", "1")
	t.Setenv("TERRANOVA_API_TOKEN", "")
}

func TestProfileResolutionOrder(t *testing.T) {
	testDir(t)
	c := &Config{Profiles: map[string]*Profile{}}
	t.Setenv("TERRANOVA_PROFILE", "")
	if got := c.ResolveProfile(""); got != "default" {
		t.Fatalf("défaut : %q", got)
	}
	c.DefaultProfile = "michael"
	if got := c.ResolveProfile(""); got != "michael" {
		t.Fatalf("default_profile : %q", got)
	}
	t.Setenv("TERRANOVA_PROFILE", "nova")
	if got := c.ResolveProfile(""); got != "nova" {
		t.Fatalf("env : %q", got)
	}
	if got := c.ResolveProfile("agent"); got != "agent" {
		t.Fatalf("drapeau : %q", got)
	}
}

// ISC-370/376 — le secret vit dans credentials.json 0600 (repli), JAMAIS dans
// config.json.
func TestTokenStorageFileFallback(t *testing.T) {
	testDir(t)
	if err := SetToken("p1", "secret-1"); err != nil {
		t.Fatal(err)
	}
	got, err := Token("p1")
	if err != nil || got != "secret-1" {
		t.Fatalf("lecture : %q %v", got, err)
	}
	info, err := os.Stat(filepath.Join(Dir(), "credentials.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("credentials.json doit être 0600, est %v", info.Mode().Perm())
	}
	cfg := &Config{Profiles: map[string]*Profile{"p1": {Hub: "1"}}, dir: Dir()}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(Dir(), "config.json"))
	if strings.Contains(string(raw), "secret-1") {
		t.Fatal("le secret a fui dans config.json")
	}
	if err := DeleteToken("p1"); err != nil {
		t.Fatal(err)
	}
	if _, err := Token("p1"); err == nil {
		t.Fatal("après logout, la lecture doit échouer")
	}
}

func TestEnvTokenWinsForScripts(t *testing.T) {
	testDir(t)
	t.Setenv("TERRANOVA_API_TOKEN", "env-token")
	got, err := Token("nimporte")
	if err != nil || got != "env-token" {
		t.Fatalf("env d'abord : %q %v", got, err)
	}
}
