#!/bin/sh
# Installe le binaire terranova depuis la dernière release GitHub.
# Usage : curl -fsSL https://raw.githubusercontent.com/semisto-org/terranova-cli/main/install.sh | bash
set -e

REPO="semisto-org/terranova-cli"
DEST="${TERRANOVA_INSTALL_DIR:-/usr/local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64) arch="amd64" ;;
  aarch64 | arm64) arch="arm64" ;;
  *) echo "architecture non gérée : $arch" && exit 1 ;;
esac
case "$os" in
  darwin | linux) ;;
  *) echo "OS non géré : $os (macOS et Linux seulement)" && exit 1 ;;
esac

tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
[ -n "$tag" ] || { echo "impossible de lire la dernière release"; exit 1; }

asset="terranova_${os}_${arch}"
url="https://github.com/$REPO/releases/download/$tag/$asset"
tmp=$(mktemp)
echo "Téléchargement de terranova $tag ($os/$arch)…"
curl -fsSL "$url" -o "$tmp"

sumurl="https://github.com/$REPO/releases/download/$tag/checksums.txt"
if curl -fsSL "$sumurl" -o "$tmp.sums" 2>/dev/null; then
  expected=$(grep " $asset\$" "$tmp.sums" | cut -d' ' -f1)
  actual=$(shasum -a 256 "$tmp" | cut -d' ' -f1)
  [ "$expected" = "$actual" ] || { echo "somme de contrôle invalide"; exit 1; }
fi

chmod +x "$tmp"
if [ -w "$DEST" ]; then
  mv "$tmp" "$DEST/terranova"
else
  echo "(sudo requis pour écrire dans $DEST)"
  sudo mv "$tmp" "$DEST/terranova"
fi
echo "terranova $tag installé dans $DEST. Commence par : terranova quick-start"
