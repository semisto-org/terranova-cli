// Package commands — la surface du binaire. Chaque commande se déclare au
// registre ; le catalogue, la surface et la complétion en découlent.
package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/api"
	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/config"
	"golang.org/x/term"
)

// Version est injectée par main (elle-même posée par -ldflags).
var Version = "dev"

func init() {
	cli.Register(&cli.Command{
		Name: "version", Group: "Auth & Config",
		Summary: "Affiche la version du binaire.",
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			return &cli.Result{Data: map[string]string{"version": Version}}, nil
		},
	})

	cli.Register(&cli.Command{
		Name: "auth", Group: "Auth & Config",
		Summary: "Jeton d'accès : login, status, logout, token.",
		Sub: []*cli.Command{
			{
				Name:    "login",
				Summary: "Range un jeton API (trousseau système, repli fichier 0600).",
				Flags: []cli.Flag{
					{Name: "token", Arg: "jeton", Help: "Le jeton — sinon demandé au clavier (saisie masquée)."},
					{Name: "hub", Arg: "id", Help: "Hub par défaut du profil."},
				},
				AgentHelp: "Émettre le jeton dans l'app : Compte & réglages → Jetons CLI (/account/api_tokens). En mode --agent, --token est obligatoire (pas de prompt).",
				Run:       runAuthLogin,
			},
			{
				Name:    "status",
				Summary: "Dit qui l'on est, sur quel hub, avec quels droits.",
				APIOps:  []string{"GET /me"},
				Run:     runAuthStatus,
			},
			{
				Name:    "logout",
				Summary: "Purge le jeton du profil courant.",
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					if err := config.DeleteToken(c.Profile); err != nil {
						return nil, err
					}
					return &cli.Result{Data: map[string]any{"profile": c.Profile, "logged_out": true},
						Summary: fmt.Sprintf("Jeton du profil %q purgé.", c.Profile)}, nil
				},
			},
			{
				Name:    "token",
				Summary: "Imprime le jeton pour les scripts (jamais dans les logs).",
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					token, err := config.Token(c.Profile)
					if err != nil {
						return nil, err
					}
					fmt.Println(token)
					return nil, nil
				},
			},
		},
	})

	cli.Register(&cli.Command{
		Name: "me", Group: "Auth & Config",
		Summary: "Identité, hubs et rôles, grants, scopes effectifs du jeton (« qu'est-ce que j'ai le droit de faire ici »).",
		APIOps:  []string{"GET /me"},
		Run:     runMe,
	})

	cli.Register(&cli.Command{
		Name: "hubs", Group: "Auth & Config",
		Summary: "La constellation : lister les hubs, choisir le hub par défaut.",
		APIOps:  []string{"GET /hubs", "GET /hubs/{id}"},
		Sub: []*cli.Command{
			{
				Name: "list", Summary: "Liste MES hubs — la constellation telle que je la vois (ISC-424).",
				APIOps:    []string{"GET /me"},
				AgentHelp: "La liste vient de /me : un membre voit ses hubs, pas ceux du réseau. La vue superadmin cross-hubs vit sous `network hubs`.",
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out struct {
						Me struct {
							Hubs []map[string]any `json:"hubs"`
						} `json:"me"`
					}
					if err := client.Get("/me", &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, h := range out.Me.Hubs {
						def := ""
						if b, ok := h["default"].(bool); ok && b {
							def = "✓"
						}
						rows = append(rows, []string{fmt.Sprintf("%v", h["id"]), str(h["name"]), str(h["role"]), def})
					}
					return &cli.Result{Data: out.Me.Hubs, Headers: []string{"ID", "NOM", "RÔLE", "DÉFAUT"}, Rows: rows,
						Summary: fmt.Sprintf("%d hub(s).", len(out.Me.Hubs)),
						Crumbs:  []cli.Crumb{{Action: "choisir le hub par défaut", Cmd: "terranova hubs use <id>"}}}, nil
				},
			},
			{
				Name: "use", Summary: "Persiste le hub par défaut du profil (ISC-375).",
				ArgSpec: "<id>", MinArgs: 1,
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					p := c.Config.Profile(c.Profile)
					p.Hub = args[0]
					if c.Config.DefaultProfile == "" {
						c.Config.DefaultProfile = c.Profile
					}
					if err := c.Config.Save(); err != nil {
						return nil, err
					}
					return &cli.Result{Data: map[string]any{"profile": c.Profile, "hub": args[0]},
						Summary: fmt.Sprintf("Hub %s par défaut pour le profil %q.", args[0], c.Profile)}, nil
				},
			},
		},
	})

	cli.Register(&cli.Command{
		Name: "api", Group: "Additional",
		Summary: "Échappatoire brute : GET/POST/PATCH/PUT/DELETE sur un chemin de l'API (ISC-393).",
		ArgSpec: "<get|post|patch|put|delete> <chemin> [json]", MinArgs: 2,
		AgentHelp: "Le corps JSON se passe en 3e argument ou sur stdin. Si une tâche exige cette commande, il manque une vraie commande — le signaler.",
		Run:       runAPI,
	})

	cli.Register(&cli.Command{
		Name: "commands", Group: "Additional",
		Summary: "Le catalogue complet de la surface (--json pour les agents).",
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			type entry struct {
				Cmd     string   `json:"cmd"`
				Group   string   `json:"group,omitempty"`
				Summary string   `json:"summary"`
				Args    string   `json:"args,omitempty"`
				APIOps  []string `json:"api_ops,omitempty"`
			}
			list := []entry{}
			rows := [][]string{}
			cli.Walk(func(path string, cmd *cli.Command) {
				list = append(list, entry{Cmd: path, Group: cmd.Group, Summary: cmd.Summary, Args: cmd.ArgSpec, APIOps: cmd.APIOps})
				rows = append(rows, []string{path, cmd.Summary})
			})
			return &cli.Result{Data: list, Headers: []string{"COMMANDE", "RÉSUMÉ"}, Rows: rows,
				Summary: fmt.Sprintf("%d commandes.", len(list))}, nil
		},
	})
}

