package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

// typeSpec décrit un recordable exposé par l'endpoint polymorphe /recordings.
// UNE table, un générateur : chaque type gagne list/show/add/edit/trash + les
// gestes de la spine (ISC-405) sans une ligne d'architecture propre (ISC-423).
type typeSpec struct {
	Cmd     string // nom de commande (anglais, pluriel)
	Type    string // recordable_type côté API
	Group   string
	Summary string
	// Create : drapeaux propres à la création (kebab-case → snake_case en param).
	Create []cli.Flag
	// Parent : indice d'aide — dans quoi ça se crée.
	Parent string
	// NoCreate : types qui naissent autrement (outils du dock, singletons…).
	NoCreate bool
}

var typeSpecs = []typeSpec{
	// ── Projecto — les 9 outils Basecamp (ISC-397→404) ──
	{Cmd: "todosets", Type: "Todoset", Group: "Core", Summary: "Les racines To-dos des projets.", NoCreate: true},
	{Cmd: "todolists", Type: "Todolist", Group: "Core", Summary: "Listes de tâches.", Parent: "todoset",
		Create: []cli.Flag{{Name: "description", Arg: "texte", Help: "Description de la liste."}}},
	{Cmd: "todogroups", Type: "TodoGroup", Group: "Core", Summary: "Groupes à l'intérieur d'une liste.", Parent: "todolist"},
	{Cmd: "todos", Type: "Todo", Group: "Core", Summary: "Tâches : création, échéances, récurrence, complétion.", Parent: "todolist",
		Create: []cli.Flag{
			{Name: "due-on", Arg: "date", Help: "Échéance (AAAA-MM-JJ)."},
			{Name: "starts-on", Arg: "date", Help: "Date de début."},
			{Name: "recurrence", Arg: "règle", Help: "Récurrence (daily|weekly|monthly|yearly)."},
			{Name: "recurrence-until", Arg: "date", Help: "Fin de la série."},
		}},
	{Cmd: "messageboards", Type: "MessageBoard", Group: "Core", Summary: "Les tableaux de messages des projets.", NoCreate: true},
	{Cmd: "messages", Type: "Message", Group: "Core", Summary: "Messages : publier, catégoriser.", Parent: "messageboard",
		Create: []cli.Flag{
			{Name: "body", Arg: "html", Help: "Corps (HTML ActionText)."},
			{Name: "category", Arg: "clé", Help: "Catégorie du message."},
		}},
	{Cmd: "cardtables", Type: "CardTable", Group: "Core", Summary: "Les tables de cartes des projets.", NoCreate: true},
	{Cmd: "cardcolumns", Type: "CardColumn", Group: "Core", Summary: "Colonnes d'une table de cartes.", Parent: "cardtable",
		Create: []cli.Flag{{Name: "kind", Arg: "genre", Help: "normal|triage|done (défaut normal)."}}},
	{Cmd: "cards", Type: "Card", Group: "Core", Summary: "Cartes : création, colonne, échéance — la complétion EST la colonne done.", Parent: "cardcolumn",
		Create: []cli.Flag{
			{Name: "due-on", Arg: "date", Help: "Échéance (AAAA-MM-JJ)."},
			{Name: "body", Arg: "html", Help: "Description de la carte."},
		}},
	{Cmd: "folders", Type: "Vault", Group: "Files & Docs", Summary: "Coffres et dossiers de Docs & Fichiers.", Parent: "vault parent"},
	{Cmd: "docs", Type: "Document", Group: "Files & Docs", Summary: "Documents riches.", Parent: "vault",
		Create: []cli.Flag{{Name: "body", Arg: "html", Help: "Corps (HTML ActionText)."}}},
	{Cmd: "uploads", Type: "Upload", Group: "Files & Docs", Summary: "Fichiers : dépôt réel (multipart) et métadonnées.", Parent: "vault",
		Create: []cli.Flag{
			{Name: "file", Arg: "chemin", Help: "Le fichier à téléverser (multipart)."},
			{Name: "caption", Arg: "texte", Help: "Légende."},
		}},
	{Cmd: "links", Type: "CloudLink", Group: "Files & Docs", Summary: "Liens externes (Drive, Notion…).", Parent: "vault",
		Create: []cli.Flag{
			{Name: "url", Arg: "url", Help: "L'URL du lien (obligatoire)."},
			{Name: "service", Arg: "nom", Help: "Le service (drive, notion…)."},
			{Name: "notes", Arg: "texte", Help: "Notes."},
		}},
	{Cmd: "chat", Type: "Campfire", Group: "Core", Summary: "Les feux de camp (chat) des projets.", NoCreate: true},
	{Cmd: "checkins", Type: "Questionnaire", Group: "Core", Summary: "Les questionnaires automatiques des projets.", NoCreate: true},
	{Cmd: "questions", Type: "Question", Group: "Core", Summary: "Questions de check-in (fréquence, pause).", Parent: "questionnaire",
		Create: []cli.Flag{{Name: "frequency", Arg: "règle", Help: "daily|weekly|monthly (défaut weekly)."}}},
	{Cmd: "answers", Type: "Answer", Group: "Core", Summary: "Réponses aux questions de check-in.", Parent: "question",
		Create: []cli.Flag{{Name: "body", Arg: "html", Help: "La réponse (HTML ActionText)."}}},
	{Cmd: "schedules", Type: "Schedule", Group: "Core", Summary: "Les agendas des projets.", NoCreate: true},
	{Cmd: "events", Type: "ScheduleEntry", Group: "Core", Summary: "Événements : datés, journée entière, lieu.", Parent: "schedule",
		Create: []cli.Flag{
			{Name: "starts-at", Arg: "horodate", Help: "Début (ISO 8601)."},
			{Name: "ends-at", Arg: "horodate", Help: "Fin."},
			{Name: "all-day", Help: "Journée entière."},
			{Name: "location", Arg: "lieu", Help: "Lieu."},
		}},
	{Cmd: "forwards", Type: "InboxForward", Group: "Communication", Summary: "Emails entrants du projet.", Parent: "inbox",
		Create: []cli.Flag{
			{Name: "from-email", Arg: "email", Help: "Expéditeur."},
			{Name: "subject", Arg: "objet", Help: "Objet."},
		}},
	{Cmd: "hillcharts", Type: "HillChartSnapshot", Group: "Core", Summary: "Instantanés de Hill Chart.", NoCreate: true},
	{Cmd: "comments", Type: "Comment", Group: "Communication", Summary: "Commentaires de la spine (sur tout recordable).", Parent: "recording",
		Create: []cli.Flag{{Name: "body", Arg: "html", Help: "Corps du commentaire."}}},
	{Cmd: "timesheets", Type: "Timesheet", Group: "Scheduling & Time", Summary: "Heures pointées (facturable/bénévole).", Parent: "projet",
		Create: []cli.Flag{
			{Name: "worked-on", Arg: "date", Help: "Jour travaillé."},
			{Name: "hours", Arg: "n", Help: "Heures."},
			{Name: "billable", Help: "Facturable."},
			{Name: "description", Arg: "texte", Help: "Description."},
			{Name: "phase", Arg: "phase", Help: "Phase du projet."},
		}},

	// ── Contacto (ISC-413) ──
	{Cmd: "contacts", Type: "Contact", Group: "Lentilles", Summary: "Contacts du CRM (vivent au QG du hub).",
		Create: []cli.Flag{{Name: "kind", Arg: "genre", Help: "person|organization (défaut person)."}}},

	// ── Academio (ISC-415, la part recordable) ──
	{Cmd: "participant-lists", Type: "ParticipantList", Group: "Lentilles", Summary: "Listes de participants d'une activité.", NoCreate: true},
	{Cmd: "participant-categories", Type: "ParticipantCategory", Group: "Lentilles", Summary: "Tarifs/quotas d'une liste de participants.", Parent: "participant-list",
		Create: []cli.Flag{
			{Name: "label", Arg: "libellé", Help: "Libellé du tarif."},
			{Name: "price-cents", Arg: "n", Help: "Prix en centimes."},
			{Name: "quota", Arg: "n", Help: "Quota de places."},
		}},
	{Cmd: "participants", Type: "Participant", Group: "Lentilles", Summary: "Inscriptions d'une activité.", Parent: "participant-list",
		Create: []cli.Flag{
			{Name: "contact-id", Arg: "id", Help: "Le contact inscrit (obligatoire)."},
			{Name: "participant-category-id", Arg: "id", Help: "Le tarif choisi."},
			{Name: "status", Arg: "statut", Help: "registered|waitlist|cancelled (défaut registered)."},
		}},
	{Cmd: "training-documents", Type: "TrainingDocument", Group: "Lentilles", Summary: "Documents de formation d'une séance.", Parent: "projet",
		Create: []cli.Flag{
			{Name: "name", Arg: "nom", Help: "Nom du document."},
			{Name: "schedule-entry-id", Arg: "id", Help: "La séance liée."},
		}},
	{Cmd: "session-feedbacks", Type: "SessionFeedback", Group: "Lentilles", Summary: "Retours de séance.", Parent: "projet",
		Create: []cli.Flag{
			{Name: "schedule-entry-id", Arg: "id", Help: "La séance."},
			{Name: "contact-id", Arg: "id", Help: "Le participant."},
			{Name: "rating", Arg: "n", Help: "Note."},
			{Name: "comment", Arg: "texte", Help: "Commentaire."},
		}},

	// ── Conceptio (ISC-416, la part recordable) ──
	{Cmd: "palettes", Type: "Palette", Group: "Lentilles", Summary: "Palettes végétales des designs.", NoCreate: true},
	{Cmd: "palette-items", Type: "PaletteItem", Group: "Lentilles", Summary: "Plantes d'une palette (strate, quantité, prix, conduite).", Parent: "palette",
		Create: []cli.Flag{
			{Name: "name", Arg: "latin", Help: "Nom latin."},
			{Name: "common-name", Arg: "nom", Help: "Nom commun."},
			{Name: "strata", Arg: "strate", Help: "Strate (canopy, shrub…)."},
			{Name: "quantity", Arg: "n", Help: "Quantité."},
			{Name: "unit-price-cents", Arg: "n", Help: "Prix unitaire en centimes."},
			{Name: "external-variety-id", Arg: "id", Help: "Variété du catalogue Planto."},
		}},
	{Cmd: "concepts", Type: "Concept", Group: "Lentilles", Summary: "Scènes de design (fond de plan, calibration).", Parent: "projet design",
		Create: []cli.Flag{
			{Name: "width-m", Arg: "m", Help: "Largeur en mètres."},
			{Name: "height-m", Arg: "m", Help: "Hauteur en mètres."},
		}},
	{Cmd: "markers", Type: "PlantMarker", Group: "Lentilles", Summary: "Marqueurs de plantation d'un plan.", Parent: "concept",
		Create: []cli.Flag{
			{Name: "x", Arg: "0..1", Help: "Position X normalisée."},
			{Name: "y", Arg: "0..1", Help: "Position Y normalisée."},
			{Name: "species-name", Arg: "nom", Help: "Espèce."},
			{Name: "palette-item-id", Arg: "id", Help: "La plante de la palette."},
		}},
	{Cmd: "quotes", Type: "Quote", Group: "Lentilles", Summary: "Devis d'un design — les totaux se calculent des lignes, jamais posés.", Parent: "projet design",
		Create: []cli.Flag{
			{Name: "vat-rate", Arg: "taux", Help: "Taux de TVA (défaut 21)."},
			{Name: "valid-until", Arg: "date", Help: "Validité."},
		}},
	{Cmd: "quote-lines", Type: "QuoteLine", Group: "Lentilles", Summary: "Lignes de devis.", Parent: "quote",
		Create: []cli.Flag{
			{Name: "description", Arg: "texte", Help: "Description."},
			{Name: "quantity", Arg: "n", Help: "Quantité."},
			{Name: "unit", Arg: "unité", Help: "Unité."},
			{Name: "unit-price-cents", Arg: "n", Help: "Prix unitaire en centimes."},
		}},
	{Cmd: "plant-records", Type: "PlantRecord", Group: "Lentilles", Summary: "Suivi de vie des plantations.", Parent: "projet design",
		Create: []cli.Flag{
			{Name: "status", Arg: "statut", Help: "alive|dead|replaced (défaut alive)."},
			{Name: "plant-marker-id", Arg: "id", Help: "Le marqueur suivi."},
		}},
	{Cmd: "interventions", Type: "Intervention", Group: "Lentilles", Summary: "Interventions sur une plantation.", Parent: "projet design",
		Create: []cli.Flag{
			{Name: "happened-on", Arg: "date", Help: "Date."},
			{Name: "kind", Arg: "genre", Help: "Genre d'intervention."},
			{Name: "notes", Arg: "texte", Help: "Notes."},
		}},
}

