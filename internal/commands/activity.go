package commands

import (
	"fmt"
	"net/url"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

// ISC-407 (solde) — le fil d'activité, le journal d'un recording, la totale.
func init() {
	cli.Register(&cli.Command{
		Name: "activity", Group: msg.GroupSearchBrowse,
		Summary: msg.HelpActivity,
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
			return &cli.Result{Data: out.Events, Headers: msg.HeadersActivityFeed, Rows: rows,
				Summary: fmt.Sprintf(msg.ResEventCount, len(out.Events))}, nil
		},
	})

	cli.Register(&cli.Command{
		Name: "everything", Group: msg.GroupSearchBrowse,
		Summary: msg.HelpEverything,
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
				Crumbs: []cli.Crumb{{Action: msg.CrumbCreuserUnType, Cmd: "terranova recordings list --type <Type>"}}}, nil
		},
	})

	cli.Register(&cli.Command{
		Name: "journal", Group: msg.GroupSearchBrowse,
		Summary: msg.HelpJournal,
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
			return &cli.Result{Data: out.Events, Headers: msg.HeadersJournal, Rows: rows,
				Summary: fmt.Sprintf(msg.ResEventCount, len(out.Events))}, nil
		},
	})
}
