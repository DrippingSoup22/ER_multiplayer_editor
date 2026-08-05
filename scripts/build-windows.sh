#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

arch="amd64"
version="${VERSION:-0.0.0}"
version="${version#v}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "error: VERSION must use semantic versioning (for example, 2.7.0)" >&2
  exit 2
fi

package="ER-PvP-Mod-windows-$arch-$version"
bundle="dist/windows-$arch/$package"
archive="dist/$package.zip"

rm -rf "$bundle"
mkdir -p "$bundle"

echo "building ER PvP Mod for windows/$arch"
CGO_ENABLED=0 GOOS=windows GOARCH="$arch" \
  go build -trimpath -ldflags "-H windowsgui -s -w" \
  -o "$bundle/ERPvPMod.exe" ./cmd/gui

rm -f "$archive"
go run ./tools/packagezip -root "$bundle" -out "$archive"
echo "created $archive"