func init() {
	for _, spec := range typeSpecs {
		cli.Register(buildTypeCommand(spec))
	}
	cli.Register(buildRecordingsCommand())
}

// buildTypeCommand fabrique la commande complète d'un type recordable.
func buildTypeCommand(spec typeSpec) *cli.Command {
	cmd := &cli.Command{
		Name: spec.Cmd, Group: spec.Group, Summary: spec.Summary,
		APIOps: []string{"GET /recordings", "GET /recordings/{id}"},
	}
	cmd.Sub = append(cmd.Sub,
		&cli.Command{
			Name: "list", Summary: fmt.Sprintf("Liste (type %s) — filtres --project, --parent.", spec.Type),
			Flags:  []cli.Flag{{Name: "parent", Arg: "id", Help: "Restreint aux enfants de ce recording."}},
			APIOps: []string{"GET /recordings"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return listRecordings(c, spec, args) },
		},
		&cli.Command{
			Name: "show", Summary: "Détail d'un recording.", ArgSpec: "<id>", MinArgs: 1,
			APIOps: []string{"GET /recordings/{id}"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return showRecording(c, spec, args[0]) },
		},
	)
	if !spec.NoCreate {
		parentHelp := "Recording parent (contenant)."
		if spec.Parent != "" {
			parentHelp = "Recording parent — ici : " + spec.Parent + "."
		}
		addFlags := append([]cli.Flag{
			{Name: "project", Short: "p", Arg: "id", Help: "Le projet (obligatoire sauf contacts)."},
			{Name: "parent", Arg: "id", Help: parentHelp},
		}, spec.Create...)
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "add", Summary: "Crée un " + spec.Type + ".", ArgSpec: "<titre>", MinArgs: 0,
			Flags:  addFlags,
			APIOps: []string{"POST /recordings"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return createRecording(c, spec, args) },
		})
	}
	cmd.Sub = append(cmd.Sub,
		&cli.Command{
			Name: "edit", Summary: "Modifie titre, corps, échéance, complétion.", ArgSpec: "<id>", MinArgs: 1,
			Flags: []cli.Flag{
				{Name: "title", Arg: "titre", Help: "Nouveau titre."},
				{Name: "body", Arg: "html", Help: "Nouveau corps."},
				{Name: "due-on", Arg: "date", Help: "Échéance (tâches)."},
				{Name: "completed", Arg: "bool", Help: "true|false — passe par Todo#complete!, jamais un booléen nu."},
			},
			APIOps: []string{"PATCH /recordings/{id}"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return editRecording(c, args) },
		},
		&cli.Command{
			Name: "trash", Summary: "Met à la corbeille (restaurable 30 j). Confirmation, ou --yes.", ArgSpec: "<id>", MinArgs: 1,
			APIOps: []string{"DELETE /recordings/{id}"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return trashRecording(c, args[0]) },
		},
	)
	cmd.Sub = append(cmd.Sub, spineGestures(spec)...)
	return cmd
}

