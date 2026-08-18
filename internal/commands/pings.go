package commands

import (
	"fmt"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

// ISC-408 — les pings : mes conversations, poster vers des personnes précises
// (idempotent), répondre, archiver (irréversible, comme à l'écran).
func init() {
	cli.Register(&cli.Command{
		Name: "pings", Group: msg.GroupCommunication,
		Summary: msg.HelpPings,
		Sub: []*cli.Command{
			{
				Name: "list", Summary: msg.HelpPingsList,
				APIOps: []string{"GET /pings"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out struct {
						Pings []map[string]any `json:"pings"`
					}
					if err := client.Get("/pings", &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, p := range out.Pings {
						names := []string{}
						if list, ok := p["participants"].([]any); ok {
							for _, item := range list {
								if m, ok := item.(map[string]any); ok {
									names = append(names, str(m["name"]))
								}
							}
						}
						rows = append(rows, []string{str(p["id"]), strings.Join(names, ", ")})
					}
					return &cli.Result{Data: out.Pings, Headers: msg.HeadersPings, Rows: rows,
						Summary: fmt.Sprintf(msg.ResPingCount, len(out.Pings)),
						Crumbs:  []cli.Crumb{{Action: msg.CrumbLire, Cmd: "terranova pings show <id>"}}}, nil
				},
			},
			{
				Name: "show", Summary: msg.HelpPingsShow, ArgSpec: "<id>", MinArgs: 1,
				APIOps: []string{"GET /pings/{id}"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Get("/pings/"+args[0], &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out,
						Crumbs: []cli.Crumb{{Action: msg.CrumbRepondre, Cmd: "terranova pings reply " + args[0] + " <texte>"}}}, nil
				},
			},
			{
				Name: "send", Summary: msg.HelpPingsSend,
				ArgSpec: "<texte…> --to <user_id,user_id…>", MinArgs: 1,
				Flags:  []cli.Flag{{Name: "to", Arg: "ids", Help: msg.FlagPingsSendTo}},
				APIOps: []string{"POST /pings"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					to, rest := cli.FlagValue(args, "to")
					if to == "" || len(rest) == 0 {
						return nil, cli.Usagef(msg.UsagePingsSend)
					}
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/pings", map[string]any{
						"user_ids": strings.Split(to, ","),
						"body":     strings.Join(rest, " "),
					}, &out); err != nil {
						return nil, err
					}
					id := str(dig(out, "ping", "id"))
					return &cli.Result{Data: out, Summary: msg.ResPingEnvoye,
						Crumbs: []cli.Crumb{{Action: msg.CrumbSuivreLaConversation, Cmd: "terranova pings show " + id}}}, nil
				},
			},
			{
				Name: "reply", Summary: msg.HelpPingsReply, ArgSpec: "<id> <texte…>", MinArgs: 2,
				APIOps: []string{"POST /pings/{id}/messages"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/pings/"+args[0]+"/messages", map[string]any{"body": strings.Join(args[1:], " ")}, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: msg.ResRepondu}, nil
				},
			},
			{
				Name: "archive", Summary: msg.HelpPingsArchive,
				ArgSpec: "<id>", MinArgs: 1,
				APIOps: []string{"POST /pings/{id}/archive"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					if err := confirm(c, msg.AskArchivePing); err != nil {
						return nil, err
					}
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/pings/"+args[0]+"/archive", nil, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: msg.ResArchivee}, nil
				},
			},
		},
	})
}
