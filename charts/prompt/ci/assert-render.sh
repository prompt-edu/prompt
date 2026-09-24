#!/usr/bin/env bash
#
# Renders the chart and asserts the properties that regressions would silently
# break. Run from the repository root: charts/prompt/ci/assert-render.sh
#
# These are render-time checks only. Anything the kubelet decides at container
# start (image UIDs, capabilities, whether the DB gate actually blocks) needs a
# real cluster - see docs/admin/kubernetes for the manual install check.
set -uo pipefail

CHART=${CHART:-charts/prompt}
CI=$CHART/ci
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
failures=0

pass() { printf 'ok   %s\n' "$1"; }
fail() { printf 'FAIL %s\n' "$1" >&2; failures=$((failures + 1)); }

render() {
  local out=$1
  shift
  if ! helm template prompt "$CHART" "$@" >"$out" 2>"$out.err"; then
    fail "render failed: $* ($(tail -1 "$out.err"))"
    return 1
  fi
}

# assert_count <description> <expected> <pattern> <file>
assert_count() {
  local desc=$1 expected=$2 pattern=$3 file=$4 actual
  actual=$(grep -c -- "$pattern" "$file")
  if [ "$actual" = "$expected" ]; then pass "$desc"; else
    fail "$desc (expected $expected, got $actual)"
  fi
}

# assert_render_fails <description> <expected message substring> <helm args...>
assert_render_fails() {
  local desc=$1 needle=$2
  shift 2
  local out=$WORK/negative
  if helm template prompt "$CHART" "$@" >"$out" 2>&1; then
    fail "$desc (render succeeded, expected failure)"
  elif grep -q -- "$needle" "$out"; then
    pass "$desc"
  else
    fail "$desc (wrong error: $(tail -1 "$out"))"
  fi
}

echo "== rendering fixtures =="
for f in "$CI"/*-values.yaml; do
  name=$(basename "$f" -values.yaml)
  render "$WORK/$name.yaml" -f "$f" && pass "renders: $name"
done

DEFAULT=$WORK/default.yaml
[ -s "$DEFAULT" ] || { echo "default fixture did not render; aborting" >&2; exit 1; }

echo
echo "== default render =="
# One securityContext per backend pod (6) carries an explicit non-root UID,
# because the Go images build on distroless static (root variant, no USER).
assert_count "backends pin runAsUser" 6 'runAsUser: 65532' "$DEFAULT"
assert_count "every pod sets seccompProfile" 18 'type: RuntimeDefault' "$DEFAULT"
# 6 backend containers + 6 DB-wait init containers + 4 SeaweedFS containers.
# Frontends must NOT appear here: stock nginx needs its capabilities.
assert_count "capabilities dropped 16x, never on a frontend" 16 'drop: \["ALL"\]' "$DEFAULT"
assert_count "backends carry a DB checksum" 6 'checksum/db:' "$DEFAULT"
# The 14 application workloads (6 backends + 8 frontends). The S3 gateway reads
# only the Secret, so it carries that checksum and not the config one.
assert_count "config checksum on every application workload" 14 'checksum/appconfig:' "$DEFAULT"
assert_count "secret checksum also on the S3 gateway" 15 'checksum/appsecrets:' "$DEFAULT"
assert_count "DB gate uses a psql client" 6 'image: postgres:17-alpine' "$DEFAULT"
assert_count "no broken realm import is shipped" 0 'KeycloakRealmImport' "$DEFAULT"
assert_count "no template phase (its image is never built)" 0 'template-component' "$DEFAULT"
assert_count "example phase is deployed" 1 'prompt-clients-example-component' "$DEFAULT"
assert_count "example phase is routed" 1 'value: "/example"' "$DEFAULT"
assert_count "SeaweedFS internals are firewalled" 1 'kind: NetworkPolicy' "$DEFAULT"
# CNPG Cluster + one Secret per managed role.
assert_count "data-bearing resources survive uninstall" 7 'helm.sh/resource-policy: keep' "$DEFAULT"
assert_count "SSL_MODE is set" 1 'SSL_MODE:' "$DEFAULT"
assert_count "intro-course remote is unset, not the apex" 1 'INTRO_COURSE_HOST: ""' "$DEFAULT"
assert_count "devops-challenge remote is unset, not the apex" 1 'DEVOPS_CHALLENGE_HOST: ""' "$DEFAULT"

echo
echo "== mode combinations =="
render "$WORK/norl.yaml" -f "$CI/default-values.yaml" \
  --set global.gateway.rateLimiting.enabled=false &&
  assert_count "disabling rate limiting keeps S3 CORS" 1 'kind: SecurityPolicy' "$WORK/norl.yaml" &&
  assert_count "disabling rate limiting drops the rate limit" 0 'kind: BackendTrafficPolicy' "$WORK/norl.yaml"

assert_count "provider=none emits no Envoy CRDs" 0 'gateway.envoyproxy.io' "$WORK/phases-disabled.yaml"
assert_count "external PostgreSQL uses the configured account" 6 'DB_USER: "prompt"' "$WORK/external-postgres.yaml"
assert_count "external PostgreSQL requires TLS by default" 1 'SSL_MODE: "require"' "$WORK/external-postgres.yaml"
assert_count "bundled Keycloak is deployed" 1 '^kind: Keycloak$' "$WORK/in-cluster-keycloak.yaml"

echo
echo "== external credentials are stable across renders =="
render "$WORK/ext-a.yaml" -f "$CI/external-postgres-values.yaml" &&
  render "$WORK/ext-b.yaml" -f "$CI/external-postgres-values.yaml" &&
  if diff -q <(grep 'DB_PASSWORD' "$WORK/ext-a.yaml") <(grep 'DB_PASSWORD' "$WORK/ext-b.yaml") >/dev/null; then
    pass "external DB passwords do not change between renders"
  else
    fail "external DB passwords changed between renders"
  fi
# In-cluster passwords are deliberately not checked: `lookup` returns nothing
# without a cluster, so those renders are expected to differ.

echo
echo "== required values =="
assert_render_fails "S3 secret is required" \
  'S3_SECRET_KEY is required' \
  --set global.host=prompt.ci.example.com
assert_render_fails "external PostgreSQL password is required" \
  'external.password is required' \
  -f "$CI/default-values.yaml" --set global.postgresql.mode=external \
  --set global.postgresql.external.host=db.example.com
assert_render_fails "external PostgreSQL host is required" \
  'external.host is required' \
  -f "$CI/default-values.yaml" --set global.postgresql.mode=external \
  --set global.postgresql.external.password=pw
assert_render_fails "external S3 access key is required" \
  'S3_ACCESS_KEY is required' \
  -f "$CI/external-object-storage-values.yaml" \
  --set global.appSecrets.data.S3_ACCESS_KEY=
assert_render_fails "bundled Keycloak rejects external PostgreSQL" \
  'requires global.postgresql.mode=in-cluster' \
  -f "$CI/external-postgres-values.yaml" --set global.keycloak.mode=in-cluster

echo
if [ "$failures" -gt 0 ]; then
  echo "$failures assertion(s) failed" >&2
  exit 1
fi
echo "all assertions passed"
