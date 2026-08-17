---
name: terranova-doctor
description: Diagnostique une panne du CLI terranova — authentification, connexion, permissions, profil. USE WHEN terranova renvoie une erreur, un 401/403, un timeout, ou qu'une commande ne fait rien.
---

# terranova-doctor — diagnostiquer une panne

1. `terranova doctor --agent` — binaire, config, jeton, connexion, identité, en un appel.
2. Codes de sortie : 0 succès · 1 exécution · 2 usage · 3 authentification · 4 permission · 5 introuvable.
3. **Exit 3 (401)** : jeton absent/révoqué. `terranova auth login --token <jeton>` — l'émission vit dans l'app, Compte & réglages → Jetons CLI.
4. **Exit 4 (403)** : le scope manque — au jeton OU à son porteur (l'intersection est revérifiée à chaque requête). `terranova me --agent` montre `token.effective_scopes` ; si le scope voulu n'y est pas, le porteur n'a pas le grant : c'est un admin du hub qui l'attribue.
5. **Exit 5 (404)** : l'objet n'existe pas POUR TOI — mauvais hub (`--hub <id>`) ou projet invite_only où tu n'es pas.
6. Réseau : `-vv` trace chaque requête sur stderr (jamais de secret dans les traces). `terranova api get /health --agent` teste la connexion nue.
7. Profils : `TERRANOVA_PROFILE` ou `--profile` — un jeton par identité (humain, nocturne, Nova). `terranova auth status --agent` dit lequel est actif.
