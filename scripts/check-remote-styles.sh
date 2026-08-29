#!/usr/bin/env bash
# Fail if a micro-frontend declares a Tailwind entry point of its own.
#
# clients/core builds the only Tailwind stylesheet in the document and scans
# every component directory for it. A second build injected by a remote lands
# in <head> after core's and overrides core's utilities at equal specificity
# (see issue #2086). Plain CSS in a remote is fine; Tailwind directives are not.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}/clients"

failed=0

configs=$(find . -maxdepth 2 -path './*_component/tailwind.config.*' -print)
if [[ -n "${configs}" ]]; then
  echo "Remote Tailwind config found:" >&2
  echo "${configs}" >&2
  failed=1
fi

directives=$(grep -rnE "@(tailwind|config|source|apply)\b|@import[^;]*tailwindcss" \
  --include='*.css' ./*_component || true)
if [[ -n "${directives}" ]]; then
  echo "Tailwind directive in a remote stylesheet:" >&2
  echo "${directives}" >&2
  failed=1
fi

if [[ "${failed}" -ne 0 ]]; then
  cat >&2 <<'EOF'

Remotes must not build Tailwind. Add the utilities you need by letting
clients/core/tailwind.config.js scan your sources - it already globs
clients/*_component/{src,routes,sidebar}.
EOF
  exit 1
fi
