#!/usr/bin/env bash
# Run Biome (check-only, matching `make lint`) on the client files passed by the
# `pre-commit` framework, using the version pinned in the clients workspace. The
# framework forwards the filtered file set (repo-relative paths), so this also
# works with `pre-commit run --all-files`.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}/clients"

# Nothing to check (e.g. all matched files were unstaged and stashed away).
if [[ $# -eq 0 ]]; then
  exit 0
fi

# Biome runs from clients/, so strip the clients/ prefix from the repo-relative
# paths the framework passes.
files=()
for f in "$@"; do
  files+=("${f#clients/}")
done

exec yarn biome check "${files[@]}" --no-errors-on-unmatched
