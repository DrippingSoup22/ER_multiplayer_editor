#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

arch="amd64"
version="${VERSION:-0.0.0}"
version="${version#v}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  version="0.0.0"
fi
package="ER-Multiplayer-Editor-windows-$arch-$version"
archive="dist/$package.zip"
staging="$(mktemp -d)"
trap 'rm -rf "$staging"' EXIT
bundle="$staging/$package"
mkdir -p dist "$bundle"

echo "building ER Multiplayer Editor for windows/$arch"
CGO_ENABLED=0 GOOS=windows GOARCH="$arch" \
  go build -trimpath -ldflags "-H windowsgui -s -w" \
  -o "$bundle/ERMultiplayerEditor.exe" ./cmd/ermultiplayereditor

rm -f "$archive"
go run github.com/DrippingSoup22/ER_shared/releasezip -root "$bundle" -out "$archive"
