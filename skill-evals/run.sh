#!/bin/zsh
# ISC-389 — le banc d'évaluation : des cas réels joués au binaire réel contre
# l'instance e2e. v1 déterministe : la séquence de commandes qu'un agent
# devrait composer est REJOUÉE et son résultat vérifié — la version « un agent
# joue le prompt et on note » se branchera au même format de cas.
set -uo pipefail
cd "$(dirname "$0")"
T="${TERRANOVA_BIN:-/tmp/terranova-e2e} --agent"
pass=0; fail=0
while IFS=$'\t' read -r name sequence verify expected; do
  [[ "$name" == \#* || -z "$name" ]] && continue
  pid=""; out=""
  ok=1
  IFS=';' read -rA steps <<< "$sequence"
  for step in "${steps[@]}"; do
    step="${step## }"; step="${step%% }"
    step="${step//\{pid\}/$pid}"; step="${step//\{out\}/$out}"
    if [[ "$step" == SETPID* ]]; then
      pname="${step#SETPID }"
      pid=$(eval "$T projects list --jq '.projects[] | select(.name==\"$pname\") | .id' --quiet" | head -1)
      [[ -n "$pid" ]] || { ok=0; break }
      continue
    fi
    result=$(eval "$T $step" 2>&1)
    code=$?
    [[ $code -ne 0 ]] && { echo "  ✗ $name — étape « $step » → exit $code"; ok=0; break }
    last=$(echo "$result" | grep -Eo '^[0-9]+$' | head -1)
    [[ -n "$last" ]] && out="$last"
  done
  if [[ $ok -eq 1 ]]; then
    v="${verify//\{pid\}/$pid}"; v="${v//\{out\}/$out}"
    result=$(eval "$T $v" 2>&1)
    if echo "$result" | grep -q "$expected"; then
      echo "  ✓ $name"
      pass=$((pass+1))
      continue
    fi
    echo "  ✗ $name — « $expected » absent de la vérification"
  fi
  fail=$((fail+1))
done < cases.tsv
echo "---"
echo "evals : $pass réussi(s), $fail raté(s)"
[[ $fail -eq 0 ]]
