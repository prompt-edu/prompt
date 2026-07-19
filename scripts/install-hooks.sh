#!/usr/bin/env bash
# Install the project's git hooks via the pre-commit framework
# (https://pre-commit.com). Idempotent.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

if ! command -v pre-commit >/dev/null 2>&1; then
  echo "Error: pre-commit is not installed."
  echo "Install it with one of:"
  echo "  pipx install pre-commit       # recommended"
  echo "  pip install --user pre-commit"
  echo "  brew install pre-commit       # macOS"
  echo "See https://pre-commit.com/#install"
  exit 1
fi

# Older setups pointed core.hooksPath at .githooks; pre-commit installs into the
# default .git/hooks, so clear that override if it is still set.
if git config --local --get core.hooksPath >/dev/null 2>&1; then
  git config --local --unset core.hooksPath
  echo "Cleared legacy core.hooksPath override."
fi

# The swagger hook shells out to `swag`; install it when Go is available so the
# hook works out of the box.
if command -v go >/dev/null 2>&1; then
  if ! command -v swag >/dev/null 2>&1; then
    echo "Installing swag (used by the swagger hook)..."
    go install github.com/swaggo/swag/cmd/swag@latest
    echo "Ensure \$(go env GOPATH)/bin (or \$GOBIN) is on your PATH."
  fi
else
  echo "Warning: Go not found; install 'swag' before committing Go changes."
fi

pre-commit install
echo "pre-commit hooks installed. Run 'pre-commit run --all-files' to check the whole repo."
