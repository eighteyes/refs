#!/usr/bin/env bash
# sync.sh — kept for old docs; the logic now lives in `refs sync`
# Responsibilities:
#   - forward all args to `refs sync [--github] [--cactus] [--dry-run]`
#   - fall back to this repo's bin/refs launcher when refs is not on PATH
here=$(cd "$(dirname "$0")" && pwd)
refs_bin=$(command -v refs || echo "$here/../../../bin/refs")
exec "$refs_bin" sync "$@"
