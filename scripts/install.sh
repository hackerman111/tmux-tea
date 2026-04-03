#!/usr/bin/env bash
set -euo pipefail

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$CURRENT_DIR"

echo "Building tmux-tea..."
mkdir -p bin
go build -o bin/tmux-tea ./cmd/tmux-tea
echo "Done: bin/tmux-tea"
