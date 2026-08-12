#!/usr/bin/env bash
#
# update-tsgo.sh — vendor microsoft/typescript-go into third_party/typescript-go
#
# typescript-go exposes only `internal/` packages, which Go forbids importing
# from other modules. This script copies the source tree into this module,
# renames the top-level `internal` directory to `ts` (so the packages become
# importable from anywhere in this module), and rewrites import paths to
#   github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/...
# The nested `ts/vfs/internal` package is left alone: it is only imported by
# code under `ts/vfs/`, so the internal-visibility rule is still satisfied.
#
# To upgrade: put a new commit hash in third_party/TSGO_COMMIT and re-run.
# After running, verify `go.mod` requirements match typescript-go's go.mod
# (see the diff printed at the end) and run `go mod tidy`.
set -euo pipefail

MODULE="github.com/jclyons52/ts-go-morph"
UPSTREAM="https://github.com/microsoft/typescript-go.git"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$REPO_ROOT/third_party/typescript-go"
COMMIT_FILE="$REPO_ROOT/third_party/TSGO_COMMIT"
COMMIT="$(cat "$COMMIT_FILE")"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "==> Fetching microsoft/typescript-go @ $COMMIT"
git clone --quiet --filter=blob:none "$UPSTREAM" "$TMP/clone"
git -C "$TMP/clone" checkout --quiet "$COMMIT"

echo "==> Copying tree (excluding .git, testdata, _extension, _packages, _submodules)"
rm -rf "$DEST"
mkdir -p "$DEST"
(cd "$TMP/clone" && \
  find . -type d \( -name .git -o -name testdata -o -name _extension -o -name _packages -o -name _submodules \) -prune -o -type f -print \
  | grep -v '^\./\.' \
  | while read -r f; do
      mkdir -p "$DEST/$(dirname "$f")"
      cp "$f" "$DEST/$f"
    done)

echo "==> Renaming internal/ to ts/ and rewriting import paths"
mv "$DEST/internal" "$DEST/ts"
find "$DEST" -name '*.go' -exec sed -i '' \
  -e "s|github.com/microsoft/typescript-go/internal/|$MODULE/third_party/typescript-go/ts/|g" \
  -e "s|github.com/microsoft/typescript-go|$MODULE/third_party/typescript-go|g" \
  {} +

echo "==> Done. Vendored typescript-go @ $COMMIT"
echo "==> Next steps:"
echo "    1. Compare third_party/typescript-go/go.mod with ./go.mod and sync 'require' versions."
echo "    2. go mod tidy"
echo "    3. go build ./third_party/..."