func runAuthLogin(c *cli.Ctx, args []string) (*cli.Result, error) {
	token, args := cli.FlagValue(args, "token")
	hub, _ := cli.FlagValue(args, "hub")
	if token == "" {
		if c.Flags.Agent {
			return nil, cli.Usagef("en mode --agent, --token est obligatoire (aucun prompt)")
		}
		fmt.Fprint(os.Stderr, "Jeton API (Compte & réglages → Jetons CLI) : ")
		if term.IsTerminal(int(os.Stdin.Fd())) {
			raw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return nil, err
			}
			token = strings.TrimSpace(string(raw))
		} else {
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			token = strings.TrimSpace(line)
		}
	}
	if token == "" {
		return nil, cli.Usagef("aucun jeton fourni")
	}

	// Validation AVANT stockage : un jeton mort rangé est un piège différé.
	probe := api.Bare(c.Config.BaseURL(c.Profile), token, hub)
	var me struct {
		Me map[string]any `json:"me"`
	}
	if err := probe.Get("/me", &me); err != nil {
		return nil, fmt.Errorf("jeton refusé par %s : %w", c.Config.BaseURL(c.Profile), err)
	}
	if err := config.SetToken(c.Profile, token); err != nil {
		return nil, err
	}
	p := c.Config.Profile(c.Profile)
	if hub != "" {
		p.Hub = hub
	}
	if c.Config.DefaultProfile == "" {
		c.Config.DefaultProfile = c.Profile
	}
	if err := c.Config.Save(); err != nil {
		return nil, err
	}
	name := str(me.Me["name"])
	return &cli.Result{Data: map[string]any{"profile": c.Profile, "user": me.Me["name"], "hub": me.Me["current_hub"]},
		Summary: fmt.Sprintf("Connecté : %s (profil %q).", name, c.Profile),
		Crumbs: []cli.Crumb{
			{Action: "voir ses droits", Cmd: "terranova me"},
			{Action: "lister les hubs", Cmd: "terranova hubs list"},
		}}, nil
}

func runAuthStatus(c *cli.Ctx, args []string) (*cli.Result, error) {
	return runMe(c, args)
}

func runMe(c *cli.Ctx, args []string) (*cli.Result, error) {
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	var out struct {
		Me map[string]any `json:"me"`
	}
	if err := client.Get("/me", &out); err != nil {
		return nil, err
	}
	me := out.Me
	summary := fmt.Sprintf("%s · hub %s", str(me["name"]), str(dig(me, "current_hub", "name")))
	return &cli.Result{Data: me, Summary: summary,
		Crumbs: []cli.Crumb{
			{Action: "lister les projets", Cmd: "terranova projects list"},
			{Action: "changer de hub", Cmd: "terranova hubs use <id>"},
		}}, nil
}

func runAPI(c *cli.Ctx, args []string) (*cli.Result, error) {
	method := strings.ToUpper(args[0])
	path := args[1]
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	var body any
	if len(args) > 2 {
		if err := json.Unmarshal([]byte(args[2]), &body); err != nil {
			return nil, cli.Usagef("corps JSON invalide : %v", err)
		}
	}
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	var out any
	if err := client.Do(method, path, body, &out); err != nil {
		return nil, err
	}
	return &cli.Result{Data: out}, nil
}

// ── petits utilitaires partagés ────────────────────────────────────────────

func str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func dig(m map[string]any, keys ...string) any {
	var cur any = m
	for _, k := range keys {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = obj[k]
	}
	return cur
}
