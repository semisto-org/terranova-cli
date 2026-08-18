package msg

// Vague du 2026-08-18 après-midi : rapports, remontées (bubble-ups), silence
// global et fusion de contacts — les 9 chemins d'API nés le même jour côté app.
const (
	// Gestes de spine.
	HelpBubbleUp   = "Remonte l'élément dans Hey! (préréglage optionnel : now, this_evening, tomorrow, next_week)."
	HelpUnbubbleUp = "Retire ta remontée."

	// Notifications.
	HelpNotificationsSilence   = "Shhh… — coupe globalement les notifications calmes (comme l'écran)."
	HelpNotificationsUnsilence = "Rallume les notifications."
	ResSilenceOn               = "Notifications coupées (Shhh…)."
	CrumbRallumer              = "rallumer"
	CrumbAssigner              = "assigner"
	ResSilenceOff              = "Notifications rallumées."

	// Mes remontées.
	HelpMyBubbleUps  = "Mes remontées à venir, par échéance croissante."
	ResBubbleUpCount = "%d remontée(s) à venir."

	// Fusion de contacts.
	HelpContactsMerge = "Fusionne deux fiches : la première SURVIT, la seconde disparaît (irréversible — même hub, même type)."
	AskMergeContacts  = "Fusionner la fiche %s dans la fiche %s (la seconde citée disparaît)"
	ResContactsMerged = "Fiches fusionnées — la fiche %s porte tout."

	// Rapports.
	HelpReports            = "Les rapports de travail — le miroir des écrans /reports."
	HelpReportsUpcoming    = "Ce qui arrive à échéance (aujourd'hui inclus)."
	HelpReportsOverdue     = "En retard, groupé par ampleur du retard."
	HelpReportsUnassigned  = "Le travail sans personne dessus."
	HelpReportsAssignments = "Le travail assigné à quelqu'un (défaut : toi)."
	HelpReportsThroughput  = "Débit : créé vs terminé, par jour."
	FlagWithinDays         = "Horizon en jours (défaut 14, max 365)"
	FlagPerson             = "ID de la personne (défaut : le porteur du jeton)"
	FlagReportDays         = "Fenêtre en jours (7 à 365, défaut 30)"
	FlagReportKind         = "all, todo ou card (défaut all)"
	ResUpcomingCount       = "%d élément(s) à échéance sous %s jours."
	ResOverdueCount        = "%d en retard (today %d · this_week %d · last_week %d · older %d)."
	ResUnassignedCount     = "%d élément(s) sans assigné."
	ResAssignmentsCount    = "%d élément(s) sur les épaules de %s."
	ResThroughputSummary   = "%s jours · +%d créés · %d terminés."
)

// HeadersReportItems : les colonnes communes aux rapports de travail.
var HeadersReportItems = []string{"ID", "GENRE", "TITRE", "ÉCHÉANCE", "PROJET", "ASSIGNÉS"}

// HeadersBubbleUps : mes remontées.
var HeadersBubbleUps = []string{"ID", "RECORDING", "TITRE", "PROJET", "REMONTE LE"}
