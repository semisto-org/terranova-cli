package commands

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/config"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

func init() {
	cli.Register(&cli.Command{
		Name: "surface", Group: msg.GroupAdditional,
		Summary: msg.HelpSurface,
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			fmt.Print(SurfaceSnapshot())
			return nil, nil
		},
	})

	cli.Register(&cli.Command{
		Name: "doctor", Group: msg.GroupAuthConfig,
		Summary: msg.HelpDoctor,
		APIOps:  []string{"GET /health", "GET /me"},
		Run:     runDoctor,
	})

	cli.Register(&cli.Command{
		Name: "completion", Group: msg.GroupAdditional,
		Summary: msg.HelpCompletion,
		ArgSpec: "<bash|zsh|fish>", MinArgs: 1,
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			script, err := completionScript(args[0])
			if err != nil {
				return nil, err
			}
			fmt.Print(script)
			return nil, nil
		},
	})

	cli.Register(&cli.Command{
		Name: "url", Group: msg.GroupSearchBrowse,
		Summary: msg.HelpUrl,
		ArgSpec: "<url>", MinArgs: 1,
		Run: runURL,
	})

	cli.Register(&cli.Command{
		Name: "quick-start", Group: msg.GroupAuthConfig,
		Summary: msg.HelpQuickStart,
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			steps := msg.QuickStartSteps
			if !c.Flags.JSON && c.IsTTY {
				for _, s := range steps {
					fmt.Println(s)
				}
				return nil, nil
			}
			return &cli.Result{Data: steps}, nil
		},
	})
}