// spineGestures — les gestes hérités par tout recordable (ISC-405/440).
func spineGestures(spec typeSpec) []*cli.Command {
	g := func(name, summary, method, suffix string, body func(args []string) any, ops ...string) *cli.Command {
		return &cli.Command{
			Name: name, Summary: summary, ArgSpec: "<id>", MinArgs: 1, APIOps: ops,
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var payload any
				if body != nil {
					payload = body(args[1:])
				}
				var out map[string]any
				if err := client.Do(method, "/recordings/"+args[0]+suffix, payload, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: summary + " ✓"}, nil
			},
		}
	}
	cmds := []*cli.Command{
		{
			Name: "comment", Summary: "Commente le recording.", ArgSpec: "<id> <corps html>", MinArgs: 2,
			APIOps: []string{"POST /recordings/{id}/comments"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Post("/recordings/"+args[0]+"/comments", map[string]any{"body": strings.Join(args[1:], " ")}, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: "Commentaire posté."}, nil
			},
		},
		g("boost", "Booste (réaction courte).", "POST", "/boost", func(a []string) any {
			content := "👍"
			if len(a) > 0 {
				content = a[0]
			}
			return map[string]any{"content": content}
		}, "POST /recordings/{id}/boost"),
		g("unboost", "Retire son boost.", "DELETE", "/boost", nil, "DELETE /recordings/{id}/boost"),
		g("subscribe", "S'abonne aux notifications du recording.", "POST", "/subscription", nil, "POST /recordings/{id}/subscription"),
		g("unsubscribe", "Se désabonne.", "DELETE", "/subscription", nil, "DELETE /recordings/{id}/subscription"),
		g("read", "Marque lu.", "POST", "/read", nil, "POST /recordings/{id}/read"),
		g("bookmark", "Pose un marque-page (idempotent).", "POST", "/bookmark", nil, "POST /recordings/{id}/bookmark"),
		g("unbookmark", "Retire son marque-page.", "DELETE", "/bookmark", nil, "DELETE /recordings/{id}/bookmark"),
		g("archive", "Archive (avec les enfants).", "POST", "/archive", nil, "POST /recordings/{id}/archive"),
		g("restore", "Restaure depuis l'archive ou la corbeille.", "POST", "/restore", nil, "POST /recordings/{id}/restore"),
	}
	// move/copy : changer de projet — l'atterrissage se fait dans le même outil.
	moveCopy := func(name, verb, path string) *cli.Command {
		return &cli.Command{
			Name: name, Summary: verb + " vers un autre projet (même outil, descendance comprise).",
			ArgSpec: "<id> --to <project_id>", MinArgs: 1,
			Flags:  []cli.Flag{{Name: "to", Arg: "project_id", Help: "Le projet cible."}},
			APIOps: []string{strings.ToUpper(pathVerb(name)) + " /recordings/{id}/" + path},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				to, rest := cli.FlagValue(args, "to")
				if to == "" || len(rest) == 0 {
					return nil, cli.Usagef("usage : … %s <id> --to <project_id>", name)
				}
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				method := "PATCH"
				if name == "copy" {
					method = "POST"
				}
				if err := client.Do(method, "/recordings/"+rest[0]+"/"+path, map[string]any{"project_id": to}, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: verb + " ✓"}, nil
			},
		}
	}
	cmds = append(cmds, moveCopy("move", "Déplace", "move"), moveCopy("copy", "Copie", "copy"))

	if spec.Type == "Todo" || spec.Type == "Card" {
		cmds = append(cmds, &cli.Command{
			Name: "assign", Summary: "Assigne — REMPLACE la liste des assignés (idempotent, membres du projet seulement).",
			ArgSpec: "<id> <user_id…>", MinArgs: 1,
			APIOps: []string{"PATCH /recordings/{id}/assignees"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Patch("/recordings/"+args[0]+"/assignees", map[string]any{"user_ids": args[1:]}, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: "Assignés remplacés."}, nil
			},
		})
	}
	if spec.Type == "Card" {
		cmds = append(cmds, &cli.Command{
			Name: "move-column", Summary: "Déplace la carte de colonne — entrer dans done horodate la complétion.",
			ArgSpec: "<id> --to <column_recording_id>", MinArgs: 1,
			Flags:  []cli.Flag{{Name: "to", Arg: "id", Help: "Le recording de la colonne cible."}},
			APIOps: []string{"PATCH /recordings/{id}/column"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				to, rest := cli.FlagValue(args, "to")
				if to == "" || len(rest) == 0 {
					return nil, cli.Usagef("usage : … move-column <id> --to <column_recording_id>")
				}
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Patch("/recordings/"+rest[0]+"/column", map[string]any{"column_id": to}, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: "Carte déplacée."}, nil
			},
		})
	}
	return cmds
}

