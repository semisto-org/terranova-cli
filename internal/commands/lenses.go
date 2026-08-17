package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
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
		cmd.AgentHelp = "Scope requis : " + s.Scope + "."
	}
	if s.List {
		cmd.APIOps = append(cmd.APIOps, "GET "+s.Base)
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "list", Summary: "Liste.", Flags: s.Filters,
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
				return &cli.Result{Data: out, Summary: fmt.Sprintf("%d élément(s).", len(items))}, nil
			},
		})
	}
	if s.Show {
		cmd.APIOps = append(cmd.APIOps, "GET "+s.Base+"/{id}")
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "show", Summary: "Détail.", ArgSpec: "<id>", MinArgs: 1,
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
			Name: "add", Summary: "Crée.", Flags: s.Create,
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
				return &cli.Result{Data: out, Summary: "Créé."}, nil
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
			Name: "edit", Summary: "Modifie.", ArgSpec: "<id>", MinArgs: 1, Flags: s.Update,
			APIOps: []string{method + " " + s.Base + "/{id}"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				body, _ := flagsToBody(s.Update, args[1:])
				if len(body) == 0 {
					return nil, cli.Usagef("rien à modifier — vois les drapeaux avec --help")
				}
				client, err := c.API()
				if err != nil {
					return nil, err
				}
				var out map[string]any
				if err := client.Do(method, s.Base+"/"+args[0], body, &out); err != nil {
					return nil, err
				}
				return &cli.Result{Data: out, Summary: "Modifié."}, nil
			},
		})
	}
	if s.Delete {
		cmd.APIOps = append(cmd.APIOps, "DELETE "+s.Base+"/{id}")
		cmd.Sub = append(cmd.Sub, &cli.Command{
			Name: "remove", Summary: "Supprime (confirmation, ou --yes).", ArgSpec: "<id>", MinArgs: 1,
			APIOps: []string{"DELETE " + s.Base + "/{id}"},
			Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
				if err := confirm(c, fmt.Sprintf("Supprimer %s/%s ?", s.Base, args[0])); err != nil {
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
				return &cli.Result{Data: out, Summary: "Supprimé."}, nil
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
		Cmd: "projects", Base: "/projects", Summary: "Projets du hub : la visibilité est celle de l'accueil web (ISC-374).",
		List: true, Show: true,
		Update: []cli.Flag{
			{Name: "name", Arg: "nom", Help: "Nom du projet."},
			{Name: "description", Arg: "texte", Help: "Description."},
			{Name: "color", Arg: "couleur", Help: "Couleur."},
			{Name: "status", Arg: "statut", Help: "Statut (valeurs archivables seulement)."},
		},
		Create: []cli.Flag{
			{Name: "name", Arg: "nom", Help: "Nom du projet (obligatoire)."},
			{Name: "kind", Arg: "genre", Help: "circle|design|hq… (défaut circle)."},
			{Name: "description", Arg: "texte", Help: "Description."},
			{Name: "access", Arg: "accès", Help: "invite_only|all_access."},
		},
	})
	projects.Group = "Core"
	projects.Sub = append(projects.Sub, &cli.Command{
		Name: "tools", Summary: "Le dock : installer un outil dans un projet (ISC-396).",
		ArgSpec: "<project_id>", MinArgs: 1,
		Flags: []cli.Flag{
			{Name: "install", Arg: "kind", Help: "Installe : message_board|todoset|vault|campfire|schedule|questionnaire|card_table|inbox…"},
			{Name: "name", Arg: "nom", Help: "Nom de l'outil installé."},
		},
		APIOps: []string{"POST /projects/{project_id}/tools"},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			kind, rest := cli.FlagValue(args, "install")
			name, rest := cli.FlagValue(rest, "name")
			if kind == "" {
				return nil, cli.Usagef("usage : terranova projects tools <project_id> --install <kind> [--name <nom>]")
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
			return &cli.Result{Data: out, Summary: "Outil installé."}, nil
		},
	})
	projects.Sub = append(projects.Sub, &cli.Command{
		Name: "people", Summary: "Qui participe au projet : lister, ajouter (--add), retirer (--remove) — membres du hub seulement (ISC-395).",
		ArgSpec: "<project_id>", MinArgs: 1,
		Flags: []cli.Flag{
			{Name: "add", Arg: "user_id", Help: "Ajoute ce membre du hub au projet."},
			{Name: "remove", Arg: "user_id", Help: "Retire cette personne du projet."},
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
				return &cli.Result{Data: out, Summary: "Ajouté au projet."}, nil
			}
			if remove != "" {
				if err := client.Delete("/projects/"+pid+"/people/"+remove, nil); err != nil {
					return nil, err
				}
				return &cli.Result{Data: map[string]any{"removed": remove}, Summary: "Retiré du projet."}, nil
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
			return &cli.Result{Data: out, Headers: []string{"ID", "NOM", "EMAIL"}, Rows: rows,
				Summary: fmt.Sprintf("%d participant(s) · accès %s.", len(out.People), out.Access)}, nil
		},
	})
	cli.Register(projects)

	// ── People & organisation (ISC-409/420) ──
	people := buildRestCommand(restSpec{
		Cmd: "people", Base: "/members", Key: "members", Summary: "L'annuaire du hub (membres).",
		List: true, Show: true,
		Create: []cli.Flag{
			{Name: "email", Arg: "email", Help: "Email de la personne."},
			{Name: "first-name", Arg: "prénom", Help: "Prénom."},
			{Name: "last-name", Arg: "nom", Help: "Nom."},
		},
	})
	people.Group = "Organisation"
	cli.Register(people)

	enroll := buildRestCommand(restSpec{
		Cmd: "enrollments", Base: "/member_enrollments", Summary: "Adhésions membre ↔ hub.",
		List: true, Delete: true,
		Create: []cli.Flag{
			{Name: "member-id", Arg: "id", Help: "Le membre."},
			{Name: "role", Arg: "rôle", Help: "Le rôle."},
		},
	})
	enroll.Group = "Organisation"
	cli.Register(enroll)

	grants := buildRestCommand(restSpec{
		Cmd: "grants", Base: "/interface_grants", Summary: "Grants d'interface (qui voit quelle lentille).",
		List: true, Delete: true, Scope: "projecto (lecture), admin (écriture)",
		Create: []cli.Flag{
			{Name: "user-id", Arg: "id", Help: "La personne."},
			{Name: "interface", Arg: "clé", Help: "administratio|contacto|conceptio|academio|in_situ|nurserio|planto."},
		},
	})
	grants.Group = "Organisation"
	cli.Register(grants)

	// ── Planto (ISC-414) — lecture ouverte, écriture derrière le grant ──
	cli.Register(group("planto", "Lentilles",
		"Catalogue botanique global : genres, espèces, variétés (lecture ouverte à tout membre).",
		"Écriture derrière le grant global planto (READ_OPEN_SHELLS).",
		buildRestCommand(restSpec{Cmd: "genera", Base: "/plant/genera", Summary: "Genres botaniques.",
			List: true, Show: true, Filters: []cli.Flag{{Name: "q", Arg: "texte", Help: "Recherche."}},
			Create: []cli.Flag{{Name: "name", Arg: "nom", Help: "Nom latin du genre."}},
			Update: []cli.Flag{{Name: "name", Arg: "nom", Help: "Nom latin."}, {Name: "description", Arg: "texte", Help: "Description."}}}),
		buildRestCommand(restSpec{Cmd: "species", Base: "/plant/species", Summary: "Espèces.",
			List: true, Show: true,
			Filters: []cli.Flag{{Name: "q", Arg: "texte", Help: "Recherche."}, {Name: "genus-id", Arg: "id", Help: "Filtre par genre."}},
			Create:  []cli.Flag{{Name: "genus-id", Arg: "id", Help: "Le genre."}, {Name: "name", Arg: "nom", Help: "Épithète/nom latin."}},
			Update:  []cli.Flag{{Name: "name", Arg: "nom", Help: "Nom."}, {Name: "description", Arg: "texte", Help: "Description."}}}),
		buildRestCommand(restSpec{Cmd: "varieties", Base: "/plant/varieties", Summary: "Variétés/cultivars.",
			List: true, Show: true,
			Filters: []cli.Flag{{Name: "q", Arg: "texte", Help: "Recherche."}, {Name: "species-id", Arg: "id", Help: "Filtre par espèce."}},
			Create:  []cli.Flag{{Name: "species-id", Arg: "id", Help: "L'espèce."}, {Name: "name", Arg: "nom", Help: "Nom de la variété."}},
			Update:  []cli.Flag{{Name: "name", Arg: "nom", Help: "Nom."}, {Name: "description", Arg: "texte", Help: "Description."}}}),
	))

	// ── Administratio (ISC-411) ──
	transactions := buildRestCommand(restSpec{
		Cmd: "transactions", Base: "/administratio/transactions", Summary: "Grand livre : transactions bancaires.",
		List: true, Show: true, Scope: "administratio",
	})
	transactions.Sub = append(transactions.Sub,
		gesture("reconcile", "Rapproche la transaction d'une pièce.", "POST", "/administratio/transactions/%s/reconcile",
			[]cli.Flag{{Name: "expense-id", Arg: "id", Help: "La dépense rapprochée."}, {Name: "revenue-id", Arg: "id", Help: "La recette rapprochée."}}),
		gesture("unreconcile", "Défait le rapprochement.", "DELETE", "/administratio/transactions/%s/reconcile", nil),
		gesture("ignore", "Écarte la transaction du rapprochement.", "POST", "/administratio/transactions/%s/ignore", nil),
		gesture("restore", "Remet la transaction écartée.", "POST", "/administratio/transactions/%s/restore", nil),
	)
	moneyFlags := []cli.Flag{
		{Name: "label", Arg: "libellé", Help: "Libellé."},
		{Name: "amount-cents", Arg: "n", Help: "Montant TTC en centimes."},
		{Name: "vat-rate", Arg: "taux", Help: "Taux de TVA."},
		{Name: "happened-on", Arg: "date", Help: "Date."},
		{Name: "category-id", Arg: "id", Help: "Catégorie."},
		{Name: "structure-id", Arg: "id", Help: "Structure juridique."},
		{Name: "project-id", Arg: "id", Help: "Projet (ventilation simple)."},
	}
	cli.Register(group("administratio", "Lentilles",
		"Comptabilité : dépenses, recettes, grand livre et rapprochement, structures, notes kilométriques.",
		"Scope requis : administratio (grant + jeton).",
		buildRestCommand(restSpec{Cmd: "expenses", Base: "/administratio/expenses", Summary: "Dépenses (HT/TVA/TTC calculés).",
			List: true, Show: true, Create: moneyFlags, Update: moneyFlags, Scope: "administratio"}),
		buildRestCommand(restSpec{Cmd: "revenues", Base: "/administratio/revenues", Summary: "Recettes.",
			List: true, Show: true, Create: moneyFlags, Update: moneyFlags, Scope: "administratio"}),
		transactions,
		buildRestCommand(restSpec{Cmd: "mileage-claims", Base: "/administratio/mileage_claims", Key: "mileage_claims",
			Summary: "Notes de frais kilométriques.", List: true, Show: true, Scope: "administratio"}),
		buildRestCommand(restSpec{Cmd: "structures", Base: "/administratio/structures", Summary: "Structures juridiques du hub.",
			List: true, Scope: "administratio"}),
	))

	// ── Nurserio (ISC-417) ──
	cli.Register(group("nurserio", "Lentilles",
		"Pépinière : sites, contenants, lots de stock.",
		"Scope requis : nurserio.",
		buildRestCommand(restSpec{Cmd: "stock-batches", Base: "/nurserio/stock_batches", Key: "stock_batches",
			Summary: "Lots de stock (machine d'états gardée).", List: true, Show: true,
			Create: []cli.Flag{
				{Name: "variety-id", Arg: "id", Help: "La variété (catalogue Planto)."},
				{Name: "site-id", Arg: "id", Help: "Le site."},
				{Name: "container-id", Arg: "id", Help: "Le contenant."},
				{Name: "quantity", Arg: "n", Help: "Quantité."},
				{Name: "status", Arg: "statut", Help: "Statut du lot."},
			},
			Update: []cli.Flag{{Name: "quantity", Arg: "n", Help: "Quantité."}, {Name: "status", Arg: "statut", Help: "Statut."}},
			Scope:  "nurserio"}),
		buildRestCommand(restSpec{Cmd: "sites", Base: "/nurserio/sites", Summary: "Sites de production.", List: true, Scope: "nurserio"}),
		buildRestCommand(restSpec{Cmd: "containers", Base: "/nurserio/containers", Summary: "Contenants.", List: true, Scope: "nurserio"}),
	))

	// ── Network (ISC-419) — superadmin ──
	networkHubs := buildRestCommand(restSpec{
		Cmd: "hubs", Base: "/hubs", Summary: "Tous les hubs du réseau (superadmin).",
		List: true, Show: true, Delete: true, Scope: "network", UpdateMethod: "PUT",
		Create: []cli.Flag{
			{Name: "name", Arg: "nom", Help: "Nom du hub."},
			{Name: "subdomain", Arg: "clé", Help: "Sous-domaine."},
		},
		Update: []cli.Flag{
			{Name: "name", Arg: "nom", Help: "Nom du hub."},
			{Name: "default-locale", Arg: "locale", Help: "Locale par défaut."},
		},
	})
	cli.Register(group("network", "Lentilles",
		"Réseau (superadmin) : hubs de la constellation. L'usurpation d'identité reste hors CLI, à dessein.",
		"Scope requis : network (superadmin). ISC-419 : impersonation exclue — un geste de cette portée laisse une trace d'interface.",
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
				return nil, cli.Usagef("il manque l'identifiant")
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
