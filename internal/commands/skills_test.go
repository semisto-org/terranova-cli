package commands

import (
	"regexp"
	"strings"
	"testing"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

// ISC-388 — la dérive des skills : chaque `terranova …` cité par un SKILL.md
// doit exister dans la surface réelle du binaire, sinon rouge.
func TestSkillsDontDrift(t *testing.T) {
	names, err := skillNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 2 {
		t.Fatalf("skills embarqués : %d, attendu ≥ 2", len(names))
	}
	re := regexp.MustCompile("`terranova ([a-z][a-z0-9 -]*)")
	checked := 0
	for _, name := range names {
		raw, err := embeddedSkills.ReadFile(skillsRoot + "/" + name + "/SKILL.md")
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range re.FindAllStringSubmatch(string(raw), -1) {
			words := strings.Fields(m[1])
			path := []string{}
			for _, w := range words {
				if strings.HasPrefix(w, "<") || strings.HasPrefix(w, "-") || strings.HasPrefix(w, "\"") {
					break
				}
				path = append(path, w)
			}
			if len(path) == 0 {
				continue
			}
			cmd, rest := cli.Find(path)
			if cmd == nil {
				t.Errorf("%s cite `terranova %s` : commande inconnue", name, m[1])
				continue
			}
			// Les mots non consommés doivent être des arguments plausibles, pas
			// des sous-commandes ratées : on refuse si le mot restant existe
			// nulle part comme sous-commande mais RESSEMBLE à une (lettres pures).
			if len(rest) > 0 && regexp.MustCompile(`^[a-z-]+$`).MatchString(rest[0]) && len(cmd.Sub) > 0 {
				t.Errorf("%s cite `terranova %s` : %q ne résout pas (sous-commande inconnue ?)", name, m[1], rest[0])
			}
			checked++
		}
	}
	if checked < 20 {
		t.Fatalf("seulement %d commandes citées vérifiées — le test ne mord pas assez", checked)
	}
}