func pathVerb(name string) string {
	if name == "copy" {
		return "POST"
	}
	return "PATCH"
}

// ── exécution partagée ─────────────────────────────────────────────────────

func listRecordings(c *cli.Ctx, spec typeSpec, args []string) (*cli.Result, error) {
	parent, _ := cli.FlagValue(args, "parent")
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	items := []map[string]any{}
	offset := 0
	for {
		path := fmt.Sprintf("/recordings?type=%s&offset=%d", url.QueryEscape(spec.Type), offset)
		if c.Flags.Project != "" {
			path += "&project_id=" + url.QueryEscape(c.Flags.Project)
		}
		var page struct {
			Recordings []map[string]any `json:"recordings"`
		}
		if err := client.Get(path, &page); err != nil {
			return nil, err
		}
		items = append(items, page.Recordings...)
		if len(page.Recordings) < 100 {
			break
		}
		offset += 100
	}
	if parent != "" {
		filtered := items[:0]
		for _, r := range items {
			if str(r["parent_id"]) == parent {
				filtered = append(filtered, r)
			}
		}
		items = filtered
	}
	rows := [][]string{}
	for _, r := range items {
		rows = append(rows, []string{str(r["id"]), str(r["title"]), str(r["status"]), str(r["bucket_id"])})
	}
	crumbs := []cli.Crumb{{Action: "voir un élément", Cmd: "terranova " + spec.Cmd + " show <id>"}}
	if !spec.NoCreate {
		crumbs = append(crumbs, cli.Crumb{Action: "en créer un", Cmd: "terranova " + spec.Cmd + " add <titre> --project <id>"})
	}
	return &cli.Result{Data: items, Headers: []string{"ID", "TITRE", "STATUT", "PROJET"}, Rows: rows,
		Summary: fmt.Sprintf("%d %s.", len(items), spec.Cmd), Crumbs: crumbs}, nil
}

