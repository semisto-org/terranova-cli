package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/semisto-org/terranova-cli/internal/api"
	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/config"
)

// ── ISC-392 — le snapshot de surface : tout changement est un diff en revue ──

func TestSurfaceSnapshotIsCommitted(t *testing.T) {
	want, err := os.ReadFile("../../.surface")
	if err != nil {
		t.Fatalf(".surface absent : %v — lance `terranova surface > .surface`", err)
	}
	got := SurfaceSnapshot()
	if string(want) != got {
		if os.Getenv("UPDATE_SURFACE") == "1" {
			if err := os.WriteFile("../../.surface", []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Log(".surface mis à jour")
			return
		}
		t.Fatalf("la surface a changé — relance avec UPDATE_SURFACE=1 pour l'assumer en diff de revue")
	}
}

// ── ISC-390 — la couverture d'API : chaque opération du spec a une commande
// ou une exemption NOMMÉE dans API-COVERAGE.md ──

func TestEveryAPIOperationIsCoveredOrExempted(t *testing.T) {
	raw, err := os.ReadFile("testdata/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}

	covered := AllAPIOps()
	exempted := loadExemptions(t)

	missing := []string{}
	for path, verbs := range spec.Paths {
		for verb := range verbs {
			op := strings.ToUpper(verb) + " " + path
			if covered[op] || exempted[op] {
				continue
			}
			missing = append(missing, op)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("opérations d'API sans commande NI exemption dans API-COVERAGE.md :\n  %s",
			strings.Join(missing, "\n  "))
	}
}

// Les exemptions se déclarent en ligne de tableau : `| VERB /path | motif |`.
func loadExemptions(t *testing.T) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	raw, err := os.ReadFile("../../API-COVERAGE.md")
	if err != nil {
		return out
	}
	re := regexp.MustCompile(`(?m)^\|\s*([A-Z]+ /\S+)\s*\|`)
	for _, m := range re.FindAllStringSubmatch(string(raw), -1) {
		out[m[1]] = true
	}
	return out
}

// L'inverse : une commande qui déclare couvrir une opération ABSENTE du spec
// ment (ou le spec a reculé) — les deux méritent un rouge.
func TestDeclaredOpsExistInSpec(t *testing.T) {
	raw, err := os.ReadFile("testdata/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}
	specOps := map[string]bool{}
	for path, verbs := range spec.Paths {
		for verb := range verbs {
			specOps[strings.ToUpper(verb)+" "+path] = true
		}
	}
	// Routes de gestes de la spine : livrées côté app (ISC-440) mais pas encore
	// documentées par le générateur OpenAPI — écart connu, consigné dans
	// API-COVERAGE.md § « Gestes non documentés au spec ».
	knownUndocumented := regexp.MustCompile(`^/recordings/\{id\}/`)
	for op := range AllAPIOps() {
		parts := strings.SplitN(op, " ", 2)
		if knownUndocumented.MatchString(parts[1]) {
			continue
		}
		if !specOps[op] {
			t.Errorf("la surface déclare %q, absent du spec OpenAPI", op)
		}
	}
}

// ── ISC-384 — les breadcrumbs sont exécutables : on les REJOUE ──

func TestBreadcrumbsResolveToRealCommands(t *testing.T) {
	server := stubServer()
	defer server.Close()
	ctx := stubCtx(t, server.URL)

	// Un échantillon de commandes qui émettent des crumbs.
	samples := [][]string{
		{"todos", "list"},
		{"hubs", "list"},
		{"me"},
		{"url", "https://app.semisto.org/projects/19"},
	}
	seen := []cli.Crumb{}
	for _, sample := range samples {
		cmd, rest := cli.Find(sample)
		if cmd == nil || cmd.Run == nil {
			t.Fatalf("échantillon introuvable : %v", sample)
		}
		res, err := cmd.Run(ctx, rest)
		if err != nil {
			t.Fatalf("%v : %v", sample, err)
		}
		if res != nil {
			seen = append(seen, res.Crumbs...)
		}
	}
	if len(seen) == 0 {
		t.Fatal("aucun breadcrumb émis par l'échantillon — le test ne prouve rien")
	}
	for _, crumb := range seen {
		assertCrumbResolves(t, crumb)
	}
}

func assertCrumbResolves(t *testing.T, crumb cli.Crumb) {
	t.Helper()
	words := strings.Fields(crumb.Cmd)
	if len(words) == 0 || words[0] != "terranova" {
		t.Errorf("crumb %q : doit commencer par `terranova`", crumb.Cmd)
		return
	}
	// On garde les mots de commande : ni <placeholders>, ni --drapeaux, ni ids.
	path := []string{}
	for _, w := range words[1:] {
		if strings.HasPrefix(w, "<") || strings.HasPrefix(w, "-") || regexp.MustCompile(`^\d`).MatchString(w) {
			break
		}
		path = append(path, w)
	}
	cmd, rest := cli.Find(path)
	if cmd == nil {
		t.Errorf("crumb %q : la commande n'existe pas dans la surface", crumb.Cmd)
		return
	}
	if len(rest) > 0 {
		t.Errorf("crumb %q : le chemin %v ne se résout pas entièrement (reste %v)", crumb.Cmd, path, rest)
	}
}

// ── ISC-385 — l'aide agent et le catalogue tiennent leurs promesses ──

func TestAgentHelpAndCatalog(t *testing.T) {
	help := cli.AgentHelpFor("", nil)
	cmds, ok := help["commands"].([]map[string]any)
	if !ok || len(cmds) < 500 {
		t.Fatalf("catalogue racine --agent : %d entrées, attendu ≥ 500", len(cmds))
	}
	todo, _ := cli.Find([]string{"todos", "add"})
	if todo == nil {
		t.Fatal("todos add introuvable")
	}
	detail := cli.AgentHelpFor("todos add", todo)
	if detail["flags"] == nil {
		t.Fatal("todos add --help --agent : pas de flags")
	}
}

// ── plomberie de test ──────────────────────────────────────────────────────

func stubServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"me":{"id":1,"name":"Test","current_hub":{"id":1,"name":"Hub"},"hubs":[{"id":1,"name":"Hub","role":"member","default":true}],"grants":[],"token":{"effective_scopes":["projecto"]}}}`)
	})
	mux.HandleFunc("/recordings", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"recordings":[{"id":42,"recordable_type":"Todo","title":"Tester","status":"active","bucket_id":7}]}`)
	})
	mux.HandleFunc("/recordings/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"recording":{"id":42,"recordable_type":"Todo","title":"Tester"}}`)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"ok","revision":"stub"}`)
	})
	return httptest.NewServer(mux)
}

func stubCtx(t *testing.T, baseURL string) *cli.Ctx {
	t.Helper()
	t.Setenv("TERRANOVA_CONFIG_DIR", t.TempDir())
	t.Setenv("TERRANOVA_NO_KEYCHAIN", "1")
	t.Setenv("TERRANOVA_API_TOKEN", "stub-token")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	return &cli.Ctx{
		Config:  cfg,
		Profile: "test",
		Client:  api.Bare(baseURL, "stub-token", "1"),
		Version: "test",
	}
}
