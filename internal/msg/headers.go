package msg

// En-têtes des tableaux du terminal — et les libellés de cellules fixes.
var (
	HeadersRecordingList  = []string{"ID", "TITRE", "STATUT", "PROJET"}
	HeadersRecordingKinds = []string{"ID", "TYPE", "TITRE", "STATUT"}
	HeadersTypedList      = []string{"ID", "TYPE", "TITRE", "PROJET"}
	HeadersHubList        = []string{"ID", "NOM", "RÔLE", "DÉFAUT"}
	HeadersCommandCatalog = []string{"COMMANDE", "RÉSUMÉ"}
	HeadersPeople         = []string{"ID", "NOM", "EMAIL"}
	HeadersDoctor         = []string{"", "CONTRÔLE", "NOTE"}
	HeadersNotifications  = []string{"", "ID", "GENRE", "QUI", "QUOI"}
	HeadersMyTodos        = []string{"ID", "TÂCHE", "ÉCHÉANCE"}
	HeadersMyTimesheets   = []string{"JOUR", "PROJET", "H", "QUOI"}
	HeadersActivities     = []string{"ID", "ACTIVITÉ", "TYPE", "LIEU", "INSCRITS"}
	HeadersAttendances    = []string{"ID", "PARTICIPANT", "SÉANCE", "PRÉSENT"}
	HeadersPayments       = []string{"ID", "ÉTAT", "GENRE", "CENTIMES", "PROJET"}
	HeadersOutstanding    = []string{"PARTICIPANT", "PROJET", "ATTENDU", "PAYÉ", "RESTE"}
	HeadersPings          = []string{"ID", "AVEC"}
	HeadersActivityFeed   = []string{"QUAND", "QUI", "GESTE", "SUR"}
	HeadersJournal        = []string{"QUAND", "QUI", "GESTE"}
	HeadersFees           = []string{"ID", "MEMBRE", "ÉTAT", "CENTIMES", "DÉBUT"}
	HeadersChatLines      = []string{"ID", "QUI", "QUOI"}
	HeadersOrders         = []string{"ID", "N°", "ÉTAT", "CLIENT", "TOTAL (c)"}
)

// LabelProjet — la cellule « type » des projets dans les résultats de recherche.
const LabelProjet = "Projet"
