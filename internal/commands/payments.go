package commands

import (
	"fmt"
	"net/url"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

// ISC-437 — les paiements : enregistrer un encaissement depuis un script,
// marquer payé par le geste du modèle, et répondre « qui n'a pas soldé ».
func init() {
	cli.Register(&cli.Command{
		Name: "payments", Group: "Lentilles",
		Summary: "Paiements Academy : lister, enregistrer, marquer payé, « qui n'a pas soldé ».",
		Sub: []*cli.Command{
			{
				Name: "list", Summary: "Liste (— --status pending|paid|refunded, --participant <id>).",
				Flags: []cli.Flag{
					{Name: "status", Arg: "statut", Help: "Filtre par état."},
					{Name: "participant", Arg: "id", Help: "Filtre par participant."},
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
					return &cli.Result{Data: out.Payments, Headers: []string{"ID", "ÉTAT", "GENRE", "CENTIMES", "PROJET"}, Rows: rows,
						Summary: fmt.Sprintf("%d paiement(s).", len(out.Payments)),
						Crumbs:  []cli.Crumb{{Action: "qui n'a pas soldé", Cmd: "terranova payments outstanding"}}}, nil
				},
			},
			{
				Name: "show", Summary: "Détail.", ArgSpec: "<id>", MinArgs: 1,
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
				Name: "add", Summary: "Enregistre un paiement (pending) pour un participant.",
				Flags: []cli.Flag{
					{Name: "participant", Arg: "id", Help: "Le participant (obligatoire)."},
					{Name: "amount-cents", Arg: "n", Help: "Montant en centimes (obligatoire)."},
					{Name: "kind", Arg: "genre", Help: "deposit|balance|full (défaut full)."},
					{Name: "reference", Arg: "réf", Help: "Référence libre (virement…)."},
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
						return nil, cli.Usagef("--participant et --amount-cents sont obligatoires")
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
					return &cli.Result{Data: out, Summary: "Paiement enregistré (pending).",
						Crumbs: []cli.Crumb{{Action: "encaissé ? le marquer payé", Cmd: "terranova payments mark-paid " + id}}}, nil
				},
			},
			{
				Name: "mark-paid", Summary: "Marque payé — par Payment#mark_paid!, le geste du webhook.", ArgSpec: "<id>", MinArgs: 1,
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
					return &cli.Result{Data: out, Summary: "Payé."}, nil
				},
			},
			{
				Name: "plans", Summary: "Les échéanciers : plan, échéances, découvert.",
				Flags:  []cli.Flag{{Name: "participant", Arg: "id", Help: "Le participant."}},
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
				Name: "outstanding", Summary: "Qui n'a pas soldé : attendu vs payé par participant.",
				Flags:  []cli.Flag{{Name: "list", Arg: "recording_id", Help: "Borne à une liste de participants (id de recording)."}},
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
					return &cli.Result{Data: out, Headers: []string{"PARTICIPANT", "PROJET", "ATTENDU", "PAYÉ", "RESTE"}, Rows: rows,
						Summary: fmt.Sprintf("%d participant(s) en attente · %d centimes de découvert.", len(out.Outstanding), out.Total)}, nil
				},
			},
		},
	})
}
