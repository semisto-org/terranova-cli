package msg

// L'aide humaine — la charpente que `cli.HelpText` assemble autour du registre.
const (
	HelpRootBanner        = "terranova — le compagnon en ligne de commande de Terranova (app.semisto.org)\n\n"
	HelpRootUsage         = "Usage : terranova <commande> [sous-commande] [arguments] [drapeaux]\n\n"
	GroupOther            = "Autres"
	HelpGroupHeader       = "%s :\n"
	HelpRootGlobalFlags1  = "Drapeaux globaux : --json/-j --quiet/-q --md/-m --agent --ids-only --count --jq <filtre>\n"
	HelpRootGlobalFlags2  = "                   --hub <id> --project/-p <id> --profile/-P <nom> --yes -v/-vv\n"
	HelpRootFooter        = "\n`terranova <commande> --help` pour le détail · `--help --agent` pour l'aide structurée.\n"
	HelpCommandTitle      = "terranova %s — %s\n"
	HelpCommandUsage      = "Usage : terranova %s %s\n"
	HelpFlagsHeader       = "\nDrapeaux :\n"
	HelpSubcommandsHeader = "\nSous-commandes :\n"
	HelpNotesPrefix       = "\nNotes : "
)