// SurfaceSnapshot rend la surface complète en texte stable, triée.
func SurfaceSnapshot() string {
	lines := []string{}
	cli.Walk(func(path string, c *cli.Command) {
		entry := path
		if c.ArgSpec != "" {
			entry += " " + c.ArgSpec
		}
		for _, f := range c.Flags {
			entry += " [--" + f.Name
			if f.Arg != "" {
				entry += " <" + f.Arg + ">"
			}
			entry += "]"
		}
		lines = append(lines, entry)
	})
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

// AllAPIOps agrège les opérations d'API couvertes par la surface (gate ISC-390).
func AllAPIOps() map[string]bool {
	ops := map[string]bool{}
	cli.Walk(func(path string, c *cli.Command) {
		for _, op := range c.APIOps {
			ops[op] = true
		}
	})
	return ops
}

func runDoctor(c *cli.Ctx, args []string) (*cli.Result, error) {
	type check struct {
		Name string `json:"name"`
		OK   bool   `json:"ok"`
		Note string `json:"note,omitempty"`
	}
	checks := []check{}
	add := func(name string, ok bool, note string) { checks = append(checks, check{name, ok, note}) }

	add(msg.DoctorBinary, true, fmt.Sprintf(msg.DoctorNoteVersion, c.Version))
	add(msg.DoctorConfig, true, config.Dir())

	token, err := config.Token(c.Profile)
	add(fmt.Sprintf(msg.DoctorTokenProfile, c.Profile), err == nil && token != "", noteIf(err))

	client, err := c.API()
	if err != nil {
		add(msg.DoctorConnection, false, err.Error())
	} else {
		var health map[string]any
		if err := client.Get("/health", &health); err != nil {
			add(fmt.Sprintf(msg.DoctorConnectionTo, client.BaseURL), false, err.Error())
		} else {
			add(fmt.Sprintf(msg.DoctorConnectionTo, client.BaseURL), true, str(health["revision"]))
		}
		var me struct {
			Me map[string]any `json:"me"`
		}
		if err := client.Get("/me", &me); err != nil {
			add(msg.DoctorIdentity, false, err.Error())
		} else {
			add(msg.DoctorIdentity, true, fmt.Sprintf(msg.ResMeSummary, str(me.Me["name"]), str(dig(me.Me, "current_hub", "name"))))
		}
	}

	for _, plugin := range []string{".claude-plugin", ".codex-plugin"} {
		home, _ := os.UserHomeDir()
		_, err := os.Stat(home + "/.claude/plugins/terranova")
		_ = plugin
		add(msg.DoctorPlugin, err == nil, msg.DoctorPluginNote)
		break
	}

	allOK := true
	rows := [][]string{}
	for _, ch := range checks {
		mark := "✓"
		if !ch.OK {
			mark = "✗"
			allOK = false
		}
		rows = append(rows, []string{mark, ch.Name, ch.Note})
	}
	summary := msg.DoctorAllOK
	if !allOK {
		summary = msg.DoctorFailing
	}
	return &cli.Result{Data: map[string]any{"ok": allOK, "checks": checks},
		Headers: msg.HeadersDoctor, Rows: rows, Summary: summary}, nil
}

func noteIf(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// runURL — le pont entre « je regarde une page » et « je l'automatise ».
func runURL(c *cli.Ctx, args []string) (*cli.Result, error) {
	u, err := url.Parse(args[0])
	if err != nil {
		return nil, cli.Usagef(msg.UsageInvalidURL, err)
	}
	path := u.Path
	type match struct {
		re    *regexp.Regexp
		kind  string
		crumb func(ids []string) []cli.Crumb
	}
	matches := []match{
		{regexp.MustCompile(`^/projects/(\d+)`), "project", func(ids []string) []cli.Crumb {
			return []cli.Crumb{
				{Action: msg.CrumbVoirLeProjet, Cmd: "terranova projects show " + ids[0]},
				{Action: msg.CrumbSesTaches, Cmd: "terranova todos list -p " + ids[0]},
				{Action: msg.CrumbSesRecordings, Cmd: "terranova recordings list -p " + ids[0]},
			}
		}},
		{regexp.MustCompile(`^/(?:todos|messages|documents|cards|events|recordings)/(\d+)`), "recording", func(ids []string) []cli.Crumb {
			return []cli.Crumb{
				{Action: msg.CrumbVoir, Cmd: "terranova recordings show " + ids[0]},
				{Action: msg.CrumbCommenter, Cmd: "terranova recordings comment " + ids[0] + " <corps>"},
			}
		}},
		{regexp.MustCompile(`^/(administratio|contacto|planto|academio|conceptio|nurserio)`), "lens", func(ids []string) []cli.Crumb {
			return []cli.Crumb{{Action: msg.CrumbLaLentilleAuCLI, Cmd: "terranova " + ids[0] + " --help"}}
		}},
	}
	for _, m := range matches {
		if got := m.re.FindStringSubmatch(path); got != nil {
			ids := got[1:]
			data := map[string]any{"kind": m.kind, "url": args[0]}
			if len(ids) > 0 {
				data["id"] = ids[0]
			}
			// Enrichir : si c'est un recording, on va le chercher.
			if m.kind == "recording" {
				if client, err := c.API(); err == nil {
					var out struct {
						Recording map[string]any `json:"recording"`
					}
					if client.Get("/recordings/"+ids[0], &out) == nil {
						data["recording"] = out.Recording
					}
				}
			}
			return &cli.Result{Data: data, Crumbs: m.crumb(ids),
				Summary: fmt.Sprintf(msg.ResURLRecognized, m.kind)}, nil
		}
	}
	return nil, fmt.Errorf(msg.ErrURLNotRecognized)
}

// completionScript génère la complétion depuis le registre.
func completionScript(shell string) (string, error) {
	// La table plate commande → sous-commandes.
	subs := map[string][]string{}
	tops := []string{}
	for _, cmd := range cli.Registry {
		tops = append(tops, cmd.Name)
		for _, s := range cmd.Sub {
			subs[cmd.Name] = append(subs[cmd.Name], s.Name)
		}
	}
	sort.Strings(tops)
	switch shell {
	case "bash", "zsh":
		var b strings.Builder
		b.WriteString(fmt.Sprintf(msg.CompletionHeader, shell))
		b.WriteString("_terranova() {\n  local cur prev\n  cur=\"${COMP_WORDS[COMP_CWORD]}\"\n  prev=\"${COMP_WORDS[COMP_CWORD-1]}\"\n")
		b.WriteString("  case \"$prev\" in\n")
		for _, top := range tops {
			if len(subs[top]) > 0 {
				b.WriteString("    " + top + ") COMPREPLY=($(compgen -W \"" + strings.Join(subs[top], " ") + "\" -- \"$cur\")); return;;\n")
			}
		}
		b.WriteString("  esac\n")
		b.WriteString("  COMPREPLY=($(compgen -W \"" + strings.Join(tops, " ") + "\" -- \"$cur\"))\n}\n")
		b.WriteString("complete -F _terranova terranova\n")
		if shell == "zsh" {
			return "autoload -U +X bashcompinit && bashcompinit\n" + b.String(), nil
		}
		return b.String(), nil
	case "fish":
		var b strings.Builder
		for _, top := range tops {
			b.WriteString("complete -c terranova -n '__fish_use_subcommand' -a '" + top + "'\n")
			for _, s := range subs[top] {
				b.WriteString("complete -c terranova -n '__fish_seen_subcommand_from " + top + "' -a '" + s + "'\n")
			}
		}
		return b.String(), nil
	}
	return "", cli.Usagef(msg.UsageUnknownShell, shell)
}
