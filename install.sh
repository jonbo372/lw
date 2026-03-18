#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

command -v go &>/dev/null || { echo "error: go is required but not installed" >&2; exit 1; }
command -v make &>/dev/null || { echo "error: make is required but not installed" >&2; exit 1; }

case "$(uname -s)" in
  Linux)  DEST="$HOME/.local/bin" ;;
  Darwin) DEST="/usr/local/bin"   ;;
  *)      echo "error: unsupported OS" >&2; exit 1 ;;
esac

mkdir -p "$DEST"

echo "Building lw…"
(cd "$SCRIPT_DIR" && make clean build)

cp "$SCRIPT_DIR/bin/lw" "$DEST/lw"
echo "Installed lw to $DEST/lw"