func showRecording(c *cli.Ctx, spec typeSpec, id string) (*cli.Result, error) {
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	var out struct {
		Recording map[string]any `json:"recording"`
	}
	if err := client.Get("/recordings/"+id, &out); err != nil {
		return nil, err
	}
	return &cli.Result{Data: out.Recording,
		Crumbs: []cli.Crumb{
			{Action: "commenter", Cmd: "terranova " + spec.Cmd + " comment " + id + " <corps>"},
			{Action: "s'abonner", Cmd: "terranova " + spec.Cmd + " subscribe " + id},
		}}, nil
}

func createRecording(c *cli.Ctx, spec typeSpec, args []string) (*cli.Result, error) {
	body := map[string]any{"type": spec.Type}
	// Drapeaux déclarés → params snake_case ; les booléens déclarés sans Arg
	// passent true.
	for _, f := range spec.Create {
		if f.Arg == "" {
			on, rest := cli.FlagBool(args, f.Name)
			args = rest
			if on {
				body[snake(f.Name)] = true
			}
			continue
		}
		v, rest := cli.FlagValue(args, f.Name)
		args = rest
		if v != "" {
			body[snake(f.Name)] = v
		}
	}
	parent, args := cli.FlagValue(args, "parent")
	project, args := cli.FlagValue(args, "project")
	if project == "" {
		project = c.Flags.Project
	}
	if parent != "" {
		body["parent_id"] = parent
	}
	if project != "" {
		body["project_id"] = project
	}
	if len(args) > 0 {
		body["title"] = strings.Join(args, " ")
	}
	if spec.Type != "Contact" && project == "" {
		return nil, cli.Usagef("--project est obligatoire pour créer un %s", spec.Type)
	}
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	// Upload : le fichier passe en multipart, comme à l'écran (ISC-400).
	if spec.Type == "Upload" {
		file := str(body["file"])
		if file == "" {
			return nil, cli.Usagef("--file est obligatoire pour un upload")
		}
		delete(body, "file")
		fields := map[string]string{}
		for k, v := range body {
			fields[k] = str(v)
		}
		var out map[string]any
		if err := client.Upload("/recordings", fields, "file", file, &out); err != nil {
			return nil, err
		}
		return &cli.Result{Data: out, Summary: "Fichier téléversé."}, nil
	}
	var out map[string]any
	if err := client.Post("/recordings", body, &out); err != nil {
		return nil, err
	}
	id := str(dig(anyMap(out), "recording", "id"))
	return &cli.Result{Data: out, Summary: spec.Type + " créé.",
		Crumbs: []cli.Crumb{{Action: "le voir", Cmd: "terranova " + spec.Cmd + " show " + id}}}, nil
}

