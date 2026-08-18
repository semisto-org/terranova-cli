// Package api — le client HTTP de l'API privée. Reprises avec backoff sur les
// LECTURES seulement : une écriture non idempotente n'est jamais rejouée — une
// tâche créée deux fois est pire qu'une commande échouée (ISC-380). L'enveloppe
// d'erreur de l'API est rendue telle quelle, jamais reformulée (ISC-379).
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/semisto-org/terranova-cli/internal/config"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

type Client struct {
	BaseURL string
	Token   string
	HubID   string
	Verbose int
	HTTP    *http.Client
}

// Error porte le statut HTTP et l'enveloppe {error:{…}} brute de l'API.
type Error struct {
	Status int
	Code   string
	Body   json.RawMessage
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf(msg.ErrHTTPWithCode, e.Status, e.Code)
	}
	return fmt.Sprintf(msg.ErrHTTP, e.Status)
}

// Bare construit un client sans passer par le stockage — pour valider un jeton
// AVANT de le ranger (auth login).
func Bare(baseURL, token, hub string) *Client {
	return &Client{BaseURL: baseURL, Token: token, HubID: hub, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func New(cfg *config.Config, profile, hubFlag string, verbose int) (*Client, error) {
	token, err := config.Token(profile)
	if err != nil {
		return nil, err
	}
	hub := hubFlag
	if hub == "" {
		if p := cfg.Profiles[profile]; p != nil {
			hub = p.Hub
		}
	}
	return &Client{
		BaseURL: cfg.BaseURL(profile),
		Token:   token,
		HubID:   hub,
		Verbose: verbose,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Do exécute une requête. body non-nil = JSON. Les GET sont rejoués (3 essais,
// backoff exponentiel avec gigue) sur 429/5xx et erreurs réseau ; rien d'autre.
func (c *Client) Do(method, path string, body any, out any) error {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	attempts := 1
	if method == http.MethodGet {
		attempts = 3
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			backoff := time.Duration(250*(1<<i))*time.Millisecond + time.Duration(rand.Intn(200))*time.Millisecond
			time.Sleep(backoff)
		}
		lastErr = c.once(method, path, payload, out)
		if lastErr == nil {
			return nil
		}
		var apiErr *Error
		if ok := asError(lastErr, &apiErr); ok {
			// 4xx (sauf 429) : définitif, on ne rejoue pas.
			if apiErr.Status != 429 && apiErr.Status < 500 {
				return lastErr
			}
		}
	}
	return lastErr
}

func asError(err error, target **Error) bool {
	e, ok := err.(*Error)
	if ok {
		*target = e
	}
	return ok
}

func (c *Client) once(method, path string, payload []byte, out any) error {
	url := c.BaseURL + path
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	c.decorate(req)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Verbose >= 2 {
		fmt.Fprintf(os.Stderr, msg.VerboseRequest, method, url)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if c.Verbose >= 2 {
		fmt.Fprintf(os.Stderr, msg.VerboseResponse, res.StatusCode, len(raw))
	}
	if res.StatusCode >= 400 {
		apiErr := &Error{Status: res.StatusCode, Body: raw}
		var envelope struct {
			Error any `json:"error"`
		}
		if json.Unmarshal(raw, &envelope) == nil {
			switch v := envelope.Error.(type) {
			case string:
				apiErr.Code = v
			case map[string]any:
				if code, ok := v["code"].(string); ok {
					apiErr.Code = code
				}
			}
		}
		return apiErr
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}

// Upload envoie un fichier en multipart (le dépôt de fichier réel d'ISC-400).
func (c *Client) Upload(path string, fields map[string]string, fileField, filePath string, out any) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	part, err := w.CreateFormFile(fileField, filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, f); err != nil {
		return err
	}
	w.Close()
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, &buf)
	if err != nil {
		return err
	}
	c.decorate(req)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return &Error{Status: res.StatusCode, Body: raw}
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func (c *Client) decorate(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("User-Agent", "terranova-cli")
	if c.HubID != "" {
		req.Header.Set("X-Hub-Id", c.HubID)
	}
}

// Get / Post / Patch / Delete — sucre.
func (c *Client) Get(path string, out any) error { return c.Do(http.MethodGet, path, nil, out) }
func (c *Client) Post(path string, body, out any) error {
	return c.Do(http.MethodPost, path, body, out)
}
func (c *Client) Patch(path string, body, out any) error {
	return c.Do(http.MethodPatch, path, body, out)
}
func (c *Client) Delete(path string, out any) error { return c.Do(http.MethodDelete, path, nil, out) }

// QueryEscape aide les commandes à construire leurs chemins.
func Query(params map[string]string) string {
	parts := []string{}
	for k, v := range params {
		if v != "" {
			parts = append(parts, k+"="+strings.ReplaceAll(v, " ", "%20"))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "?" + strings.Join(parts, "&")
}
