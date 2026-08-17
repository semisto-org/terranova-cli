package cli

import "testing"

func TestParseGlobalsInterleaved(t *testing.T) {
	f, rest, err := ParseGlobals([]string{"todos", "--json", "list", "-p", "42", "--jq", ".x", "--agent"})
	if err != nil {
		t.Fatal(err)
	}
	if !f.JSON || !f.Agent || !f.Quiet || f.Project != "42" || f.JQ != ".x" {
		t.Fatalf("drapeaux mal lus : %+v", f)
	}
	if len(rest) != 2 || rest[0] != "todos" || rest[1] != "list" {
		t.Fatalf("reste inattendu : %v", rest)
	}
}

func TestParseGlobalsMissingValue(t *testing.T) {
	if _, _, err := ParseGlobals([]string{"--jq"}); err == nil {
		t.Fatal("--jq sans valeur doit échouer")
	}
}

func TestFlagValueBothForms(t *testing.T) {
	v, rest := FlagValue([]string{"a", "--due-on", "2026-09-01", "b"}, "due-on")
	if v != "2026-09-01" || len(rest) != 2 {
		t.Fatalf("forme espacée : %q %v", v, rest)
	}
	v, rest = FlagValue([]string{"--due-on=2026-10-01"}, "due-on")
	if v != "2026-10-01" || len(rest) != 0 {
		t.Fatalf("forme égale : %q %v", v, rest)
	}
}

func TestFindResolvesDeepPaths(t *testing.T) {
	Registry = nil
	Register(&Command{Name: "planto", Sub: []*Command{{Name: "genera", Sub: []*Command{{Name: "list"}}}}})
	cmd, rest := Find([]string{"planto", "genera", "list", "extra"})
	if cmd == nil || cmd.Name != "list" || len(rest) != 1 || rest[0] != "extra" {
		t.Fatalf("résolution : %v %v", cmd, rest)
	}
}
