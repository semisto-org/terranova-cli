package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

// Vague du 2026-08-18 après-midi — les 9 chemins nés côté app le même jour :
// rapports (ISC-39), remontées + silence (solde ISC-408), fusion de contacts
// (ISC-441/#549). Fichier zz_* : les greffes sur `notifications`, `my` et
// `contacts` doivent passer APRÈS leur enregistrement (ordre d'init de Go).
func init() {
	registerReports()
	graftNotificationsSilence()
	graftMyBubbleUps()
	graftContactsMerge()
}

func reportRows(items []map[string]any) [][]string {
	rows := [][]string{}
	for _, i := range items {
		assignees := []string{}
		if list, ok := i["assignees"].([]any); ok {
			for _, a := range list {
				if m, ok := a.(map[string]any); ok {
					assignees = append(assignees, str(m["name"]))
				}
			}
		}
		rows = append(rows, []string{str(i["id"]), str(i["kind"]), str(i["title"]), str(i["due_on"]),
			str(dig(i, "project", "name")), strings.Join(assignees, ", ")})
	}
	return rows
}

func registerReports() {
	cli.Register(&cli.Command{
		Name: "reports", Group: msg.GroupSchedulingTime,
		Summary: msg.HelpReports,
		Sub: []*cli.Command{
			{
				Name: "upcoming", Summary: msg.HelpReportsUpcoming,
				Flags:  []cli.Flag{{Name: "within-days", Arg: "<n>", Help: msg.FlagWithinDays}},
				APIOps: []string{"GET /reports/upcoming"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					within, _ := cli.FlagValue(args, "within-days")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/reports/upcoming"
					if within != "" {
						path += "?within_days=" + url.QueryEscape(within)
					}
					var out struct {
						HorizonDays any              `json:"horizon_days"`
						Items       []map[string]any `json:"items"`
					}
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Headers: msg.HeadersReportItems, Rows: reportRows(out.Items),
						Summary: fmt.Sprintf(msg.ResUpcomingCount, len(out.Items), str(out.HorizonDays))}, nil
				},
			},
			{
				Name: "overdue", Summary: msg.HelpReportsOverdue,
				APIOps: []string{"GET /reports/overdue"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out struct {
						Buckets map[string][]map[string]any `json:"buckets"`
						Total   int                         `json:"total"`
					}
					if err := client.Get("/reports/overdue", &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					// L'ordre des groupes est celui de l'écran.
					for _, k := range []string{"today", "this_week", "last_week", "older"} {
						for _, r := range reportRows(out.Buckets[k]) {
							rows = append(rows, append([]string{k}, r...))
						}
					}
					headers := append([]string{"RETARD"}, msg.HeadersReportItems...)
					return &cli.Result{Data: out, Headers: headers, Rows: rows,
						Summary: fmt.Sprintf(msg.ResOverdueCount, out.Total, len(out.Buckets["today"]),
							len(out.Buckets["this_week"]), len(out.Buckets["last_week"]), len(out.Buckets["older"]))}, nil
				},
			},
			{
				Name: "unassigned", Summary: msg.HelpReportsUnassigned,
				APIOps: []string{"GET /reports/unassigned"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out struct {
						Items []map[string]any `json:"items"`
					}
					if err := client.Get("/reports/unassigned", &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out.Items, Headers: msg.HeadersReportItems, Rows: reportRows(out.Items),
						Summary: fmt.Sprintf(msg.ResUnassignedCount, len(out.Items)),
						Crumbs:  []cli.Crumb{{Action: msg.CrumbAssigner, Cmd: "terranova todos assign <id> <user_id>"}}}, nil
				},
			},
			{
				Name: "assignments", Summary: msg.HelpReportsAssignments,
				Flags:  []cli.Flag{{Name: "person", Arg: "<user_id>", Help: msg.FlagPerson}},
				APIOps: []string{"GET /reports/assignments"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					person, _ := cli.FlagValue(args, "person")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/reports/assignments"
					if person != "" {
						path += "?person_id=" + url.QueryEscape(person)
					}
					var out struct {
						Person map[string]any   `json:"person"`
						Items  []map[string]any `json:"items"`
					}
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Headers: msg.HeadersReportItems, Rows: reportRows(out.Items),
						Summary: fmt.Sprintf(msg.ResAssignmentsCount, len(out.Items), str(out.Person["name"]))}, nil
				},
			},
			{
				Name: "throughput", Summary: msg.HelpReportsThroughput,
				Flags: []cli.Flag{
					{Name: "days", Arg: "<n>", Help: msg.FlagReportDays},
					{Name: "kind", Arg: "<genre>", Help: msg.FlagReportKind},
				},
				APIOps: []string{"GET /reports/throughput"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					days, rest := cli.FlagValue(args, "days")
					kind, _ := cli.FlagValue(rest, "kind")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					q := url.Values{}
					if days != "" {
						q.Set("days", days)
					}
					if kind != "" {
						q.Set("kind", kind)
					}
					path := "/reports/throughput"
					if len(q) > 0 {
						path += "?" + q.Encode()
					}
					var out struct {
						Days   any              `json:"days"`
						Kind   string           `json:"kind"`
						Series []map[string]any `json:"series"`
					}
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					added, completed := 0, 0
					rows := [][]string{}
					for _, p := range out.Series {
						a, c2 := intOf(p["added"]), intOf(p["completed"])
						added += a
						completed += c2
						rows = append(rows, []string{str(p["date"]), str(p["added"]), str(p["completed"])})
					}
					return &cli.Result{Data: out, Headers: []string{"DATE", "CRÉÉS", "TERMINÉS"}, Rows: rows,
						Summary: fmt.Sprintf(msg.ResThroughputSummary, str(out.Days), added, completed)}, nil
				},
			},
		},
	})
}

