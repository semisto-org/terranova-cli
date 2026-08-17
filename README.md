# terranova

Le compagnon en ligne de commande de [Terranova](https://app.semisto.org) — l'app qui fait tourner Semisto.

Ce qu'aucune autre CLI ne fait : piloter la **gestion de projet** ET la **comptabilité** ET le **catalogue botanique** ET la **pépinière** d'une association, depuis le même binaire, avec le même jeton.

```sh
terranova todos add "Planter la haie nord" -p 19 --due-on 2026-11-15
terranova administratio expenses add --label "Paillage" --amount-cents 24000 --vat-rate 21
terranova planto species list --q "Asimina"
terranova nurserio stock-batches list
terranova me   # qui je suis, mes hubs, mes droits — la réponse en un appel
```

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/semisto-org/terranova-cli/main/install.sh | bash
```

Ou : `go install github.com/semisto-org/terranova-cli/cmd/terranova@latest`.

Première fois : `terranova quick-start`. Jeton : dans l'app, **Compte & réglages → Jetons CLI**, puis `terranova auth login`.

## Pour les agents

`--agent` sur toute commande = JSON nu ; `--help --agent` = aide structurée ; `terranova commands --agent` = le catalogue complet (600+ entrées) ; breadcrumbs exécutables dans chaque réponse. `terranova skill install` pose les skills Claude/Codex embarqués.

## Développement

`bin/ci` rejoue la CI complète : gofmt, vet, tests, gates (surface `.surface`, couverture d'API contre le spec OpenAPI, rejeu des breadcrumbs, dérive des skills). La discipline : **chaque opération de l'API a une commande ou une exemption nommée** (`API-COVERAGE.md`) — et une commande sans endpoint fait naître l'endpoint côté app, jamais l'inverse.
