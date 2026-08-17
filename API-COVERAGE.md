# Couverture de l'API — le gate ISC-390

Chaque opération de `/api/v1/openapi` a une commande CLI **ou** une ligne d'exemption nommée ici ; le test `TestEveryAPIOperationIsCoveredOrExempted` échoue sinon. Le fixture du spec vit dans `internal/commands/testdata/openapi.json` — le rafraîchir : `curl -s https://app.semisto.org/api/v1/openapi > internal/commands/testdata/openapi.json`.

## Exemptions

| Opération | Motif |
|---|---|

*(aucune — la surface couvre les 53 chemins du spec du 2026-08-18, gestes de la spine compris)*

## Le gate inverse (ISC-391) — ce que l'app doit encore faire naître

Une commande listée par l'ISA sans endpoint est un défaut de l'APP. Relevé au 2026-08-18 :

| Manque côté app | ISC | Commande CLI en attente |
|---|---|---|
| Pings (conversations 1:1\/groupe) | ISC-408 | `pings` |
| Rapports + mes timesheets\/bookmarks | ISC-410 | `reports`, `my timesheets`, `my bookmarks` |
| Academio (activités, séances, packs, présences) | ISC-436 | `academio …` |
| Paiements et échéanciers | ISC-437 | `payments …` |
| Adhésion ASBL | ISC-438 | `administratio members …` |
| Commandes de pépinière | ISC-439 | `nurserio orders …` |
| Analyse de palette, fusion de contacts | ISC-441 | `conceptio palette-analysis`, `contacto merge` |
| Épingler/catégoriser un message, versions de docs, iCal | ISC-399/400/402 | sucre à venir |