func intOf(v any) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

func graftNotificationsSilence() {
	for _, cmd := range cli.Registry {
		if cmd.Name != "notifications" {
			continue
		}
		cmd.Sub = append(cmd.Sub,
			&cli.Command{
				Name: "silence", Summary: msg.HelpNotificationsSilence,
				APIOps: []string{"POST /notifications/silence"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/notifications/silence", nil, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: msg.ResSilenceOn,
						Crumbs: []cli.Crumb{{Action: msg.CrumbRallumer, Cmd: "terranova notifications unsilence"}}}, nil
				},
			},
			&cli.Command{
				Name: "unsilence", Summary: msg.HelpNotificationsUnsilence,
				APIOps: []string{"DELETE /notifications/silence"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Delete("/notifications/silence", &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: msg.ResSilenceOff}, nil
				},
			},
		)
	}
}

func graftMyBubbleUps() {
	for _, cmd := range cli.Registry {
		if cmd.Name != "my" {
			continue
		}
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "bubble-ups", Summary: msg.HelpMyBubbleUps,
			APIOps: []string{"GET /my/bubble_ups"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out struct {
					BubbleUps []map[string]any `json:"bubble_ups"`
				}
				if err := client.Get("/my/bubble_ups", &out); err != nil {
					return nil, err
				}
				rows := [][]string{}
				for _, b := range out.BubbleUps {
					rows = append(rows, []string{str(b["id"]), str(b["recording_id"]), str(b["title"]),
						str(b["project"]), str(b["surface_at"])})
				}
				return &cli.Result{Data: out.BubbleUps, Headers: msg.HeadersBubbleUps, Rows: rows,
					Summary: fmt.Sprintf(msg.ResBubbleUpCount, len(out.BubbleUps))}, nil
			},
		})
	}
}

func graftContactsMerge() {
	for _, cmd := range cli.Registry {
		if cmd.Name != "contacts" {
			continue
		}
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "merge", Summary: msg.HelpContactsMerge,
			ArgSpec: "<recording_id_survivant> <recording_id_absorbé>", MinArgs: 2,
			APIOps: []string{"POST /recordings/{id}/merge"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				// Irréversible (ISC-381) : confirmation, ou --yes ; jamais
				// silencieux en mode --agent.
				if err := confirm(c, fmt.Sprintf(msg.AskMergeContacts, args[0], args[1])); err != nil {
					return nil, err
				}
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Post("/recordings/"+args[0]+"/merge",
					map[string]any{"merged_recording_id": args[1]}, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: fmt.Sprintf(msg.ResContactsMerged, args[0])}, nil
			},
		})
	}
}
