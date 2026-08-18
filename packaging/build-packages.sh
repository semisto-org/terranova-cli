#!/bin/bash
# Construit les paquets deb/rpm (amd64 + arm64) via nfpm, sans toucher au
# go.mod du module : nfpm est exécuté par `go run …@latest`.
# Usage : packaging/build-packages.sh [version]
# Version par défaut : dernier tag git, sinon celle de la formule Homebrew
# (internal/commands/core.go porte "dev" hors release — inutilisable ici).
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  VERSION="$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')" || true
fi
if [ -z "$VERSION" ]; then
  VERSION="$(sed -n 's/^ *version "\(.*\)"$/\1/p' packaging/homebrew/terranova.rb | head -1)"
fi
if [ -z "$VERSION" ]; then
  echo "version introuvable — passe-la en argument : packaging/build-packages.sh 0.4.0" >&2
  exit 1
fi
echo "── paquets terranova ${VERSION} ──"

for arch in amd64 arm64; do
  bin="dist/terranova_linux_${arch}"
  if [ ! -f "$bin" ]; then
    echo "── cross-compilation ${bin} ──"
    GOOS=linux GOARCH="$arch" go build -ldflags "-s -w -X main.Version=${VERSION}" -o "$bin" ./cmd/terranova
  fi
done

mkdir -p dist/nfpm
trap 'rm -rf dist/nfpm' EXIT
for arch in amd64 arm64; do
  cp "dist/terranova_linux_${arch}" dist/nfpm/terranova
  for fmt in deb rpm; do
    NFPM_VERSION="$VERSION" NFPM_ARCH="$arch" \
      go run github.com/goreleaser/nfpm/v2/cmd/nfpm@latest package \
      -f packaging/nfpm.yaml -p "$fmt" -t dist/
  done
done

echo
ls -lh dist/*.deb dist/*.rpm
