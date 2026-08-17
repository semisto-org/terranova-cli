---
name: terranova
description: Pilote Terranova (app.semisto.org) depuis le shell — projets, tâches, messages, fichiers, contacts, catalogue botanique, comptabilité, pépinière — via le binaire `terranova`. USE WHEN interagir avec Terranova, créer une tâche Semisto, lire un projet, inscrire un contact, consulter le catalogue de plantes, saisir une dépense.
---

# terranova — piloter Terranova depuis le shell

Toujours utiliser `--agent` : chaque commande rend alors du JSON nu, sans séquence ANSI, et l'aide devient structurée (`terranova <cmd> --help --agent`). Le catalogue complet : `terranova commands --agent`.

## S'orienter

- `terranova me --agent` — qui je suis, mes hubs, mes grants, les scopes effectifs du jeton.
- `terranova projects list --agent` — les projets visibles (exactement ceux de l'accueil web).
- `terranova url <url> --agent` — résout une URL de l'app en objet + commandes suggérées.
- `terranova doctor --agent` — si quelque chose ne répond pas, commencer ici.

## La spine (tout est un Recording)

Les 36 types (todos, messages, docs, cards, events, contacts, palettes…) partagent les mêmes gestes :

- `terranova todos list -p <projet> --agent` · `terranova todos show <id> --agent`
- `terranova todos add "Titre" -p <projet> --parent <todolist> --due-on 2026-09-01 --agent`
- `terranova todos edit <id> --completed true --agent` — la complétion passe par le modèle, jamais un booléen nu.
- Gestes hérités : `comment <id> <corps>` · `boost <id>` · `subscribe <id>` · `read <id>` · `archive <id>` · `restore <id>` · `move <id> --to <projet>` · `copy <id> --to <projet>`.
- `terranova todos assign <id> <user_id…>` — REMPLACE la liste des assignés (idempotent).
- `terranova cards move-column <id> --to <colonne>` — la complétion d'une carte EST sa colonne.
- `terranova recordings list --type <Type> --agent` — la spine brute, tous types.

## Les lentilles

- `terranova planto species list --q "Asimina" --agent` — catalogue botanique (lecture ouverte).
- `terranova administratio expenses list --agent` — comptabilité (exige le scope administratio).
- `terranova nurserio stock-batches list --agent` — pépinière.
- `terranova people list --agent` · `terranova grants list --agent` — annuaire et droits.

## Pièges

- Écritures : jamais de retry automatique — si une création échoue en réseau, VÉRIFIER avant de rejouer (`terranova todos list`), sinon doublon.
- Gestes destructifs (`trash`, `remove`) : `--yes` obligatoire en mode agent, aucun prompt.
- `--hub <id>` change de hub pour UNE commande ; `terranova hubs use <id>` le persiste.
- 403 = scope manquant (le jeton n'a jamais plus que son porteur) ; 404 = invisible pour TOI (l'isolation ne confirme pas l'existence).
- Pas de commande pour un besoin ? `terranova api get <chemin> --agent` existe, mais le signaler : il manque une commande.
