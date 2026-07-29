#!/usr/bin/env bash
# Back-compat shim. The orchestration now lives in stress/cli.py, exposed as the
# `api-stress-test` console script, so these are equivalent:
#
#   ./run.sh --intensity brutal
#   uv run api-stress-test --intensity brutal
#
# uv run provisions the locked environment (pyproject.toml + uv.lock) on demand.
# Flags: run `uv run api-stress-test --help`.
set -uo pipefail

STRESS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
command -v uv >/dev/null 2>&1 || {
  echo "uv not found on PATH (install: brew install uv, or see https://docs.astral.sh/uv/)"
  exit 1
}

exec uv run --project "$STRESS_DIR" api-stress-test "$@"
