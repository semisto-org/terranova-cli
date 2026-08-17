#!/bin/zsh
# e2e ISC-431 — le parcours joué contre une VRAIE instance Terranova (env test,
# base de test, serveur local), au binaire réel. Usage : e2e/run-local.sh
set -euo pipefail
cd "$(dirname "$0")/.."

APP_DIR="${TERRANOVA_APP_DIR:-$HOME/code/terranova-2026}"
PORT="${E2E_PORT:-3057}"
export PATH="$HOME/.asdf/shims:/opt/homebrew/bin:$PATH"

echo "── binaire ──"
go build -o /tmp/terranova-e2e ./cmd/terranova

echo "── instance test (base préparée + jeton semé) ──"
cd "$APP_DIR"
# La base de test est PARTAGÉE avec `bin/rails test` : schéma frais au départ,
# et on la REND propre en sortant (sinon les fixtures du prochain `rails test`
# butent sur nos lignes — violation FK vécue le 2026-08-18).
RAILS_ENV=test bin/rails db:schema:load >/dev/null
TOKEN=$(RAILS_ENV=test bin/rails runner '
  hub = Hub.find_or_create_by!(slug: "e2e") { |h| h.name = "E2E" }
  hub.update!(name: "E2E") unless hub.name == "E2E"
  # Une adhésion exige une structure juridique par défaut (patron Semisto ASBL).
  Structure.find_or_create_by!(hub: hub, name: "E2E ASBL") { |s| s.default = true }
  user = User.find_or_create_by!(email_address: "e2e@semisto.org") { |u| u.name = "E2E Runner"; u.password = SecureRandom.hex(12) }
  Membership.find_or_create_by!(user: user, hub: hub) { |m| m.role = "admin" }
  token = ApiToken.create!(user: user, hub: hub, name: "e2e-#{Time.now.to_i}")
  puts token.raw_token
' | tail -1)
[ -n "$TOKEN" ] || { echo "jeton non semé"; exit 1; }

RAILS_ENV=test bin/rails server -p "$PORT" -P /tmp/terranova-e2e-server.pid > /tmp/terranova-e2e-server.log 2>&1 &
SERVER_PID=$!
trap '[ -f /tmp/terranova-e2e-server.pid ] && kill $(cat /tmp/terranova-e2e-server.pid) 2>/dev/null; kill $SERVER_PID 2>/dev/null; (cd "$APP_DIR" && RAILS_ENV=test bin/rails db:schema:load >/dev/null 2>&1)' EXIT

for i in $(seq 1 60); do
  curl -sf "http://127.0.0.1:$PORT/up" >/dev/null 2>&1 && break
  sleep 1
  [ $i -eq 60 ] && { echo "le serveur test ne répond pas"; tail -20 /tmp/terranova-e2e-server.log; exit 1; }
done
echo "instance test prête sur :$PORT"

echo "── parcours bats ──"
cd - >/dev/null
TERRANOVA_BIN=/tmp/terranova-e2e \
TERRANOVA_BASE_URL="http://127.0.0.1:$PORT/api/v1" \
TERRANOVA_API_TOKEN="$TOKEN" \
TERRANOVA_NO_KEYCHAIN=1 \
TERRANOVA_CONFIG_DIR=$(mktemp -d) \
bats e2e/course.bats
