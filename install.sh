#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SOURCE="$SCRIPT_DIR/scripts/lw"

if [[ ! -f "$SOURCE" ]]; then
  echo "error: scripts/lw not found" >&2
  exit 1
fi

case "$(uname -s)" in
  Linux)  DEST="$HOME/.local/bin" ;;
  Darwin) DEST="/usr/local/bin"   ;;
  *)      echo "error: unsupported OS" >&2; exit 1 ;;
esac

mkdir -p "$DEST"
cp "$SOURCE" "$DEST/lw"
chmod +x "$DEST/lw"

echo "Installed lw to $DEST/lw"
