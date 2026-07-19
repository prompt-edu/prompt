#!/usr/bin/env bash
# Fail if any staged Go file is not gofmt-clean (check-only; does not rewrite).
# Invoked by the `pre-commit` framework with the staged .go files as arguments.
set -euo pipefail

[[ $# -eq 0 ]] && exit 0

unformatted="$(gofmt -l "$@")"
if [[ -n "${unformatted}" ]]; then
  echo "The following Go files are not gofmt-clean:"
  echo "${unformatted}"
  echo "Fix with: gofmt -w <files>"
  exit 1
fi
