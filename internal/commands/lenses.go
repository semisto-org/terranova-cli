package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

// restSpec décrit une ressource REST non-recordable (finances, botanique,
// pépinière, membres — le miroir du registre ISC-224/423).
type restSpec struct {
	Cmd     string
	Base    string // chemin API ("/plant/genera")
	Key     string // clé racine JSON au pluriel ("genera") — vide = déduite
	Summary string
	List    bool
	Show    bool
	Create  []cli.Flag // nil = pas de création
	Update  []cli.Flag // nil = pas d'édition
	// UpdateMethod : PATCH par défaut ; les hubs s'éditent en PUT.
	UpdateMethod string
	Delete       bool
	Filters      []cli.Flag // filtres de l'index (q, genus_id…)
	Scope        string     // note d'aide : le scope exigé
}

func (s restSpec) key() string {
	if s.Key != "" {
		return s.Key
	}
	parts := strings.Split(strings.Trim(s.Base, "/"), "/")
	return parts[len(parts)-1]
}

// buildRestCommand fabrique list/show/add/edit/delete d'une ressource REST.
func buildRestCommand(s restSpec) *cli.Command {
	cmd := &cli.Command{Name: s.Cmd, Summary: s.Summary}
	if s.Scope != "" {
		cmd.AgentHelp = fmt.Sprintf(msg.NotesScopeRequired, s.Scope)
	}
	if s.List {
		cmd.APIOps = append(cmd.APIOps, "GET "+s.Base)
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "list", Summary: msg.HelpRestList, Flags: s.Filters,
			APIOps: []string{"GET " + s.Base},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				params := map[string]string{}
				for _, f := range s.Filters {
					v, rest := cli.FlagValue(args, f.Name)
					args = rest
					if v != "" {
						params[snake(f.Name)] = v
					}
				}
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Get(s.Base+query(params), &out); err != nil {
					return nil, err
				}
				items, _ := out[s.key()].([]any)
				return &cli.Result{Data: out, Summary: fmt.Sprintf(msg.ResElementCount, len(items))}, nil
			},
		})
	}
	if s.Show {
		cmd.APIOps = append(cmd.APIOps, "GET "+s.Base+"/{id}")
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "show", Summary: msg.HelpRestShow, ArgSpec: "<id>", MinArgs: 1,
			APIOps: []string{"GET " + s.Base + "/{id}"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Get(s.Base+"/"+args[0], &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out}, nil
			},
		})
	}
	if s.Create != nil {
		cmd.APIOps = append(cmd.APIOps, "POST "+s.Base)
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "add", Summary: msg.HelpRestAdd, Flags: s.Create,
			APIOps: []string{"POST " + s.Base},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				body, _ := flagsToBody(s.Create, args)
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Post(s.Base, body, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: msg.ResCree}, nil
			},
		})
	}
	if s.Update != nil {
		method := s.UpdateMethod
		if method == "" {
			method = "PATCH"
		}
		cmd.APIOps = append(cmd.APIOps, method+" "+s.Base+"/{id}")
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "edit", Summary: msg.HelpRestEdit, ArgSpec: "<id>", MinArgs: 1, Flags: s.Update,
			APIOps: []string{method + " " + s.Base + "/{id}"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				body, _ := flagsToBody(s.Update, args[1:])
				if len(body) == 0 {
					return nil, cli.Usagef(msg.UsageNothingToEditRest)
				}
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Do(method, s.Base+"/"+args[0], body, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: msg.ResModifie}, nil
			},
		})
	}
	if s.Delete {
		cmd.APIOps = append(cmd.APIOps, "DELETE "+s.Base+"/{id}")
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "remove", Summary: msg.HelpRestRemove, ArgSpec: "<id>", MinArgs: 1,
			APIOps: []string{"DELETE " + s.Base + "/{id}"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				if err := confirm(c, fmt.Sprintf(msg.AskDeleteResource, s.Base, args[0])); err != nil {
					return nil, err
				}
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Delete(s.Base+"/"+args[0], &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: msg.ResSupprime}, nil
			},
		})
	}
	return cmd
}

