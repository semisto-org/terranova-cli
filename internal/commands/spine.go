package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
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
	{Cmd: "todosets", Type: "Todoset", Group: msg.GroupCore, Summary: msg.HelpTodosets, NoCreate: true},
	{Cmd: "todolists", Type: "Todolist", Group: msg.GroupCore, Summary: msg.HelpTodolists, Parent: "todoset",
		Create: []cli.Flag{{Name: "description", Arg: "texte", Help: msg.FlagTodolistsDescription}}},
	{Cmd: "todogroups", Type: "TodoGroup", Group: msg.GroupCore, Summary: msg.HelpTodogroups, Parent: "todolist"},
	{Cmd: "todos", Type: "Todo", Group: msg.GroupCore, Summary: msg.HelpTodos, Parent: "todolist",
		Create: []cli.Flag{
			{Name: "due-on", Arg: "date", Help: msg.FlagDueOn},
			{Name: "starts-on", Arg: "date", Help: msg.FlagTodosStartsOn},
			{Name: "recurrence", Arg: "règle", Help: msg.FlagTodosRecurrence},
			{Name: "recurrence-until", Arg: "date", Help: msg.FlagTodosRecurrenceUntil},
		}},
	{Cmd: "messageboards", Type: "MessageBoard", Group: msg.GroupCore, Summary: msg.HelpMessageboards, NoCreate: true},
	{Cmd: "messages", Type: "Message", Group: msg.GroupCore, Summary: msg.HelpMessages, Parent: "messageboard",
		Create: []cli.Flag{
			{Name: "body", Arg: "html", Help: msg.FlagBodyHTML},
			{Name: "category", Arg: "clé", Help: msg.FlagMessagesCategory},
		}},
	{Cmd: "cardtables", Type: "CardTable", Group: msg.GroupCore, Summary: msg.HelpCardtables, NoCreate: true},
	{Cmd: "cardcolumns", Type: "CardColumn", Group: msg.GroupCore, Summary: msg.HelpCardcolumns, Parent: "cardtable",
		Create: []cli.Flag{{Name: "kind", Arg: "genre", Help: msg.FlagCardcolumnsKind}}},
	{Cmd: "cards", Type: "Card", Group: msg.GroupCore, Summary: msg.HelpCards, Parent: "cardcolumn",
		Create: []cli.Flag{
			{Name: "due-on", Arg: "date", Help: msg.FlagDueOn},
			{Name: "body", Arg: "html", Help: msg.FlagCardsBody},
		}},
	{Cmd: "folders", Type: "Vault", Group: msg.GroupFilesDocs, Summary: msg.HelpFolders, Parent: "vault parent"},
	{Cmd: "docs", Type: "Document", Group: msg.GroupFilesDocs, Summary: msg.HelpDocs, Parent: "vault",
		Create: []cli.Flag{{Name: "body", Arg: "html", Help: msg.FlagBodyHTML}}},
	{Cmd: "uploads", Type: "Upload", Group: msg.GroupFilesDocs, Summary: msg.HelpUploads, Parent: "vault",
		Create: []cli.Flag{
			{Name: "file", Arg: "chemin", Help: msg.FlagUploadsFile},
			{Name: "caption", Arg: "texte", Help: msg.FlagUploadsCaption},
		}},
	{Cmd: "links", Type: "CloudLink", Group: msg.GroupFilesDocs, Summary: msg.HelpLinks, Parent: "vault",
		Create: []cli.Flag{
			{Name: "url", Arg: "url", Help: msg.FlagLinksUrl},
			{Name: "service", Arg: "nom", Help: msg.FlagLinksService},
			{Name: "notes", Arg: "texte", Help: msg.FlagNotes},
		}},
	{Cmd: "chat", Type: "Campfire", Group: msg.GroupCore, Summary: msg.HelpChat, NoCreate: true},
	{Cmd: "checkins", Type: "Questionnaire", Group: msg.GroupCore, Summary: msg.HelpCheckins, NoCreate: true},
	{Cmd: "questions", Type: "Question", Group: msg.GroupCore, Summary: msg.HelpQuestions, Parent: "questionnaire",
		Create: []cli.Flag{{Name: "frequency", Arg: "règle", Help: msg.FlagQuestionsFrequency}}},
	{Cmd: "answers", Type: "Answer", Group: msg.GroupCore, Summary: msg.HelpAnswers, Parent: "question",
		Create: []cli.Flag{{Name: "body", Arg: "html", Help: msg.FlagAnswersBody}}},
	{Cmd: "schedules", Type: "Schedule", Group: msg.GroupCore, Summary: msg.HelpSchedules, NoCreate: true},
	{Cmd: "events", Type: "ScheduleEntry", Group: msg.GroupCore, Summary: msg.HelpEvents, Parent: "schedule",
		Create: []cli.Flag{
			{Name: "starts-at", Arg: "horodate", Help: msg.FlagEventsStartsAt},
			{Name: "ends-at", Arg: "horodate", Help: msg.FlagEventsEndsAt},
			{Name: "all-day", Help: msg.FlagEventsAllDay},
			{Name: "location", Arg: "lieu", Help: msg.FlagEventsLocation},
		}},
	{Cmd: "forwards", Type: "InboxForward", Group: msg.GroupCommunication, Summary: msg.HelpForwards, Parent: "inbox",
		Create: []cli.Flag{
			{Name: "from-email", Arg: "email", Help: msg.FlagForwardsFromEmail},
			{Name: "subject", Arg: "objet", Help: msg.FlagForwardsSubject},
		}},
	{Cmd: "hillcharts", Type: "HillChartSnapshot", Group: msg.GroupCore, Summary: msg.HelpHillcharts, NoCreate: true},
	{Cmd: "comments", Type: "Comment", Group: msg.GroupCommunication, Summary: msg.HelpComments, Parent: "recording",
		Create: []cli.Flag{{Name: "body", Arg: "html", Help: msg.FlagCommentsBody}}},
	{Cmd: "timesheets", Type: "Timesheet", Group: msg.GroupSchedulingTime, Summary: msg.HelpTimesheets, Parent: "projet",
		Create: []cli.Flag{
			{Name: "worked-on", Arg: "date", Help: msg.FlagTimesheetsWorkedOn},
			{Name: "hours", Arg: "n", Help: msg.FlagTimesheetsHours},
			{Name: "billable", Help: msg.FlagTimesheetsBillable},
			{Name: "description", Arg: "texte", Help: msg.FlagDescription},
			{Name: "phase", Arg: "phase", Help: msg.FlagTimesheetsPhase},
		}},

	// ── Contacto (ISC-413) ──
	{Cmd: "contacts", Type: "Contact", Group: msg.GroupLentilles, Summary: msg.HelpContacts,
		Create: []cli.Flag{{Name: "kind", Arg: "genre", Help: msg.FlagContactsKind}}},

	// ── Academio (ISC-415, la part recordable) ──
	{Cmd: "participant-lists", Type: "ParticipantList", Group: msg.GroupLentilles, Summary: msg.HelpParticipantLists, NoCreate: true},
	{Cmd: "participant-categories", Type: "ParticipantCategory", Group: msg.GroupLentilles, Summary: msg.HelpParticipantCategories, Parent: "participant-list",
		Create: []cli.Flag{
			{Name: "label", Arg: "libellé", Help: msg.FlagParticipantCategoriesLabel},
			{Name: "price-cents", Arg: "n", Help: msg.FlagParticipantCategoriesPriceCents},
			{Name: "quota", Arg: "n", Help: msg.FlagParticipantCategoriesQuota},
		}},
	{Cmd: "participants", Type: "Participant", Group: msg.GroupLentilles, Summary: msg.HelpParticipants, Parent: "participant-list",
		Create: []cli.Flag{
			{Name: "contact-id", Arg: "id", Help: msg.FlagParticipantsContactId},
			{Name: "participant-category-id", Arg: "id", Help: msg.FlagParticipantsParticipantCategoryId},
			{Name: "status", Arg: "statut", Help: msg.FlagParticipantsStatus},
		}},
	{Cmd: "training-documents", Type: "TrainingDocument", Group: msg.GroupLentilles, Summary: msg.HelpTrainingDocuments, Parent: "projet",
		Create: []cli.Flag{
			{Name: "name", Arg: "nom", Help: msg.FlagTrainingDocumentsName},
			{Name: "schedule-entry-id", Arg: "id", Help: msg.FlagTrainingDocumentsScheduleEntryId},
		}},
	{Cmd: "session-feedbacks", Type: "SessionFeedback", Group: msg.GroupLentilles, Summary: msg.HelpSessionFeedbacks, Parent: "projet",
		Create: []cli.Flag{
			{Name: "schedule-entry-id", Arg: "id", Help: msg.FlagSessionFeedbacksScheduleEntryId},
			{Name: "contact-id", Arg: "id", Help: msg.FlagParticipant},
			{Name: "rating", Arg: "n", Help: msg.FlagSessionFeedbacksRating},
			{Name: "comment", Arg: "texte", Help: msg.FlagSessionFeedbacksComment},
		}},

	// ── Conceptio (ISC-416, la part recordable) ──
	{Cmd: "palettes", Type: "Palette", Group: msg.GroupLentilles, Summary: msg.HelpPalettes, NoCreate: true},
	{Cmd: "palette-items", Type: "PaletteItem", Group: msg.GroupLentilles, Summary: msg.HelpPaletteItems, Parent: "palette",
		Create: []cli.Flag{
			{Name: "name", Arg: "latin", Help: msg.FlagNomLatin},
			{Name: "common-name", Arg: "nom", Help: msg.FlagPaletteItemsCommonName},
			{Name: "strata", Arg: "strate", Help: msg.FlagPaletteItemsStrata},
			{Name: "quantity", Arg: "n", Help: msg.FlagQuantite},
			{Name: "unit-price-cents", Arg: "n", Help: msg.FlagUnitPriceCents},
			{Name: "external-variety-id", Arg: "id", Help: msg.FlagPaletteItemsExternalVarietyId},
		}},
	{Cmd: "concepts", Type: "Concept", Group: msg.GroupLentilles, Summary: msg.HelpConcepts, Parent: "projet design",
		Create: []cli.Flag{
			{Name: "width-m", Arg: "m", Help: msg.FlagConceptsWidthM},
			{Name: "height-m", Arg: "m", Help: msg.FlagConceptsHeightM},
		}},
	{Cmd: "markers", Type: "PlantMarker", Group: msg.GroupLentilles, Summary: msg.HelpMarkers, Parent: "concept",
		Create: []cli.Flag{
			{Name: "x", Arg: "0..1", Help: msg.FlagMarkersX},
			{Name: "y", Arg: "0..1", Help: msg.FlagMarkersY},
			{Name: "species-name", Arg: "nom", Help: msg.FlagMarkersSpeciesName},
			{Name: "palette-item-id", Arg: "id", Help: msg.FlagMarkersPaletteItemId},
		}},
	{Cmd: "quotes", Type: "Quote", Group: msg.GroupLentilles, Summary: msg.HelpQuotes, Parent: "projet design",
		Create: []cli.Flag{
			{Name: "vat-rate", Arg: "taux", Help: msg.FlagQuotesVatRate},
			{Name: "valid-until", Arg: "date", Help: msg.FlagQuotesValidUntil},
		}},
	{Cmd: "quote-lines", Type: "QuoteLine", Group: msg.GroupLentilles, Summary: msg.HelpQuoteLines, Parent: "quote",
		Create: []cli.Flag{
			{Name: "description", Arg: "texte", Help: msg.FlagDescription},
			{Name: "quantity", Arg: "n", Help: msg.FlagQuantite},
			{Name: "unit", Arg: "unité", Help: msg.FlagQuoteLinesUnit},
			{Name: "unit-price-cents", Arg: "n", Help: msg.FlagUnitPriceCents},
		}},
	{Cmd: "plant-records", Type: "PlantRecord", Group: msg.GroupLentilles, Summary: msg.HelpPlantRecords, Parent: "projet design",
		Create: []cli.Flag{
			{Name: "status", Arg: "statut", Help: msg.FlagPlantRecordsStatus},
			{Name: "plant-marker-id", Arg: "id", Help: msg.FlagPlantRecordsPlantMarkerId},
		}},
	{Cmd: "interventions", Type: "Intervention", Group: msg.GroupLentilles, Summary: msg.HelpInterventions, Parent: "projet design",
		Create: []cli.Flag{
			{Name: "happened-on", Arg: "date", Help: msg.FlagDate},
			{Name: "kind", Arg: "genre", Help: msg.FlagInterventionsKind},
			{Name: "notes", Arg: "texte", Help: msg.FlagNotes},
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
			Name: "list", Summary: fmt.Sprintf(msg.HelpSpineList, spec.Type),
			Flags:  []cli.Flag{{Name: "parent", Arg: "id", Help: msg.FlagSpineListParent}},
			APIOps: []string{"GET /recordings"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return listRecordings(c, spec, args) },
		},
		&cli.Command{
			Name: "show", Summary: msg.HelpSpineShow, ArgSpec: "<id>", MinArgs: 1,
			APIOps: []string{"GET /recordings/{id}"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return showRecording(c, spec, args[0]) },
		},
	)
	if !spec.NoCreate {
		parentHelp := msg.FlagSpineParent
		if spec.Parent != "" {
			parentHelp = fmt.Sprintf(msg.FlagSpineParentIn, spec.Parent)
		}
		addFlags := append([]cli.Flag{
			{Name: "project", Short: "p", Arg: "id", Help: msg.FlagSpineProject},
			{Name: "parent", Arg: "id", Help: parentHelp},
		}, spec.Create...)
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "add", Summary: fmt.Sprintf(msg.HelpSpineAdd, spec.Type), ArgSpec: "<titre>", MinArgs: 0,
			Flags:  addFlags,
			APIOps: []string{"POST /recordings"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return createRecording(c, spec, args) },
		})
	}
	cmd.Sub = append(cmd.Sub,
		&cli.Command{
			Name: "edit", Summary: msg.HelpSpineEdit, ArgSpec: "<id>", MinArgs: 1,
			Flags: []cli.Flag{
				{Name: "title", Arg: "titre", Help: msg.FlagSpineEditTitle},
				{Name: "body", Arg: "html", Help: msg.FlagSpineEditBody},
				{Name: "due-on", Arg: "date", Help: msg.FlagSpineEditDueOn},
				{Name: "completed", Arg: "bool", Help: msg.FlagSpineEditCompleted},
			},
			APIOps: []string{"PATCH /recordings/{id}"},
			Run:    func(c *cli.Ctx, args []string) (*cli.Result, error) { return editRecording(c, args) },
		},
		&cli.Command{
			Name: "trash", Summary: msg.HelpSpineTrash, ArgSpec: "<id>", MinArgs: 1,
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
			Name: "comment", Summary: msg.HelpSpineComment, ArgSpec: "<id> <corps html>", MinArgs: 2,
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
				return &cli.Result{Data: out, Summary: msg.ResCommentairePoste}, nil
			},
		},
		g("boost", msg.HelpBoost, "POST", "/boost", func(a []string) any {
			content := "👍"
			if len(a) > 0 {
				content = a[0]
			}
			return map[string]any{"content": content}
		}, "POST /recordings/{id}/boost"),
		g("unboost", msg.HelpUnboost, "DELETE", "/boost", nil, "DELETE /recordings/{id}/boost"),
		g("subscribe", msg.HelpSubscribe, "POST", "/subscription", nil, "POST /recordings/{id}/subscription"),
		g("unsubscribe", msg.HelpUnsubscribe, "DELETE", "/subscription", nil, "DELETE /recordings/{id}/subscription"),
		g("read", msg.HelpRead, "POST", "/read", nil, "POST /recordings/{id}/read"),
		g("bookmark", msg.HelpBookmark, "POST", "/bookmark", nil, "POST /recordings/{id}/bookmark"),
		g("unbookmark", msg.HelpUnbookmark, "DELETE", "/bookmark", nil, "DELETE /recordings/{id}/bookmark"),
		g("bubble-up", msg.HelpBubbleUp, "POST", "/bubble_up", func(a []string) any {
			if len(a) > 0 {
				return map[string]any{"when": a[0]}
			}
			return nil
		}, "POST /recordings/{id}/bubble_up"),
		g("unbubble-up", msg.HelpUnbubbleUp, "DELETE", "/bubble_up", nil, "DELETE /recordings/{id}/bubble_up"),
		g("archive", msg.HelpArchive, "POST", "/archive", nil, "POST /recordings/{id}/archive"),
		g("restore", msg.HelpRestore, "POST", "/restore", nil, "POST /recordings/{id}/restore"),
	}
	// move/copy : changer de projet — l'atterrissage se fait dans le même outil.
	moveCopy := func(name, verb, path string) *cli.Command {
		return &cli.Command{
			Name: name, Summary: fmt.Sprintf(msg.HelpMoveCopy, verb),
			ArgSpec: "<id> --to <project_id>", MinArgs: 1,
			Flags:  []cli.Flag{{Name: "to", Arg: "project_id", Help: msg.FlagSpineMoveTo}},
			APIOps: []string{strings.ToUpper(pathVerb(name)) + " /recordings/{id}/" + path},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				to, rest := cli.FlagValue(args, "to")
				if to == "" || len(rest) == 0 {
					return nil, cli.Usagef(msg.UsageMoveCopy, name)
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
	cmds = append(cmds, moveCopy("move", msg.VerbMove, "move"), moveCopy("copy", msg.VerbCopy, "copy"))

	if spec.Type == "Todo" || spec.Type == "Card" {
		cmds = append(cmds, &cli.Command{
			Name: "assign", Summary: msg.HelpAssign,
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
				return &cli.Result{Data: out, Summary: msg.ResAssignesRemplaces}, nil
			},
		})
	}
	if spec.Type == "Card" {
		cmds = append(cmds, &cli.Command{
			Name: "move-column", Summary: msg.HelpMoveColumn,
			ArgSpec: "<id> --to <column_recording_id>", MinArgs: 1,
			Flags:  []cli.Flag{{Name: "to", Arg: "id", Help: msg.FlagMoveColumnTo}},
			APIOps: []string{"PATCH /recordings/{id}/column"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				to, rest := cli.FlagValue(args, "to")
				if to == "" || len(rest) == 0 {
					return nil, cli.Usagef(msg.UsageMoveColumn)
				}
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Patch("/recordings/"+rest[0]+"/column", map[string]any{"column_id": to}, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: msg.ResCarteDeplacee}, nil
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
	crumbs := []cli.Crumb{{Action: msg.CrumbVoirUnElement, Cmd: "terranova " + spec.Cmd + " show <id>"}}
	if !spec.NoCreate {
		crumbs = append(crumbs, cli.Crumb{Action: msg.CrumbEnCreerUn, Cmd: "terranova " + spec.Cmd + " add <titre> --project <id>"})
	}
	return &cli.Result{Data: items, Headers: msg.HeadersRecordingList, Rows: rows,
		Summary: fmt.Sprintf(msg.ResListCount, len(items), spec.Cmd), Crumbs: crumbs}, nil
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
			{Action: msg.CrumbCommenter, Cmd: "terranova " + spec.Cmd + " comment " + id + " <corps>"},
			{Action: msg.CrumbSAbonner, Cmd: "terranova " + spec.Cmd + " subscribe " + id},
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
		return nil, cli.Usagef(msg.UsageProjectRequired, spec.Type)
	}
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	// Upload : le fichier passe en multipart, comme à l'écran (ISC-400).
	if spec.Type == "Upload" {
		file := str(body["file"])
		if file == "" {
			return nil, cli.Usagef(msg.UsageFileRequired)
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
		return &cli.Result{Data: out, Summary: msg.ResFichierTeleverse}, nil
	}
	var out map[string]any
	if err := client.Post("/recordings", body, &out); err != nil {
		return nil, err
	}
	id := str(dig(anyMap(out), "recording", "id"))
	return &cli.Result{Data: out, Summary: fmt.Sprintf(msg.ResCreatedType, spec.Type),
		Crumbs: []cli.Crumb{{Action: msg.CrumbLeVoir, Cmd: "terranova " + spec.Cmd + " show " + id}}}, nil
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
		return nil, cli.Usagef(msg.UsageNothingToEditSpine)
	}
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := client.Patch("/recordings/"+id, body, &out); err != nil {
		return nil, err
	}
	return &cli.Result{Data: out, Summary: msg.ResModifie}, nil
}

func trashRecording(c *cli.Ctx, id string) (*cli.Result, error) {
	if err := confirm(c, fmt.Sprintf(msg.AskTrashRecording, id)); err != nil {
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
	return &cli.Result{Data: out, Summary: msg.ResALaCorbeilleRestaurable30,
		Crumbs: []cli.Crumb{{Action: msg.CrumbRestaurer, Cmd: "terranova recordings restore " + id}}}, nil
}

// buildRecordingsCommand — le parcours générique par type/statut (ISC-406).
func buildRecordingsCommand() *cli.Command {
	generic := typeSpec{Cmd: "recordings", Type: "", Summary: msg.HelpRecordings}
	cmd := &cli.Command{
		Name: "recordings", Group: msg.GroupSearchBrowse,
		Summary: msg.HelpRecordingsBrowse,
		APIOps:  []string{"GET /recordings", "GET /recordings/{id}", "PATCH /recordings/{id}", "DELETE /recordings/{id}"},
		Sub: []*cli.Command{
			{
				Name: "list", Summary: msg.HelpRecordingsList,
				Flags: []cli.Flag{
					{Name: "type", Arg: "Type", Help: msg.FlagRecordingsListType},
					{Name: "assigned-to", Arg: "user_id", Help: msg.FlagRecordingsListAssignedTo},
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
					return &cli.Result{Data: page.Recordings, Headers: msg.HeadersRecordingKinds, Rows: rows,
						Summary: fmt.Sprintf(msg.ResRecordingsPage, len(page.Recordings))}, nil
				},
			},
			{
				Name: "show", Summary: msg.HelpRecordingsShow, ArgSpec: "<id>", MinArgs: 1,
				APIOps: []string{"GET /recordings/{id}"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					return showRecording(c, generic, args[0])
				},
			},
		},
	}
	cmd.Sub = append(cmd.Sub, spineGestures(generic)...)
	cmd.Sub = append(cmd.Sub, &cli.Command{
		Name: "trash", Summary: msg.HelpRecordingsTrash, ArgSpec: "<id>", MinArgs: 1,
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
		return cli.Usagef(msg.UsageDestructiveNeedsYes)
	}
	fmt.Printf(msg.AskOuiNon, question)
	var answer string
	fmt.Scanln(&answer)
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "oui" && answer != "o" && answer != "yes" && answer != "y" {
		return fmt.Errorf(msg.ErrCancelled)
	}
	return nil
}
