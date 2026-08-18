package msg

// Erreurs, messages d'usage, confirmations et invites — toutes surfaces.
const (
	// cli / main — dispatch et drapeaux globaux
	UsageFlagNeedsValue  = "le drapeau %s attend une valeur"
	ErrUnknownCommand    = "commande inconnue : %s — `terranova --help` liste la surface"
	ErrUnknownSubcommand = "sous-commande inconnue : %s — `terranova %s --help`"
	UsageCommand         = "usage : terranova %s %s"

	// spine — gestes destructifs et création
	UsageDestructiveNeedsYes = "geste destructif : ajoute --yes pour confirmer (aucun prompt en mode agent)"
	AskOuiNon                = "%s (oui/non) "
	ErrCancelled             = "annulé"
	AskTrashRecording        = "Mettre le recording %s à la corbeille ?"
	UsageNothingToEditSpine  = "rien à modifier — passe --title, --body, --due-on ou --completed"
	UsageProjectRequired     = "--project est obligatoire pour créer un %s"
	UsageFileRequired        = "--file est obligatoire pour un upload"
	UsageMoveCopy            = "usage : … %s <id> --to <project_id>"
	UsageMoveColumn          = "usage : … move-column <id> --to <column_recording_id>"

	// lenses
	UsageNothingToEditRest = "rien à modifier — vois les drapeaux avec --help"
	AskDeleteResource      = "Supprimer %s/%s ?"
	UsageProjectsTools     = "usage : terranova projects tools <project_id> --install <kind> [--name <nom>]"
	UsageMissingID         = "il manque l'identifiant"

	// auth / api brute
	UsageAgentTokenRequired = "en mode --agent, --token est obligatoire (aucun prompt)"
	ErrNoToken              = "aucun jeton fourni"
	ErrTokenRefused         = "jeton refusé par %s : %w"
	PromptAPIToken          = "Jeton API (Compte & réglages → Jetons CLI) : "
	UsageInvalidJSONBody    = "corps JSON invalide : %v"

	// comms
	UsagePingsSend = "usage : terranova pings send <texte> --to <user_id,…>"
	AskArchivePing = "Archiver cette conversation ? (irréversible)"

	// academio / payments
	UsagePaymentsAdd = "--participant et --amount-cents sont obligatoires"

	// skills / setup / completion / url / upgrade
	UsageUnknownSkill       = "skill inconnu : %s (vois `terranova skill list`)"
	UsageUnknownSetupTarget = "cible inconnue : %s (claude|codex)"
	UsageUnknownShell       = "shell inconnu : %s (bash|zsh|fish)"
	UsageInvalidURL         = "URL invalide : %v"
	ErrURLNotRecognized     = "URL non reconnue — motifs connus : /projects/<id>, /<type>/<id>, /<lentille>"
	ErrNoReleasePublished   = "aucune release publiée"
	ErrNoBinaryInRelease    = "pas de binaire %s dans la release %s"

	// chat / nurserio
	AskDeleteChatLine      = "Supprimer cette ligne de chat ?"
	UsageOrderLineFormat   = "--line attend <stock_batch_id>:<quantité>, reçu %q"
	UsageOrderLineRequired = "au moins une --line <stock_batch_id>:<quantité> est requise"

	// config (profils, jeton)
	ErrConfigUnreadable  = "config.json illisible : %w"
	ErrNoTokenForProfile = "aucun jeton pour le profil %q — lance `terranova auth login`"

	// api — enveloppe d'erreur HTTP et traces -vv
	ErrHTTPWithCode = "HTTP %d — %s"
	ErrHTTP         = "HTTP %d"
	VerboseRequest  = "→ %s %s\n"
	VerboseResponse = "← %d (%d octets)\n"

	// output
	UsageInvalidJQ = "--jq : filtre invalide : %w"
)
