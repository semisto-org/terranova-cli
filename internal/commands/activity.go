package commands

import (
	"fmt"
	"net/url"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

// ISC-407 (solde) — le fil d'activité, le journal d'un recording, la totale.
func init() {
	cli.Register(&cli.Command{
		Name: "activity", Group: "Search & Browse",
		Summary: "Le fil d'activité du hub (— --project, même règle d'accès que l'écran).",
		APIOps:  []string{"GET /activity"},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			path := "/activity"
			if c.Flags.Project != "" {
				path += "?project_id=" + url.QueryEscape(c.Flags.Project)
			}
			var out struct {
				Events []map[string]any `json:"events"`
			}
			if err := client.Get(path, &out); err != nil {
				return nil, err
			}
			rows := [][]string{}
			for _, e := range out.Events {
				rows = append(rows, []string{str(e["created_at"]), str(e["actor"]), str(e["action"]), str(e["target_type"]) + " " + str(e["target_id"])})
			}
			return &cli.Result{Data: out.Events, Headers: []string{"QUAND", "QUI", "GESTE", "SUR"}, Rows: rows,
				Summary: fmt.Sprintf("%d événement(s).", len(out.Events))}, nil
		},
	})

	cli.Register(&cli.Command{
		Name: "everything", Group: "Search & Browse",
		Summary: "La totale : les comptes par type + les 20 récents.",
		APIOps:  []string{"GET /everything"},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			var out map[string]any
			if err := client.Get("/everything", &out); err != nil {
				return nil, err
			}
			return &cli.Result{Data: out,
				Crumbs: []cli.Crumb{{Action: "creuser un type", Cmd: "terranova recordings list --type <Type>"}}}, nil
		},
	})

	cli.Register(&cli.Command{
		Name: "journal", Group: "Search & Browse",
		Summary: "Le journal d'un recording : qui a fait quoi, quand.",
		ArgSpec: "<recording_id>", MinArgs: 1,
		APIOps: []string{"GET /recordings/{recording_id}/events"},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			var out struct {
				Events []map[string]any `json:"events"`
			}
			if err := client.Get("/recordings/"+args[0]+"/events", &out); err != nil {
				return nil, err
			}
			rows := [][]string{}
			for _, e := range out.Events {
				rows = append(rows, []string{str(e["created_at"]), str(e["actor"]), str(e["action"])})
			}
			return &cli.Result{Data: out.Events, Headers: []string{"QUAND", "QUI", "GESTE"}, Rows: rows,
				Summary: fmt.Sprintf("%d événement(s).", len(out.Events))}, nil
		},
	})
}
