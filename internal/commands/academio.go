package commands

import (
	"fmt"
	"net/url"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

// ISC-415/436 — Academio au CLI : les activités et leurs séances, la feuille de
// présence, les packs, les référentiels. Les inscriptions elles-mêmes restent
// des recordables (`participants add`) — aucune architecture parallèle.
func init() {
	q := func(params map[string]string) string {
		out := ""
		sep := "?"
		for k, v := range params {
			if v != "" {
				out += sep + k + "=" + url.QueryEscape(v)
				sep = "&"
			}
		}
		return out
	}

	cli.Register(&cli.Command{
		Name: "academio", Group: "Lentilles",
		Summary:   "Academy : activités, séances, présences, packs, référentiels.",
		AgentHelp: "Scope requis : academio (grant + jeton). Les inscriptions passent par `participants add` (spine).",
		Sub: []*cli.Command{
			{
				Name: "activities", Summary: "Activités : type, lieu, inscrits (même définition que le reporting).",
				Sub: []*cli.Command{
					{
						Name: "list", Summary: "Liste des activités.",
						APIOps: []string{"GET /academio/activities"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							var out struct {
								Activities []map[string]any `json:"activities"`
							}
							if err := client.Get("/academio/activities", &out); err != nil {
								return nil, err
							}
							rows := [][]string{}
							for _, a := range out.Activities {
								rows = append(rows, []string{str(a["id"]), str(a["name"]), str(a["training_type"]), str(a["training_location"]), str(a["registered_count"])})
							}
							return &cli.Result{Data: out.Activities, Headers: []string{"ID", "ACTIVITÉ", "TYPE", "LIEU", "INSCRITS"}, Rows: rows,
								Summary: fmt.Sprintf("%d activité(s).", len(out.Activities)),
								Crumbs:  []cli.Crumb{{Action: "le détail et les séances", Cmd: "terranova academio activities show <id>"}}}, nil
						},
					},
					{
						Name: "show", Summary: "Détail avec les séances.", ArgSpec: "<id>", MinArgs: 1,
						APIOps: []string{"GET /academio/activities/{id}"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							var out map[string]any
							if err := client.Get("/academio/activities/"+args[0], &out); err != nil {
								return nil, err
							}
							return &cli.Result{Data: out,
								Crumbs: []cli.Crumb{{Action: "la feuille de présence d'une séance", Cmd: "terranova academio attendances list --session <schedule_entry_id>"}}}, nil
						},
					},
				},
			},
			{
				Name: "attendances", Summary: "La feuille de présence (participant × séance).",
				Sub: []*cli.Command{
					{
						Name: "list", Summary: "Liste (— --session <id> ou --participant <id>).",
						Flags: []cli.Flag{
							{Name: "session", Arg: "id", Help: "La séance (schedule_entry_id)."},
							{Name: "participant", Arg: "id", Help: "Le participant."},
						},
						APIOps: []string{"GET /academio/attendances"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							session, args := cli.FlagValue(args, "session")
							participant, _ := cli.FlagValue(args, "participant")
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							var out struct {
								Attendances []map[string]any `json:"attendances"`
							}
							if err := client.Get("/academio/attendances"+q(map[string]string{"schedule_entry_id": session, "participant_id": participant}), &out); err != nil {
								return nil, err
							}
							rows := [][]string{}
							for _, a := range out.Attendances {
								present := "✗"
								if b, ok := a["present"].(bool); ok && b {
									present = "✓"
								}
								rows = append(rows, []string{str(a["id"]), str(a["participant_id"]), str(a["schedule_entry_id"]), present})
							}
							return &cli.Result{Data: out.Attendances, Headers: []string{"ID", "PARTICIPANT", "SÉANCE", "PRÉSENT"}, Rows: rows,
								Summary: fmt.Sprintf("%d ligne(s) de présence.", len(out.Attendances))}, nil
						},
					},
					{
						Name: "set", Summary: "Pose une présence (— --absent pour marquer absent). Upsert idempotent.",
						ArgSpec: "<participant_id> <schedule_entry_id>", MinArgs: 2,
						Flags:  []cli.Flag{{Name: "absent", Help: "Marque ABSENT (présent par défaut)."}},
						APIOps: []string{"POST /academio/attendances"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							absent, rest := cli.FlagBool(args, "absent")
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							body := map[string]any{"participant_id": rest[0], "schedule_entry_id": rest[1], "present": !absent}
							var out map[string]any
							if err := client.Post("/academio/attendances", body, &out); err != nil {
								return nil, err
							}
							return &cli.Result{Data: out, Summary: "Présence posée."}, nil
						},
					},
					{
						Name: "remove", Summary: "Retire une ligne de présence.", ArgSpec: "<id>", MinArgs: 1,
						APIOps: []string{"DELETE /academio/attendances/{id}"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							if err := client.Delete("/academio/attendances/"+args[0], nil); err != nil {
								return nil, err
							}
							return &cli.Result{Data: map[string]any{"removed": args[0]}, Summary: "Retirée."}, nil
						},
					},
				},
			},
			{
				Name: "packs", Summary: "Les packs d'une inscription.",
				Sub: []*cli.Command{
					{
						Name: "list", Summary: "Liste (— --participant <id>).",
						Flags:  []cli.Flag{{Name: "participant", Arg: "id", Help: "Le participant."}},
						APIOps: []string{"GET /academio/registration_packs"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							participant, _ := cli.FlagValue(args, "participant")
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							var out struct {
								Packs []map[string]any `json:"registration_packs"`
							}
							if err := client.Get("/academio/registration_packs"+q(map[string]string{"participant_id": participant}), &out); err != nil {
								return nil, err
							}
							return &cli.Result{Data: out.Packs, Summary: fmt.Sprintf("%d pack(s).", len(out.Packs))}, nil
						},
					},
					{
						Name: "add", Summary: "Ajoute un pack (prix par défaut du pack).",
						ArgSpec: "<participant_id> <training_pack_id>", MinArgs: 2,
						Flags: []cli.Flag{
							{Name: "quantity", Arg: "n", Help: "Quantité (défaut 1)."},
							{Name: "price-cents", Arg: "n", Help: "Prix forcé, sinon celui du pack."},
						},
						APIOps: []string{"POST /academio/registration_packs"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							quantity, args := cli.FlagValue(args, "quantity")
							price, rest := cli.FlagValue(args, "price-cents")
							body := map[string]any{"participant_id": rest[0], "training_pack_id": rest[1]}
							if quantity != "" {
								body["quantity"] = quantity
							}
							if price != "" {
								body["price_cents"] = price
							}
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							var out map[string]any
							if err := client.Post("/academio/registration_packs", body, &out); err != nil {
								return nil, err
							}
							return &cli.Result{Data: out, Summary: "Pack ajouté."}, nil
						},
					},
					{
						Name: "remove", Summary: "Retire un pack d'une inscription.", ArgSpec: "<id>", MinArgs: 1,
						APIOps: []string{"DELETE /academio/registration_packs/{id}"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							if err := client.Delete("/academio/registration_packs/"+args[0], nil); err != nil {
								return nil, err
							}
							return &cli.Result{Data: map[string]any{"removed": args[0]}, Summary: "Retiré."}, nil
						},
					},
				},
			},
			{
				Name: "types", Summary: "Les types de formation du hub.",
				APIOps: []string{"GET /academio/types"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Get("/academio/types", &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out}, nil
				},
			},
			{
				Name: "locations", Summary: "Les lieux de formation du hub.",
				APIOps: []string{"GET /academio/locations"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Get("/academio/locations", &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out}, nil
				},
			},
		},
	})
}
