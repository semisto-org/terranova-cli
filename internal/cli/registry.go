// Package cli — le registre déclaratif des commandes. Tout en découle :
// le dispatch, l'aide (humaine et --agent), le catalogue `commands --json`,
// le snapshot `.surface`, la complétion shell et le gate de couverture d'API.
// Une commande qui n'est pas ici n'existe pas (ISC-385/392).
package cli

// Flag décrit un drapeau propre à une commande. Arg vide = booléen.
type Flag struct {
	Name  string `json:"name"`
	Short string `json:"short,omitempty"`
	Arg   string `json:"arg,omitempty"`
	Help  string `json:"help"`
}

// Crumb est un fil d'Ariane EXÉCUTABLE : cmd doit exister dans la surface
// réelle du binaire (ISC-384, vérifié par test de rejeu).
type Crumb struct {
	Action string `json:"action"`
	Cmd    string `json:"cmd"`
}

// Result est ce qu'une commande rend : l'enveloppe --json les rend tels quels,
// le mode stylé passe par Rows quand il est fourni.
type Result struct {
	Data    any     `json:"data"`
	Summary string  `json:"summary,omitempty"`
	Crumbs  []Crumb `json:"breadcrumbs,omitempty"`
	// Rows : rendu tableau optionnel pour le terminal (en-têtes + lignes).
	Headers []string   `json:"-"`
	Rows    [][]string `json:"-"`
}

// Command est un nœud de la surface. Name/flags en anglais (règle des routes),
// Summary/Help en français (messages humains, ISC-430).
type Command struct {
	Name      string     `json:"name"`
	Group     string     `json:"group,omitempty"`
	Summary   string     `json:"summary"`
	ArgSpec   string     `json:"args,omitempty"`
	MinArgs   int        `json:"-"`
	Flags     []Flag     `json:"flags,omitempty"`
	Sub       []*Command `json:"subcommands,omitempty"`
	AgentHelp string     `json:"agent_notes,omitempty"`
	// APIOps : les opérations OpenAPI ("GET /projects", …) que cette commande
	// couvre — la matière du gate de couverture (ISC-390).
	APIOps []string                                     `json:"api_ops,omitempty"`
	Run    func(c *Ctx, args []string) (*Result, error) `json:"-"`
}

// Registry est la racine de la surface.
var Registry []*Command

// Register ajoute une commande de premier niveau.
func Register(cmd *Command) { Registry = append(Registry, cmd) }

// Find résout un chemin de commande (["todos","add"]) dans le registre et rend
// la commande + le reste des arguments.
func Find(args []string) (*Command, []string) {
	var cur *Command
	list := Registry
	rest := args
	for len(rest) > 0 {
		var next *Command
		for _, c := range list {
			if c.Name == rest[0] {
				next = c
				break
			}
		}
		if next == nil {
			break
		}
		cur = next
		rest = rest[1:]
		list = next.Sub
	}
	return cur, rest
}

// Walk parcourt toute la surface, chemin compris ("todos add").
func Walk(fn func(path string, c *Command)) {
	var rec func(prefix string, list []*Command)
	rec = func(prefix string, list []*Command) {
		for _, c := range list {
			p := c.Name
			if prefix != "" {
				p = prefix + " " + c.Name
			}
			fn(p, c)
			rec(p, c.Sub)
		}
	}
	rec("", Registry)
}
