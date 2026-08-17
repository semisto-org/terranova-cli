package commands

import (
	"fmt"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

// La greffe des LIGNES sur chat (ISC-401) vit dans un fichier nommé zz_* À
// DESSEIN : Go initialise les fichiers d un paquet en ordre alphabétique, et la
// greffe doit passer APRÈS l enregistrement de chat par spine.go — comms.go
// s initialisait avant, et greffait dans le vide (vécu le 2026-08-18).
// `chat` existe déjà (typeSpec Campfire) — on lui greffe les LIGNES (ISC-401).
func init() {
	for _, cmd := range cli.Registry {
		if cmd.Name != "chat" {
			continue
		}
		cmd.Sub = append(cmd.Sub,
			&cli.Command{
				Name: "lines", Summary: "Les lignes du feu de camp (plus récentes d'abord).",
				ArgSpec: "<campfire_recording_id>", MinArgs: 1,
				APIOps: []string{"GET /recordings/{recording_id}/lines"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out struct {
						Lines []map[string]any `json:"lines"`
					}
					if err := client.Get("/recordings/"+args[0]+"/lines", &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, l := range out.Lines {
						rows = append(rows, []string{str(l["id"]), str(l["author"]), str(l["body"])})
					}
					return &cli.Result{Data: out.Lines, Headers: []string{"ID", "QUI", "QUOI"}, Rows: rows,
						Summary: fmt.Sprintf("%d ligne(s).", len(out.Lines)),
						Crumbs:  []cli.Crumb{{Action: "répondre", Cmd: "terranova chat post " + args[0] + " <texte>"}}}, nil
				},
			},
			&cli.Command{
				Name: "post", Summary: "Poste une ligne, signée du porteur du jeton.",
				ArgSpec: "<campfire_recording_id> <texte…>", MinArgs: 2,
				APIOps: []string{"POST /recordings/{recording_id}/lines"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/recordings/"+args[0]+"/lines",
						map[string]any{"body": strings.Join(args[1:], " ")}, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: "Ligne postée."}, nil
				},
			},
			&cli.Command{
				Name: "edit-line", Summary: "Édite SA propre ligne.",
				ArgSpec: "<campfire_recording_id> <line_id> <texte…>", MinArgs: 3,
				APIOps: []string{"PATCH /recordings/{recording_id}/lines/{id}"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Patch("/recordings/"+args[0]+"/lines/"+args[1],
						map[string]any{"body": strings.Join(args[2:], " ")}, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: "Ligne éditée."}, nil
				},
			},
			&cli.Command{
				Name: "delete-line", Summary: "Supprime SA propre ligne (confirmation, ou --yes).",
				ArgSpec: "<campfire_recording_id> <line_id>", MinArgs: 2,
				APIOps: []string{"DELETE /recordings/{recording_id}/lines/{id}"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					if err := confirm(c, "Supprimer cette ligne de chat ?"); err != nil {
						return nil, err
					}
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					if err := client.Delete("/recordings/"+args[0]+"/lines/"+args[1], nil); err != nil {
						return nil, err
					}
					return &cli.Result{Data: map[string]any{"deleted": true}, Summary: "Ligne supprimée."}, nil
				},
			},
		)
	}
}
