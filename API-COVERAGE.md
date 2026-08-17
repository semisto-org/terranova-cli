# Couverture de l'API — le gate ISC-390

Chaque opération de `/api/v1/openapi` a une commande CLI **ou** une ligne d'exemption nommée ici ; le test `TestEveryAPIOperationIsCoveredOrExempted` échoue sinon. Le fixture du spec vit dans `internal/commands/testdata/openapi.json` — le rafraîchir : `curl -s https://app.semisto.org/api/v1/openapi > internal/commands/testdata/openapi.json`.

## Exemptions

| Opération | Motif |
|---|---|

*(aucune — la surface couvre les 37 chemins du spec du 2026-08-18)*

## Gestes non documentés au spec (écart côté app, pas côté CLI)

Les routes de gestes de la spine (`/recordings/{id}/comments·boost·subscribe·read·archive·restore·assignees·column·move·copy`) sont **livrées et testées côté app** (ISC-440, PR #535→#537) mais le générateur OpenAPI ne les documente pas encore. Le CLI les couvre ; le test `TestDeclaredOpsExistInSpec` porte une tolérance NOMMÉE sur ce préfixe, à retirer quand le générateur les documentera.

## Le gate inverse (ISC-391) — ce que l'app doit encore faire naître

Une commande listée par l'ISA sans endpoint est un défaut de l'APP. Relevé au 2026-08-18 :

| Manque côté app | ISC | Commande CLI en attente |
|---|---|---|
| Lignes de chat (CampfireLine, annexe sans surface) | ISC-401 | `chat post` / `chat lines` |
| Recherche globale | ISC-407 | `search` |
| Notifications + pings | ISC-408 | `notifications`, `pings` |
| Rapports / mes vues | ISC-410 | `my`, `reports` |
| Academio (activités, séances, packs, présences) | ISC-436 | `academio …` |
| Paiements et échéanciers | ISC-437 | `payments …` |
| Adhésion ASBL | ISC-438 | `administratio members …` |
| Commandes de pépinière | ISC-439 | `nurserio orders …` |
| Analyse de palette, fusion de contacts | ISC-441 | `conceptio palette-analysis`, `contacto merge` |
| Épingler/catégoriser un message, versions de docs, iCal | ISC-399/400/402 | sucre à venir |