func editRecording(c *cli.Ctx, args []string) (*cli.Result, error) {
	id := args[0]
	body := map[string]any{}
	rest := args[1:]
	for _, key := range []string{"title", "body", "due-on", "completed"} {
		v, r := cli.FlagValue(rest, key)
		rest = r
		if v != "" {
			body[snake(key)] = v
		}
	}
	if len(body) == 0 {
		return nil, cli.Usagef("rien à modifier — passe --title, --body, --due-on ou --completed")
	}
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := client.Patch("/recordings/"+id, body, &out); err != nil {
		return nil, err
	}
	return &cli.Result{Data: out, Summary: "Modifié."}, nil
}

func trashRecording(c *cli.Ctx, id string) (*cli.Result, error) {
	if err := confirm(c, fmt.Sprintf("Mettre le recording %s à la corbeille ?", id)); err != nil {
		return nil, err
	}
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := client.Delete("/recordings/"+id, &out); err != nil {
		return nil, err
	}
	return &cli.Result{Data: out, Summary: "À la corbeille (restaurable 30 jours).",
		Crumbs: []cli.Crumb{{Action: "restaurer", Cmd: "terranova recordings restore " + id}}}, nil
}

// buildRecordingsCommand — le parcours générique par type/statut (ISC-406).
func buildRecordingsCommand() *cli.Command {
	generic := typeSpec{Cmd: "recordings", Type: "", Summary: "La spine brute : tout recordable, par type et statut."}
	cmd := &cli.Command{
		Name: "recordings", Group: "Search & Browse",
		Summary: "La spine brute : parcourir par type, corbeille/archive/restauration, déplacer/copier.",
		APIOps:  []string{"GET /recordings", "GET /recordings/{id}", "PATCH /recordings/{id}", "DELETE /recordings/{id}"},
		Sub: []*cli.Command{
			{
				Name: "list", Summary: "Liste — filtres --type, --project, --assigned-to.",
				Flags: []cli.Flag{
					{Name: "type", Arg: "Type", Help: "Un des 40 recordable_type (Todo, Message…)."},
					{Name: "assigned-to", Arg: "user_id", Help: "Tâches assignées à cette personne."},
				},
				APIOps: []string{"GET /recordings"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					typ, args := cli.FlagValue(args, "type")
					assigned, _ := cli.FlagValue(args, "assigned-to")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/recordings?offset=0"
					if typ != "" {
						path += "&type=" + url.QueryEscape(typ)
					}
					if c.Flags.Project != "" {
						path += "&project_id=" + url.QueryEscape(c.Flags.Project)
					}
					if assigned != "" {
						path += "&assigned_to=" + url.QueryEscape(assigned)
					}
					var page struct {
						Recordings []map[string]any `json:"recordings"`
					}
					if err := client.Get(path, &page); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, r := range page.Recordings {
						rows = append(rows, []string{str(r["id"]), str(r["recordable_type"]), str(r["title"]), str(r["status"])})
					}
					return &cli.Result{Data: page.Recordings, Headers: []string{"ID", "TYPE", "TITRE", "STATUT"}, Rows: rows,
						Summary: fmt.Sprintf("%d recordings (page de 100 max — --jq et offset pour plus).", len(page.Recordings))}, nil
				},
			},
			{
				Name: "show", Summary: "Détail d'un recording, quel que soit son type.", ArgSpec: "<id>", MinArgs: 1,
				APIOps: []string{"GET /recordings/{id}"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					return showRecording(c, generic, args[0])
				},
			},
		},
	}
	cmd.Sub = append(cmd.Sub, spineGestures(generic)...)
	cmd.Sub = append(cmd.Sub, &cli.Command{
		Name: "trash", Summary: "Corbeille (confirmation, ou --yes).", ArgSpec: "<id>", MinArgs: 1,
		APIOps: []string{"DELETE /recordings/{id}"},
		Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return trashRecording(c, args[0]) },
	})
	return cmd
}

// ── utilitaires ────────────────────────────────────────────────────────────

func snake(kebab string) string { return strings.ReplaceAll(kebab, "-", "_") }

func anyMap(m map[string]any) map[string]any { return m }

// confirm applique ISC-381 : confirmation sur les gestes destructifs, levée par
// --yes ; en mode --agent, l'absence de --yes est une erreur d'usage.
func confirm(c *cli.Ctx, question string) error {
	if c.Flags.Yes {
		return nil
	}
	if c.Flags.Agent || !c.IsTTY {
		return cli.Usagef("geste destructif : ajoute --yes pour confirmer (aucun prompt en mode agent)")
	}
	fmt.Printf("%s (oui/non) ", question)
	var answer string
	fmt.Scanln(&answer)
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "oui" && answer != "o" && answer != "yes" && answer != "y" {
		return fmt.Errorf("annulé")
	}
	return nil
}
