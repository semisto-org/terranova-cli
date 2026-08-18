package commands

import (
	"fmt"
	"net/url"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
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
		Name: "academio", Group: msg.GroupLentilles,
		Summary:   msg.HelpAcademio,
		AgentHelp: msg.NotesAcademio,
		Sub: []*cli.Command{
			{
				Name: "activities", Summary: msg.HelpAcademioActivities,
				Sub: []*cli.Command{
					{
						Name: "list", Summary: msg.HelpAcademioActivitiesList,
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
							return &cli.Result{Data: out.Activities, Headers: msg.HeadersActivities, Rows: rows,
								Summary: fmt.Sprintf(msg.ResActivityCount, len(out.Activities)),
								Crumbs:  []cli.Crumb{{Action: msg.CrumbLeDetailEtLesSeances, Cmd: "terranova academio activities show <id>"}}}, nil
						},
					},
					{
						Name: "show", Summary: msg.HelpAcademioActivitiesShow, ArgSpec: "<id>", MinArgs: 1,
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
								Crumbs: []cli.Crumb{{Action: msg.CrumbLaFeuilleDePresence, Cmd: "terranova academio attendances list --session <schedule_entry_id>"}}}, nil
						},
					},
				},
			},
			{
				Name: "attendances", Summary: msg.HelpAcademioAttendances,
				Sub: []*cli.Command{
					{
						Name: "list", Summary: msg.HelpAcademioAttendancesList,
						Flags: []cli.Flag{
							{Name: "session", Arg: "id", Help: msg.FlagAcademioAttendancesListSession},
							{Name: "participant", Arg: "id", Help: msg.FlagParticipant},
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
							return &cli.Result{Data: out.Attendances, Headers: msg.HeadersAttendances, Rows: rows,
								Summary: fmt.Sprintf(msg.ResAttendanceCount, len(out.Attendances))}, nil
						},
					},
					{
						Name: "set", Summary: msg.HelpAcademioAttendancesSet,
						ArgSpec: "<participant_id> <schedule_entry_id>", MinArgs: 2,
						Flags:  []cli.Flag{{Name: "absent", Help: msg.FlagAcademioAttendancesSetAbsent}},
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
							return &cli.Result{Data: out, Summary: msg.ResPresencePosee}, nil
						},
					},
					{
						Name: "remove", Summary: msg.HelpAcademioAttendancesRemove, ArgSpec: "<id>", MinArgs: 1,
						APIOps: []string{"DELETE /academio/attendances/{id}"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							if err := client.Delete("/academio/attendances/"+args[0], nil); err != nil {
								return nil, err
							}
							return &cli.Result{Data: map[string]any{"removed": args[0]}, Summary: msg.ResRetiree}, nil
						},
					},
				},
			},
			{
				Name: "packs", Summary: msg.HelpAcademioPacks,
				Sub: []*cli.Command{
					{
						Name: "list", Summary: msg.HelpAcademioPacksList,
						Flags:  []cli.Flag{{Name: "participant", Arg: "id", Help: msg.FlagParticipant}},
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
							return &cli.Result{Data: out.Packs, Summary: fmt.Sprintf(msg.ResPackCount, len(out.Packs))}, nil
						},
					},
					{
						Name: "add", Summary: msg.HelpAcademioPacksAdd,
						ArgSpec: "<participant_id> <training_pack_id>", MinArgs: 2,
						Flags: []cli.Flag{
							{Name: "quantity", Arg: "n", Help: msg.FlagAcademioPacksAddQuantity},
							{Name: "price-cents", Arg: "n", Help: msg.FlagAcademioPacksAddPriceCents},
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
							return &cli.Result{Data: out, Summary: msg.ResPackAjoute}, nil
						},
					},
					{
						Name: "remove", Summary: msg.HelpAcademioPacksRemove, ArgSpec: "<id>", MinArgs: 1,
						APIOps: []string{"DELETE /academio/registration_packs/{id}"},
						Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
							client, err := c.API()
							if err != nil {
								return nil, err
							}
							if err := client.Delete("/academio/registration_packs/"+args[0], nil); err != nil {
								return nil, err
							}
							return &cli.Result{Data: map[string]any{"removed": args[0]}, Summary: msg.ResRetire}, nil
						},
					},
				},
			},
			{
				Name: "types", Summary: msg.HelpAcademioTypes,
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
				Name: "locations", Summary: msg.HelpAcademioLocations,
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
