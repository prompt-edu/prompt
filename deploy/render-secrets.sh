#!/usr/bin/env bash
# Generate the real, secret-filled deploy files on the VM (kept out of git).
#   - deploy/.env.websec                  (from .env.websec.example, __GENERATE__ -> random)
#   - deploy/keycloakConfig.rendered.json (from keycloakConfig.websec.json, secret injected)
# Re-running refuses to clobber an existing .env.websec unless called with --force.
set -euo pipefail

cd "$(dirname "$0")"

FORCE="${1:-}"
ENV_OUT=".env.websec"
ENV_IN=".env.websec.example"
REALM_IN="keycloakConfig.websec.json"
REALM_OUT="keycloakConfig.rendered.json"

gen() { openssl rand -hex 24; }

if [[ -f "$ENV_OUT" && "$FORCE" != "--force" ]]; then
  echo "[render-secrets] $ENV_OUT already exists; refusing to overwrite (pass --force to regenerate)."
else
  cp "$ENV_IN" "$ENV_OUT"
  # Each secret var gets its own fresh value.
  for var in DB_CORE_PASSWORD DB_INTRO_COURSE_PASSWORD KEYCLOAK_DB_PASSWORD KEYCLOAK_CLIENT_SECRET KEYCLOAK_ADMIN_PASSWORD S3_ACCESS_KEY S3_SECRET_KEY; do
    secret="$(gen)"
    # Replace only the exact "VAR=__GENERATE__" line.
    sed -i "s|^${var}=__GENERATE__$|${var}=${secret}|" "$ENV_OUT"
  done
  if grep -q "__GENERATE__" "$ENV_OUT"; then
    echo "[render-secrets] ERROR: unresolved __GENERATE__ placeholders remain in $ENV_OUT:" >&2
    grep -n "__GENERATE__" "$ENV_OUT" >&2
    exit 1
  fi
  echo "[render-secrets] wrote $ENV_OUT"
fi

# Inject the prompt-server client secret into the realm import so it matches KEYCLOAK_CLIENT_SECRET.
KCS="$(grep -E '^KEYCLOAK_CLIENT_SECRET=' "$ENV_OUT" | head -n1 | cut -d= -f2-)"
if [[ -z "$KCS" || "$KCS" == "__GENERATE__" ]]; then
  echo "[render-secrets] ERROR: could not read KEYCLOAK_CLIENT_SECRET from $ENV_OUT" >&2
  exit 1
fi
# The realm "admin" app user reuses the generated Keycloak admin password, so no
# weak/known admin credential is ever committed or shipped.
KCAP="$(grep -E '^KEYCLOAK_ADMIN_PASSWORD=' "$ENV_OUT" | head -n1 | cut -d= -f2-)"
if [[ -z "$KCAP" || "$KCAP" == "__GENERATE__" ]]; then
  echo "[render-secrets] ERROR: could not read KEYCLOAK_ADMIN_PASSWORD from $ENV_OUT" >&2
  exit 1
fi
sed -e "s|__PROMPT_SERVER_SECRET__|${KCS}|g" -e "s|__ADMIN_PASSWORD__|${KCAP}|g" "$REALM_IN" > "$REALM_OUT"
for ph in __PROMPT_SERVER_SECRET__ __ADMIN_PASSWORD__; do
  if grep -q "$ph" "$REALM_OUT"; then
    echo "[render-secrets] ERROR: realm placeholder $ph not substituted" >&2
    exit 1
  fi
done
echo "[render-secrets] wrote $REALM_OUT (prompt-server secret + admin password injected)"
echo "[render-secrets] done."