func flagsToBody(flags []cli.Flag, args []string) (map[string]any, []string) {
	body := map[string]any{}
	for _, f := range flags {
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
	return body, args
}

func query(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	return "?" + q.Encode()
}

func group(name, groupName, summary, agentHelp string, subs ...*cli.Command) *cli.Command {
	return &cli.Command{Name: name, Group: groupName, Summary: summary, AgentHelp: agentHelp, Sub: subs}
}

func init() {
	// ── Projects & tools (ISC-395/396) ──
	projects := buildRestCommand(restSpec{
		Cmd: "projects", Base: "/projects", Summary: msg.HelpProjects,
		List: true, Show: true,
		Update: []cli.Flag{
			{Name: "name", Arg: "nom", Help: msg.FlagProjectsName},
			{Name: "description", Arg: "texte", Help: msg.FlagDescription},
			{Name: "color", Arg: "couleur", Help: msg.FlagProjectsColor},
			{Name: "status", Arg: "statut", Help: msg.FlagProjectsStatus},
		},
		Create: []cli.Flag{
			{Name: "name", Arg: "nom", Help: msg.FlagProjectsNameRequired},
			{Name: "kind", Arg: "genre", Help: msg.FlagProjectsKind},
			{Name: "description", Arg: "texte", Help: msg.FlagDescription},
			{Name: "access", Arg: "accès", Help: msg.FlagProjectsAccess},
		},
	})
	projects.Group = "Core"
	projects.Sub = append(projects.Sub, &cli.Command{
		Name: "tools", Summary: msg.HelpTools,
		ArgSpec: "<project_id>", MinArgs: 1,
		Flags: []cli.Flag{
			{Name: "install", Arg: "kind", Help: msg.FlagToolsInstall},
			{Name: "name", Arg: "nom", Help: msg.FlagToolsName},
		},
		APIOps: []string{"POST /projects/{project_id}/tools"},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			kind, rest := cli.FlagValue(args, "install")
			name, rest := cli.FlagValue(rest, "name")
			if kind == "" {
				return nil, cli.Usagef(msg.UsageProjectsTools)
			}
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			body := map[string]any{"kind": kind}
			if name != "" {
				body["name"] = name
			}
			var out map[string]any
			if err := client.Post("/projects/"+rest[0]+"/tools", body, &out); err != nil {
				return nil, err
			}
			return &cli.Result{Data: out, Summary: msg.ResOutilInstalle}, nil
		},
	})
	projects.Sub = append(projects.Sub, &cli.Command{
		Name: "people", Summary: msg.HelpPeople,
		ArgSpec: "<project_id>", MinArgs: 1,
		Flags: []cli.Flag{
			{Name: "add", Arg: "user_id", Help: msg.FlagPeopleAdd},
			{Name: "remove", Arg: "user_id", Help: msg.FlagPeopleRemove},
		},
		APIOps: []string{"GET /projects/{project_id}/people", "POST /projects/{project_id}/people", "DELETE /projects/{project_id}/people/{id}"},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			add, args := cli.FlagValue(args, "add")
			remove, args := cli.FlagValue(args, "remove")
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			pid := args[0]
			if add != "" {
				var out map[string]any
				if err := client.Post("/projects/"+pid+"/people", map[string]any{"user_id": add}, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: msg.ResAjouteAuProjet}, nil
			}
			if remove != "" {
				if err := client.Delete("/projects/"+pid+"/people/"+remove, nil); err != nil {
					return nil, err
				}
				return &cli.Result{Data: map[string]any{"removed": remove}, Summary: msg.ResRetireDuProjet}, nil
			}
			var out struct {
				People []map[string]any `json:"people"`
				Access string           `json:"access"`
			}
			if err := client.Get("/projects/"+pid+"/people", &out); err != nil {
				return nil, err
			}
			rows := [][]string{}
			for _, p := range out.People {
				rows = append(rows, []string{str(p["id"]), str(p["name"]), str(p["email"])})
			}
			return &cli.Result{Data: out, Headers: msg.HeadersPeople, Rows: rows,
				Summary: fmt.Sprintf(msg.ResPeopleCount, len(out.People), out.Access)}, nil
		},
	})
	cli.Register(projects)

	// ── People & organisation (ISC-409/420) ──
	people := buildRestCommand(restSpec{
		Cmd: "people", Base: "/members", Key: "members", Summary: msg.HelpPeopleDirectory,
		List: true, Show: true,
		Create: []cli.Flag{
			{Name: "email", Arg: "email", Help: msg.FlagPeopleEmail},
			{Name: "first-name", Arg: "prénom", Help: msg.FlagPeopleFirstName},
			{Name: "last-name", Arg: "nom", Help: msg.FlagPeopleLastName},
		},
	})
	people.Group = "Organisation"
	cli.Register(people)

	enroll := buildRestCommand(restSpec{
		Cmd: "enrollments", Base: "/member_enrollments", Summary: msg.HelpEnrollments,
		List: true, Delete: true,
		Create: []cli.Flag{
			{Name: "member-id", Arg: "id", Help: msg.FlagEnrollmentsMemberId},
			{Name: "role", Arg: "rôle", Help: msg.FlagEnrollmentsRole},
		},
	})
	enroll.Group = "Organisation"
	cli.Register(enroll)

	grants := buildRestCommand(restSpec{
		Cmd: "grants", Base: "/interface_grants", Summary: msg.HelpGrants,
		List: true, Delete: true, Scope: msg.ScopeGrants,
		Create: []cli.Flag{
			{Name: "user-id", Arg: "id", Help: msg.FlagGrantsUserId},
			{Name: "interface", Arg: "clé", Help: msg.FlagGrantsInterface},
		},
	})
	grants.Group = "Organisation"
	cli.Register(grants)

	// ── Planto (ISC-414) — lecture ouverte, écriture derrière le grant ──
	cli.Register(group("planto", msg.GroupLentilles,
		msg.HelpPlanto,
		msg.NotesPlanto,
		buildRestCommand(restSpec{Cmd: "genera", Base: "/plant/genera", Summary: msg.HelpPlantoGenera,
			List: true, Show: true, Filters: []cli.Flag{{Name: "q", Arg: "texte", Help: msg.FlagPlantoGeneraQ}},
			Create: []cli.Flag{{Name: "name", Arg: "nom", Help: msg.FlagPlantoGeneraName}},
			Update: []cli.Flag{{Name: "name", Arg: "nom", Help: msg.FlagNomLatin}, {Name: "description", Arg: "texte", Help: msg.FlagDescription}}}),
		buildRestCommand(restSpec{Cmd: "species", Base: "/plant/species", Summary: msg.HelpPlantoSpecies,
			List: true, Show: true,
			Filters: []cli.Flag{{Name: "q", Arg: "texte", Help: msg.FlagPlantoGeneraQ}, {Name: "genus-id", Arg: "id", Help: msg.FlagPlantoSpeciesGenusId}},
			Create:  []cli.Flag{{Name: "genus-id", Arg: "id", Help: msg.FlagPlantoSpeciesGenusIdCreate}, {Name: "name", Arg: "nom", Help: msg.FlagPlantoSpeciesName}},
			Update:  []cli.Flag{{Name: "name", Arg: "nom", Help: msg.FlagPeopleLastName}, {Name: "description", Arg: "texte", Help: msg.FlagDescription}}}),
		buildRestCommand(restSpec{Cmd: "varieties", Base: "/plant/varieties", Summary: msg.HelpPlantoVarieties,
			List: true, Show: true,
			Filters: []cli.Flag{{Name: "q", Arg: "texte", Help: msg.FlagPlantoGeneraQ}, {Name: "species-id", Arg: "id", Help: msg.FlagPlantoVarietiesSpeciesId}},
			Create:  []cli.Flag{{Name: "species-id", Arg: "id", Help: msg.FlagPlantoVarietiesSpeciesIdCreate}, {Name: "name", Arg: "nom", Help: msg.FlagPlantoVarietiesName}},
			Update:  []cli.Flag{{Name: "name", Arg: "nom", Help: msg.FlagPeopleLastName}, {Name: "description", Arg: "texte", Help: msg.FlagDescription}}}),
	))

	// ── Administratio (ISC-411) ──
	transactions := buildRestCommand(restSpec{
		Cmd: "transactions", Base: "/administratio/transactions", Summary: msg.HelpTransactions,
		List: true, Show: true, Scope: "administratio",
	})
	transactions.Sub = append(transactions.Sub,
		gesture("reconcile", msg.HelpReconcile, "POST", "/administratio/transactions/%s/reconcile",
			[]cli.Flag{{Name: "expense-id", Arg: "id", Help: msg.FlagReconcileExpenseId}, {Name: "revenue-id", Arg: "id", Help: msg.FlagReconcileRevenueId}}),
		gesture("unreconcile", msg.HelpUnreconcile, "DELETE", "/administratio/transactions/%s/reconcile", nil),
		gesture("ignore", msg.HelpIgnore, "POST", "/administratio/transactions/%s/ignore", nil),
		gesture("restore", msg.HelpTransactionsRestore, "POST", "/administratio/transactions/%s/restore", nil),
	)
	moneyFlags := []cli.Flag{
		{Name: "label", Arg: "libellé", Help: msg.FlagLabel},
		{Name: "amount-cents", Arg: "n", Help: msg.FlagAmountCents},
		{Name: "vat-rate", Arg: "taux", Help: msg.FlagVatRate},
		{Name: "happened-on", Arg: "date", Help: msg.FlagDate},
		{Name: "category-id", Arg: "id", Help: msg.FlagCategoryId},
		{Name: "structure-id", Arg: "id", Help: msg.FlagStructureId},
		{Name: "project-id", Arg: "id", Help: msg.FlagProjectId},
	}
	cli.Register(group("administratio", msg.GroupLentilles,
		msg.HelpAdministratio,
		msg.NotesAdministratio,
		buildRestCommand(restSpec{Cmd: "expenses", Base: "/administratio/expenses", Summary: msg.HelpAdministratioExpenses,
			List: true, Show: true, Create: moneyFlags, Update: moneyFlags, Scope: "administratio"}),
		buildRestCommand(restSpec{Cmd: "revenues", Base: "/administratio/revenues", Summary: msg.HelpAdministratioRevenues,
			List: true, Show: true, Create: moneyFlags, Update: moneyFlags, Scope: "administratio"}),
		transactions,
		buildRestCommand(restSpec{Cmd: "mileage-claims", Base: "/administratio/mileage_claims", Key: "mileage_claims",
			Summary: msg.HelpAdministratioMileageClaims, List: true, Show: true, Scope: "administratio"}),
		buildRestCommand(restSpec{Cmd: "structures", Base: "/administratio/structures", Summary: msg.HelpAdministratioStructures,
			List: true, Scope: "administratio"}),
	))

	// ── Nurserio (ISC-417) ──
	cli.Register(group("nurserio", msg.GroupLentilles,
		msg.HelpNurserio,
		msg.NotesNurserio,
		buildRestCommand(restSpec{Cmd: "stock-batches", Base: "/nurserio/stock_batches", Key: "stock_batches",
			Summary: msg.HelpNurserioStockBatches, List: true, Show: true,
			Create: []cli.Flag{
				{Name: "variety-id", Arg: "id", Help: msg.FlagNurserioStockBatchesVarietyId},
				{Name: "site-id", Arg: "id", Help: msg.FlagNurserioStockBatchesSiteId},
				{Name: "container-id", Arg: "id", Help: msg.FlagNurserioStockBatchesContainerId},
				{Name: "quantity", Arg: "n", Help: msg.FlagQuantite},
				{Name: "status", Arg: "statut", Help: msg.FlagNurserioStockBatchesStatus},
			},
			Update: []cli.Flag{{Name: "quantity", Arg: "n", Help: msg.FlagQuantite}, {Name: "status", Arg: "statut", Help: msg.FlagNurserioStockBatchesStatusEdit}},
			Scope:  "nurserio"}),
		buildRestCommand(restSpec{Cmd: "sites", Base: "/nurserio/sites", Summary: msg.HelpNurserioSites, List: true, Scope: "nurserio"}),
		buildRestCommand(restSpec{Cmd: "containers", Base: "/nurserio/containers", Summary: msg.HelpNurserioContainers, List: true, Scope: "nurserio"}),
	))

	// ── Network (ISC-419) — superadmin ──
	networkHubs := buildRestCommand(restSpec{
		Cmd: "hubs", Base: "/hubs", Summary: msg.HelpNetworkHubs,
		List: true, Show: true, Delete: true, Scope: "network", UpdateMethod: "PUT",
		Create: []cli.Flag{
			{Name: "name", Arg: "nom", Help: msg.FlagNetworkHubsName},
			{Name: "subdomain", Arg: "clé", Help: msg.FlagNetworkHubsSubdomain},
		},
		Update: []cli.Flag{
			{Name: "name", Arg: "nom", Help: msg.FlagNetworkHubsName},
			{Name: "default-locale", Arg: "locale", Help: msg.FlagNetworkHubsDefaultLocale},
		},
	})
	cli.Register(group("network", msg.GroupLentilles,
		msg.HelpNetwork,
		msg.NotesNetwork,
		networkHubs,
	))
}

// gesture fabrique une action POST/DELETE à chemin paramétré par l'id.
func gesture(name, summary, method, pathFmt string, flags []cli.Flag) *cli.Command {
	return &cli.Command{
		Name: name, Summary: summary, ArgSpec: "<id>", MinArgs: 1, Flags: flags,
		APIOps: []string{method + " " + fmt.Sprintf(pathFmt, "{id}")},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			body, rest := flagsToBody(flags, args)
			if len(rest) == 0 {
				return nil, cli.Usagef(msg.UsageMissingID)
			}
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			var payload any
			if len(body) > 0 {
				payload = body
			}
			var out map[string]any
			if err := client.Do(method, fmt.Sprintf(pathFmt, rest[0]), payload, &out); err != nil {
				return nil, err
			}
			return &cli.Result{Data: out, Summary: summary + " ✓"}, nil
		},
	}
}
