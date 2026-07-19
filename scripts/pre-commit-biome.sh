#!/usr/bin/env bash
# Run Biome (check-only, matching `make lint`) on the files staged in git, using
# the version pinned in the clients workspace. Invoked by the `pre-commit`
# framework only when clients/** files are staged. Biome resolves the staged
# file set itself via git (vcs.enabled in clients/biome.json).
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}/clients"

exec yarn biome check --staged --no-errors-on-unmatched
