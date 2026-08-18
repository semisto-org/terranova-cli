package commands

import (
	"fmt"
	"net/url"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

// ISC-437 — les paiements : enregistrer un encaissement depuis un script,
// marquer payé par le geste du modèle, et répondre « qui n'a pas soldé ».
func init() {
	cli.Register(&cli.Command{
		Name: "payments", Group: msg.GroupLentilles,
		Summary: msg.HelpPayments,
		Sub: []*cli.Command{
			{
				Name: "list", Summary: msg.HelpPaymentsList,
				Flags: []cli.Flag{
					{Name: "status", Arg: "statut", Help: msg.FlagFiltreParEtat},
					{Name: "participant", Arg: "id", Help: msg.FlagPaymentsListParticipant},
				},
				APIOps: []string{"GET /payments"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					status, args := cli.FlagValue(args, "status")
					participant, _ := cli.FlagValue(args, "participant")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/payments"
					sep := "?"
					if status != "" {
						path += sep + "status=" + url.QueryEscape(status)
						sep = "&"
					}
					if participant != "" {
						path += sep + "participant_id=" + url.QueryEscape(participant)
					}
					var out struct {
						Payments []map[string]any `json:"payments"`
					}
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, p := range out.Payments {
						rows = append(rows, []string{str(p["id"]), str(p["status"]), str(p["kind"]), str(p["amount_cents"]), str(p["project"])})
					}
					return &cli.Result{Data: out.Payments, Headers: msg.HeadersPayments, Rows: rows,
						Summary: fmt.Sprintf(msg.ResPaymentCount, len(out.Payments)),
						Crumbs:  []cli.Crumb{{Action: msg.CrumbQuiNAPasSolde, Cmd: "terranova payments outstanding"}}}, nil
				},
			},
			{
				Name: "show", Summary: msg.HelpRestShow, ArgSpec: "<id>", MinArgs: 1,
				APIOps: []string{"GET /payments/{id}"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Get("/payments/"+args[0], &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out}, nil
				},
			},
			{
				Name: "add", Summary: msg.HelpPaymentsAdd,
				Flags: []cli.Flag{
					{Name: "participant", Arg: "id", Help: msg.FlagPaymentsAddParticipant},
					{Name: "amount-cents", Arg: "n", Help: msg.FlagPaymentsAddAmountCents},
					{Name: "kind", Arg: "genre", Help: msg.FlagPaymentsAddKind},
					{Name: "reference", Arg: "réf", Help: msg.FlagPaymentsAddReference},
				},
				APIOps: []string{"POST /payments"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					body := map[string]any{}
					for flag, param := range map[string]string{"participant": "participant_id",
						"amount-cents": "amount_cents", "kind": "kind", "reference": "reference"} {
						v, r := cli.FlagValue(args, flag)
						args = r
						if v != "" {
							body[param] = v
						}
					}
					if body["participant_id"] == nil || body["amount_cents"] == nil {
						return nil, cli.Usagef(msg.UsagePaymentsAdd)
					}
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/payments", body, &out); err != nil {
						return nil, err
					}
					id := str(dig(out, "payment", "id"))
					return &cli.Result{Data: out, Summary: msg.ResPaiementEnregistrePending,
						Crumbs: []cli.Crumb{{Action: msg.CrumbEncaisseLeMarquerPaye, Cmd: "terranova payments mark-paid " + id}}}, nil
				},
			},
			{
				Name: "mark-paid", Summary: msg.HelpPaymentsMarkPaid, ArgSpec: "<id>", MinArgs: 1,
				APIOps: []string{"POST /payments/{id}/mark_paid"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/payments/"+args[0]+"/mark_paid", nil, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: msg.ResPaye}, nil
				},
			},
			{
				Name: "plans", Summary: msg.HelpPaymentsPlans,
				Flags:  []cli.Flag{{Name: "participant", Arg: "id", Help: msg.FlagParticipant}},
				APIOps: []string{"GET /payments/plans"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					participant, _ := cli.FlagValue(args, "participant")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/payments/plans"
					if participant != "" {
						path += "?participant_id=" + url.QueryEscape(participant)
					}
					var out map[string]any
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out}, nil
				},
			},
			{
				Name: "outstanding", Summary: msg.HelpPaymentsOutstanding,
				Flags:  []cli.Flag{{Name: "list", Arg: "recording_id", Help: msg.FlagPaymentsOutstandingList}},
				APIOps: []string{"GET /payments/outstanding"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					listID, _ := cli.FlagValue(args, "list")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/payments/outstanding"
					if listID != "" {
						path += "?participant_list_id=" + url.QueryEscape(listID)
					}
					var out struct {
						Outstanding []map[string]any `json:"outstanding"`
						Total       int64            `json:"total_outstanding_cents"`
					}
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, r := range out.Outstanding {
						rows = append(rows, []string{str(r["participant_id"]), str(r["project"]), str(r["expected_cents"]), str(r["paid_cents"]), str(r["outstanding_cents"])})
					}
					return &cli.Result{Data: out, Headers: msg.HeadersOutstanding, Rows: rows,
						Summary: fmt.Sprintf(msg.ResOutstandingCount, len(out.Outstanding), out.Total)}, nil
				},
			},
		},
	})
}
