#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

if [[ "$(go env GOOS)" != linux ]]; then
  echo "the Linux GUI must be built on Linux" >&2
  exit 1
fi

arch="$(go env GOARCH)"
version="${VERSION:-0.0.0}"
version="${version#v}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  version="0.0.0"
fi
package="ER-Multiplayer-Editor-linux-$arch-$version"
archive="dist/$package.zip"
staging="$(mktemp -d)"
trap 'rm -rf "$staging"' EXIT
bundle="$staging/$package"
mkdir -p dist "$bundle"

echo "building ER Multiplayer Editor for linux/$arch"
CGO_ENABLED=1 go build -trimpath \
  -ldflags "-s -w -X gioui.org/app.ID=io.github.DrippingSoup22.ERMultiplayerEditor" \
  -o "$bundle/ERMultiplayerEditor" ./cmd/ermultiplayereditor

rm -f "$archive"
go run github.com/DrippingSoup22/ER_shared/releasezip -root "$bundle" -out "$archive"
