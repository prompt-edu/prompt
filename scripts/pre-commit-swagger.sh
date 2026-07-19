#!/usr/bin/env bash
# Regenerate and re-stage the committed swagger docs for any Go service whose
# sources are staged. Invoked by the `pre-commit` framework (see
# .pre-commit-config.yaml); requires `swag` on PATH.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

staged_files="$(git diff --cached --name-only)"
if ! echo "${staged_files}" | grep -Eq '^servers/(core|example_server|self_team_allocation|assessment|team_allocation|interview)/'; then
  exit 0
fi

swag_bin=""
if command -v swag >/dev/null 2>&1; then
  swag_bin="swag"
else
  go_bin="$(go env GOBIN 2>/dev/null || true)"
  if [[ -z "${go_bin}" ]]; then
    go_bin="$(go env GOPATH 2>/dev/null || true)/bin"
  fi
  # Fallback to default Go install location when go is not on PATH
  if [[ -z "${go_bin}" || ! -x "${go_bin}/swag" ]]; then
    go_bin="${HOME}/go/bin"
  fi
  if [[ -n "${go_bin}" && -x "${go_bin}/swag" ]]; then
    swag_bin="${go_bin}/swag"
  fi
fi

if [[ -z "${swag_bin}" ]]; then
  echo "Error: swag is not installed or not on PATH."
  echo "Install with: go install github.com/swaggo/swag/cmd/swag@latest"
  echo "Then ensure \$GOPATH/bin (or \$GOBIN) is on PATH."
  exit 1
fi

regenerate() {
  local service="$1"
  if echo "${staged_files}" | grep -Eq "^servers/${service}/"; then
    SWAG_BIN="${swag_bin}" ./scripts/generate-api-spec.sh "${service}"
    git add "servers/${service}/docs"
  fi
}

regenerate core
regenerate example_server
regenerate self_team_allocation
regenerate team_allocation
regenerate assessment
regenerate interview
