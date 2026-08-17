package cli

import (
	"github.com/semisto-org/terranova-cli/internal/api"
	"github.com/semisto-org/terranova-cli/internal/config"
)

// GlobalFlags — les drapeaux valables sur toute commande (patron basecamp-cli).
type GlobalFlags struct {
	JSON    bool   // --json / -j : enveloppe {ok, data, summary, breadcrumbs}
	Quiet   bool   // --quiet / -q : les données nues
	MD      bool   // --md / -m : sortie markdown portable
	Agent   bool   // --agent = --json --quiet + comportements non interactifs
	IDsOnly bool   // --ids-only
	Count   bool   // --count
	JQ      string // --jq <filtre> (gojq intégré)
	Hub     string // --hub <id> → X-Hub-Id
	Project string // --project / -p : projet par défaut de la commande
	Profile string // --profile / -P
	Yes     bool   // --yes : lève les confirmations destructives (ISC-381)
	Verbose int    // -v / -vv (ISC-433)
	Help    bool   // --help / -h
}

// Ctx est le contexte d'exécution passé à chaque commande.
type Ctx struct {
	Flags   GlobalFlags
	Config  *config.Config
	Profile string
	Client  *api.Client
	Version string
	IsTTY   bool
}

// API rend le client HTTP, construit paresseusement (les commandes locales —
// version, commands, completion — n'exigent pas d'authentification).
func (c *Ctx) API() (*api.Client, error) {
	if c.Client != nil {
		return c.Client, nil
	}
	cl, err := api.New(c.Config, c.Profile, c.Flags.Hub, c.Flags.Verbose)
	if err != nil {
		return nil, err
	}
	c.Client = cl
	return cl, nil
}
