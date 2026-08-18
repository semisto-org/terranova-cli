package msg

// Le socle — auth, hubs, catalogue, doctor, complétion, url, quick-start,
// upgrade, skills embarqués.

const (
	GroupAuthConfig            = "Auth & Config"
	HelpVersion                = "Affiche la version du binaire."
	HelpAuth                   = "Jeton d'accès : login, status, logout, token."
	HelpAuthLogin              = "Range un jeton API (trousseau système, repli fichier 0600)."
	FlagAuthLoginToken         = "Le jeton — sinon demandé au clavier (saisie masquée)."
	FlagAuthLoginHub           = "Hub par défaut du profil."
	NotesAuthLogin             = "Émettre le jeton dans l'app : Compte & réglages → Jetons CLI (/account/api_tokens). En mode --agent, --token est obligatoire (pas de prompt)."
	HelpAuthStatus             = "Dit qui l'on est, sur quel hub, avec quels droits."
	HelpAuthLogout             = "Purge le jeton du profil courant."
	HelpAuthToken              = "Imprime le jeton pour les scripts (jamais dans les logs)."
	HelpMe                     = "Identité, hubs et rôles, grants, scopes effectifs du jeton (« qu'est-ce que j'ai le droit de faire ici »)."
	HelpHubs                   = "La constellation : lister les hubs, choisir le hub par défaut."
	HelpHubsList               = "Liste MES hubs — la constellation telle que je la vois (ISC-424)."
	NotesHubsList              = "La liste vient de /me : un membre voit ses hubs, pas ceux du réseau. La vue superadmin cross-hubs vit sous `network hubs`."
	CrumbChoisirLeHubParDefaut = "choisir le hub par défaut"
	HelpHubsUse                = "Persiste le hub par défaut du profil (ISC-375)."
	GroupAdditional            = "Additional"
	HelpApi                    = "Échappatoire brute : GET/POST/PATCH/PUT/DELETE sur un chemin de l'API (ISC-393)."
	NotesApi                   = "Le corps JSON se passe en 3e argument ou sur stdin. Si une tâche exige cette commande, il manque une vraie commande — le signaler."
	HelpCommands               = "Le catalogue complet de la surface (--json pour les agents)."
	CrumbVoirSesDroits         = "voir ses droits"
	CrumbListerLesHubs         = "lister les hubs"
	CrumbListerLesProjets      = "lister les projets"
	CrumbChangerDeHub          = "changer de hub"
	HelpSurface                = "Imprime le snapshot de surface (.surface) — chaque commande, argument, drapeau (ISC-392)."
	HelpDoctor                 = "Santé : binaire, config, jeton, connexion, identité (ISC-428)."
	HelpCompletion             = "Complétion shell : bash, zsh ou fish (ISC-427)."
	HelpUrl                    = "Résout une URL Terranova en objet + commandes suggérées (ISC-394)."
	HelpQuickStart             = "Le chemin de la première commande utile, sans documentation externe (ISC-429)."
	CrumbVoirLeProjet          = "voir le projet"
	CrumbSesTaches             = "ses tâches"
	CrumbSesRecordings         = "ses recordings"
	CrumbVoir                  = "voir"
	CrumbLaLentilleAuCLI       = "la lentille au CLI"
	HelpUpgrade                = "Met le binaire à jour en place depuis la dernière release (ISC-425)."
	HelpSkill                  = "Skills agent embarqués : lister, afficher, installer (ISC-386)."
	HelpSkillList              = "Les skills embarqués dans ce binaire."
	CrumbInstaller             = "installer"
	HelpSkillShow              = "Affiche un skill."
	HelpSkillInstall           = "Pose les skills dans ~/.claude/skills/ (ou --dir <chemin>)."
	FlagSkillInstallDir        = "Destination (défaut ~/.claude/skills)."
	HelpSetup                  = "Installe le plugin agent : claude ou codex (ISC-387)."
	CompletionHeader           = "# Complétion terranova — générée par `terranova completion %s`\n"
)

// Formats — auth, hubs, doctor, url, upgrade, skills, quick-start.
const (
	ResTokenPurged     = "Jeton du profil %q purgé."
	ResHubCount        = "%d hub(s)."
	ResHubDefault      = "Hub %s par défaut pour le profil %q."
	ResCommandCount    = "%d commandes."
	ResLoggedIn        = "Connecté : %s (profil %q)."
	ResMeSummary       = "%s · hub %s"
	ResURLRecognized   = "Reconnu : %s."
	ResAlreadyCurrent  = "Déjà à jour (%s)."
	ResUpgraded        = "Mis à jour : %s → %s."
	ResSkillCount      = "%d skill(s) embarqués."
	ResSkillsInstalled = "%d skill(s) posés dans %s."

	DoctorBinary       = "binaire"
	DoctorNoteVersion  = "version %s"
	DoctorConfig       = "config"
	DoctorTokenProfile = "jeton (profil %s)"
	DoctorConnection   = "connexion"
	DoctorConnectionTo = "connexion %s"
	DoctorIdentity     = "identité"
	DoctorPlugin       = "plugin agent"
	DoctorPluginNote   = "posable via `terranova setup claude`"
	DoctorAllOK        = "Tout est en ordre."
	DoctorFailing      = "Au moins un contrôle échoue — vois le détail."
)

// QuickStartSteps — le chemin de la première commande utile (ISC-429).
var QuickStartSteps = []string{
	"1. Émets ton jeton dans l'app : Compte & réglages → Jetons CLI (https://app.semisto.org/account/api_tokens).",
	"2. `terranova auth login` — colle le jeton (il part au trousseau, jamais dans un fichier du dépôt).",
	"3. `terranova me` — vérifie qui tu es et ce que tu peux faire.",
	"4. `terranova projects list` — tes projets ; `terranova todos list -p <id>` — les tâches d'un projet.",
	"5. `terranova todos add \"Ma tâche\" -p <id> --due-on 2026-09-01` — ta première écriture.",
	"6. `terranova commands` — toute la surface ; `--agent` sur n'importe quoi pour le JSON.",
}
