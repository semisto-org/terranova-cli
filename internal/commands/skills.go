package commands

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

// Les skills voyagent DANS le binaire (ISC-386) : pas de fichier à télécharger,
// `terranova skill install` les pose où l'agent les lit.
//
//go:embed skills_embedded/*/SKILL.md
var embeddedSkills embed.FS

const skillsRoot = "skills_embedded"

func init() {
	cli.Register(&cli.Command{
		Name: "skill", Group: "Additional",
		Summary: "Skills agent embarqués : lister, afficher, installer (ISC-386).",
		Sub: []*cli.Command{
			{
				Name: "list", Summary: "Les skills embarqués dans ce binaire.",
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					names, err := skillNames()
					if err != nil {
						return nil, err
					}
					return &cli.Result{Data: names, Summary: fmt.Sprintf("%d skill(s) embarqués.", len(names)),
						Crumbs: []cli.Crumb{{Action: "installer", Cmd: "terranova skill install"}}}, nil
				},
			},
			{
				Name: "show", Summary: "Affiche un skill.", ArgSpec: "<nom>", MinArgs: 1,
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					raw, err := embeddedSkills.ReadFile(skillsRoot + "/" + args[0] + "/SKILL.md")
					if err != nil {
						return nil, cli.Usagef("skill inconnu : %s (vois `terranova skill list`)", args[0])
					}
					fmt.Print(string(raw))
					return nil, nil
				},
			},
			{
				Name: "install", Summary: "Pose les skills dans ~/.claude/skills/ (ou --dir <chemin>).",
				Flags: []cli.Flag{{Name: "dir", Arg: "chemin", Help: "Destination (défaut ~/.claude/skills)."}},
				Run:   runSkillInstall,
			},
		},
	})

	cli.Register(&cli.Command{
		Name: "setup", Group: "Auth & Config",
		Summary: "Installe le plugin agent : claude ou codex (ISC-387).",
		ArgSpec: "<claude|codex>", MinArgs: 1,
		Run: runSetup,
	})
}

func skillNames() ([]string, error) {
	entries, err := embeddedSkills.ReadDir(skillsRoot)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func runSkillInstall(c *cli.Ctx, args []string) (*cli.Result, error) {
	dir, _ := cli.FlagValue(args, "dir")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(home, ".claude", "skills")
	}
	names, err := skillNames()
	if err != nil {
		return nil, err
	}
	installed := []string{}
	for _, name := range names {
		raw, err := embeddedSkills.ReadFile(skillsRoot + "/" + name + "/SKILL.md")
		if err != nil {
			return nil, err
		}
		dest := filepath.Join(dir, name)
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dest, "SKILL.md"), raw, 0o644); err != nil {
			return nil, err
		}
		installed = append(installed, dest)
	}
	return &cli.Result{Data: installed, Summary: fmt.Sprintf("%d skill(s) posés dans %s.", len(installed), dir)}, nil
}

func runSetup(c *cli.Ctx, args []string) (*cli.Result, error) {
	switch args[0] {
	case "claude":
		return runSkillInstall(c, nil)
	case "codex":
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		return runSkillInstall(c, []string{"--dir", filepath.Join(home, ".codex", "skills")})
	}
	return nil, cli.Usagef("cible inconnue : %s (claude|codex)", args[0])
}
