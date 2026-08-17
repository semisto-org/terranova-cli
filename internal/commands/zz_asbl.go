package commands

import (
	"fmt"
	"net/url"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

// ISC-412 — `administratio members` : l'adhésion ASBL (cotisations, AG,
// demandes d'effectif). zz_* : greffé après l'enregistrement d'administratio.
func init() {
	members := &cli.Command{
		Name: "members", Summary: "Adhésion ASBL : cotisations, taux, assemblées, demandes d'effectif (admin du hub).",
		Sub: []*cli.Command{
			{
				Name: "fees", Summary: "Cotisations (— --status pending|paid, --enrollment <id>).",
				Flags: []cli.Flag{
					{Name: "status", Arg: "statut", Help: "pending|paid."},
					{Name: "enrollment", Arg: "id", Help: "Une adhésion précise."},
				},
				APIOps: []string{"GET /asbl/fees"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					status, args := cli.FlagValue(args, "status")
					enrollment, _ := cli.FlagValue(args, "enrollment")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/asbl/fees"
					sep := "?"
					if status != "" {
						path += sep + "status=" + url.QueryEscape(status)
						sep = "&"
					}
					if enrollment != "" {
						path += sep + "member_enrollment_id=" + url.QueryEscape(enrollment)
					}
					var out struct {
						Fees []map[string]any `json:"fees"`
					}
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, f := range out.Fees {
						rows = append(rows, []string{str(f["id"]), str(f["member"]), str(f["status"]), str(f["amount_cents"]), str(f["starts_on"])})
					}
					return &cli.Result{Data: out.Fees, Headers: []string{"ID", "MEMBRE", "ÉTAT", "CENTIMES", "DÉBUT"}, Rows: rows,
						Summary: fmt.Sprintf("%d cotisation(s).", len(out.Fees)),
						Crumbs:  []cli.Crumb{{Action: "encaissée ? la marquer payée", Cmd: "terranova administratio members mark-fee-paid <id>"}}}, nil
				},
			},
			{
				Name: "mark-fee-paid", Summary: "Marque une cotisation payée — par le geste du modèle.", ArgSpec: "<id>", MinArgs: 1,
				APIOps: []string{"POST /asbl/fees/{id}/mark_paid"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/asbl/fees/"+args[0]+"/mark_paid", nil, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: "Cotisation payée."}, nil
				},
			},
			simpleGet("rates", "Les taux de cotisation (type × période de validité).", "/asbl/rates"),
			simpleGet("assemblies", "Les assemblées générales du hub.", "/asbl/assemblies"),
			{
				Name: "effectif-requests", Summary: "Demandes de statut effectif (— --status).",
				Flags:  []cli.Flag{{Name: "status", Arg: "statut", Help: "Filtre par état."}},
				APIOps: []string{"GET /asbl/effectif_requests"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					status, _ := cli.FlagValue(args, "status")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/asbl/effectif_requests"
					if status != "" {
						path += "?status=" + url.QueryEscape(status)
					}
					var out map[string]any
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out}, nil
				},
			},
		},
	}
	for _, cmd := range cli.Registry {
		if cmd.Name == "administratio" {
			cmd.Sub = append(cmd.Sub, members)
		}
	}
}

func simpleGet(name, summary, path string) *cli.Command {
	return &cli.Command{
		Name: name, Summary: summary,
		APIOps: []string{"GET " + path},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			var out map[string]any
			if err := client.Get(path, &out); err != nil {
				return nil, err
			}
			return &cli.Result{Data: out}, nil
		},
	}
}
