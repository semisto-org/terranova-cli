# Couverture de l'API — le gate ISC-390

Chaque opération de `/api/v1/openapi` a une commande CLI **ou** une ligne d'exemption nommée ici ; le test `TestEveryAPIOperationIsCoveredOrExempted` échoue sinon. Le fixture du spec vit dans `internal/commands/testdata/openapi.json` — le rafraîchir : `curl -s https://app.semisto.org/api/v1/openapi > internal/commands/testdata/openapi.json`.

## Exemptions

| Opération | Motif |
|---|---|

*(aucune — la surface couvre les 53 chemins du spec du 2026-08-18, gestes de la spine compris)*

## Le gate inverse (ISC-391) — ce que l'app doit encore faire naître

Une commande listée par l'ISA sans endpoint est un défaut de l'APP. Relevé au 2026-08-18 — **la nuit a fait naître** : /me, gestes de la spine documentés, chat lines, search, notifications, people, my, nurserio orders, payments (+plans, +outstanding), academio (activités/présences/packs/référentiels), asbl (cotisations/AG/effectif), pings. Reste :

| Manque côté app | ISC | Commande CLI en attente |
|---|---|---|
| Rapports (Lineup, Hilltop, financier…) | ISC-410 | `reports` — mes timesheets/bookmarks sont NÉS (/my/*) |
| Analyse de palette, fusion de contacts | ISC-441 | `conceptio palette-analysis`, `contacto merge` |
| Épingler/catégoriser un message, versions de docs, iCal | ISC-399/400/402 | sucre à venir |
