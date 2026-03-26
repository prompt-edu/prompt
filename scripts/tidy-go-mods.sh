#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVERS_DIR="$REPO_ROOT/servers"

if [[ ! -d "$SERVERS_DIR" ]]; then
  echo "Error: servers directory not found at $SERVERS_DIR" >&2
  exit 1
fi

MODULE_DIRS=()
while IFS= read -r module_dir; do
  MODULE_DIRS+=("$module_dir")
done < <(find "$SERVERS_DIR" -name go.mod -type f -exec dirname {} \; | sort)

if [[ ${#MODULE_DIRS[@]} -eq 0 ]]; then
  echo "No go.mod files found under $SERVERS_DIR"
  exit 0
fi

echo "Running go mod tidy for ${#MODULE_DIRS[@]} server module(s)..."

for module_dir in "${MODULE_DIRS[@]}"; do
  rel_path="${module_dir#"$REPO_ROOT"/}"
  echo "==> $rel_path"
  (cd "$module_dir" && go mod tidy)
done

echo "Done."
