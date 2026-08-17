#!/usr/bin/env bats
# Le parcours réel (ISC-431) : identité → projet → outil → liste → tâche →
# geste de spine → complétion → corbeille — chaque étape au binaire réel,
# contre une instance Terranova de test.

T="$TERRANOVA_BIN --agent"

@test "me : l'identité répond et nomme le hub E2E" {
  run $T me
  [ "$status" -eq 0 ]
  [[ "$output" == *"E2E Runner"* ]]
  [[ "$output" == *'"name": "E2E"'* || "$output" == *'"name":"E2E"'* ]]
}

@test "projects add : un projet naît et se retrouve dans la liste" {
  run $T projects add --name "Chantier e2e" --kind circle
  [ "$status" -eq 0 ]
  run $T projects list
  [ "$status" -eq 0 ]
  [[ "$output" == *"Chantier e2e"* ]]
}

@test "tools : le dock s'installe (todoset) et la spine suit" {
  pid=$($T projects list --jq '.projects[] | select(.name=="Chantier e2e") | .id' --quiet)
  [ -n "$pid" ]
  run $T projects tools "$pid" --install todoset --name "To-dos"
  [ "$status" -eq 0 ]
}

@test "todolist + todo : création, échéance, listing" {
  pid=$($T projects list --jq '.projects[] | select(.name=="Chantier e2e") | .id' --quiet)
  set_id=$($T todosets list -p "$pid" --ids-only | head -1)
  [ -n "$set_id" ]
  run $T todolists add "Plantations" -p "$pid" --parent "$set_id"
  [ "$status" -eq 0 ]
  list_id=$($T todolists list -p "$pid" --ids-only | head -1)
  run $T todos add "Planter la haie e2e" -p "$pid" --parent "$list_id" --due-on 2026-11-15
  [ "$status" -eq 0 ]
  run $T todos list -p "$pid"
  [[ "$output" == *"Planter la haie e2e"* ]]
  [[ "$output" == *"2026-11-15"* ]]
}

@test "spine : commenter puis s'abonner, hérités sans architecture propre" {
  pid=$($T projects list --jq '.projects[] | select(.name=="Chantier e2e") | .id' --quiet)
  todo_id=$($T todos list -p "$pid" --ids-only | head -1)
  run $T todos comment "$todo_id" "<div>posé par le parcours e2e</div>"
  [ "$status" -eq 0 ]
  run $T todos subscribe "$todo_id"
  [ "$status" -eq 0 ]
  run $T todos show "$todo_id"
  [[ "$output" == *'"comments": 1'* || "$output" == *'"comments":1'* ]]
}

@test "complétion : par le modèle, pas un booléen nu" {
  pid=$($T projects list --jq '.projects[] | select(.name=="Chantier e2e") | .id' --quiet)
  todo_id=$($T todos list -p "$pid" --ids-only | head -1)
  run $T todos edit "$todo_id" --completed true
  [ "$status" -eq 0 ]
  run $T todos show "$todo_id"
  [[ "$output" == *'"completed": true'* || "$output" == *'"completed":true'* ]]
}

@test "corbeille : refusée sans --yes en mode agent, passée avec, restaurée ensuite" {
  pid=$($T projects list --jq '.projects[] | select(.name=="Chantier e2e") | .id' --quiet)
  todo_id=$($T todos list -p "$pid" --ids-only | head -1)
  run $T todos trash "$todo_id"
  [ "$status" -eq 2 ]
  run $T todos trash "$todo_id" --yes
  [ "$status" -eq 0 ]
  run $T recordings restore "$todo_id"
  [ "$status" -eq 0 ]
}

@test "codes de sortie : 404 isolé rend 5, jamais une confirmation d'existence" {
  run $T recordings show 99999999
  [ "$status" -eq 5 ]
}

@test "chat : le dock installe un campfire, on poste, on lit, on édite sa ligne" {
  pid=$($T projects list --jq '.projects[] | select(.name=="Chantier e2e") | .id' --quiet)
  run $T projects tools "$pid" --install campfire --name "Feu e2e"
  [ "$status" -eq 0 ]
  fire_id=$($T chat list -p "$pid" --ids-only | head -1)
  [ -n "$fire_id" ]
  run $T chat post "$fire_id" "Bonjour du parcours e2e"
  [ "$status" -eq 0 ]
  run $T chat lines "$fire_id"
  [ "$status" -eq 0 ]
  [[ "$output" == *"Bonjour du parcours e2e"* ]]
  line_id=$($T chat lines "$fire_id" --ids-only | head -1)
  run $T chat edit-line "$fire_id" "$line_id" "Bonjour, édité"
  [ "$status" -eq 0 ]
}

@test "search : la recherche globale trouve la tâche du parcours" {
  run $T search "haie e2e"
  [ "$status" -eq 0 ]
  [[ "$output" == *"Planter la haie e2e"* ]]
}

@test "notifications : la liste répond (vide ou pas), read-all passe" {
  run $T notifications list
  [ "$status" -eq 0 ]
  run $T notifications read-all
  [ "$status" -eq 0 ]
}
