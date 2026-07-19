#!/usr/bin/env bash
# Orchestrate one full stress + fuzz run against the ISOLATED prompt-stress stack.
#
#   preflight (all 7 servers + keycloak healthy)
#   -> seed fixtures (2 courses, phases, student, per-service data, kc roles)
#   -> mint tokens (AFTER role assignment)
#   -> k6 scenarios (smoke, load, spike, soak, exhaustion) -> raw/k6_<scen>.json
#   -> python fuzzer (inputs, authz, idor, files, slowloris)
#   -> build prioritized report (report.md / report.html / findings.json)
#   -> teardown the courses this run created (unless --keep-fixtures)
#
# Usage: api-stress/run.sh [--intensity gentle|medium|brutal]
#                          [--scenarios "smoke load spike soak exhaustion"]
#                          [--smoke-only] [--no-exhaustion] [--keep-fixtures]
set -uo pipefail

STRESS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
command -v uv >/dev/null 2>&1 || { echo "uv not found on PATH (install: brew install uv, or see https://docs.astral.sh/uv/)"; exit 1; }
# Create/refresh the project venv from pyproject.toml + uv.lock. Fast and
# idempotent, so it is safe to run on every invocation.
( cd "$STRESS_DIR" && uv sync --quiet ) || { echo "uv sync failed"; exit 1; }
PY="$STRESS_DIR/.venv/bin/python"
K6="$(command -v k6)" || { echo "k6 not found on PATH (install: brew install k6)"; exit 1; }
CORE_URL="http://localhost:18089"
KC_URL="http://localhost:18081"

INTENSITY="medium"
SCENARIOS="smoke load spike soak exhaustion"
KEEP_FIXTURES=0

# stress.env is gitignored (it is an env file); create it from the committed
# example on first run so the suite works out of the box.
if [[ ! -f "$STRESS_DIR/stress.env" && -f "$STRESS_DIR/stress.env.example" ]]; then
  cp "$STRESS_DIR/stress.env.example" "$STRESS_DIR/stress.env"
  echo "created stress.env from stress.env.example"
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --intensity) INTENSITY="$2"; shift 2;;
    --scenarios) SCENARIOS="$2"; shift 2;;
    --smoke-only) SCENARIOS="smoke"; shift;;
    --no-exhaustion) SCENARIOS="${SCENARIOS//exhaustion/}"; shift;;
    --keep-fixtures) KEEP_FIXTURES=1; shift;;
    *) echo "unknown arg: $1"; exit 2;;
  esac
done

echo "==> regenerate endpoint catalog from partials"
"$PY" "$STRESS_DIR/catalog/merge_catalog.py" || { echo "catalog merge failed"; exit 1; }

echo "==> preflight"
fail=0
for url in \
  "$CORE_URL/api/hello" \
  "http://localhost:18083/team-allocation/api/info" \
  "http://localhost:18084/self-team-allocation/api/info" \
  "http://localhost:18085/assessment/api/info" \
  "http://localhost:18086/example-service/api/info" \
  "http://localhost:18087/interview/api/info" \
  "http://localhost:18088/certificate/api/info" \
  "$KC_URL/realms/prompt/.well-known/openid-configuration"; do
  code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 "$url")
  if [[ "$code" != "200" ]]; then echo "  UNHEALTHY ($code): $url"; fail=1; else echo "  ok: $url"; fi
done
if [[ $fail -ne 0 ]]; then
  echo "Stack not healthy. Bring it up with:"
  echo "  docker compose --env-file api-stress/stress.env -f docker-compose.yml -f api-stress/docker-compose.stress.yml -p prompt-stress up -d"
  exit 1
fi

# Ensure access tokens outlive a full run (default realm lifespan is 300s, but a
# full run is several minutes). Idempotent; only touches the isolated realm.
echo "==> ensure long token lifespan"
"$PY" - <<'PYEOF'
import httpx
KC="http://localhost:18081"
try:
    t=httpx.post(f"{KC}/realms/master/protocol/openid-connect/token",
      data={"grant_type":"password","client_id":"admin-cli","username":"admin","password":"admin"},timeout=15).json()["access_token"]
    h={"Authorization":f"Bearer {t}","Content-Type":"application/json"}
    httpx.put(f"{KC}/admin/realms/prompt",headers=h,
      json={"accessTokenLifespan":3600,"ssoSessionIdleTimeout":7200,"ssoSessionMaxLifespan":7200},timeout=15)
    print("  accessTokenLifespan -> 3600s")
except Exception as e:
    print("  WARN could not bump token lifespan:",e)
PYEOF

# raise file-descriptor limit so a localhost connection flood stresses the SERVER,
# not the runner (ephemeral-port / fd exhaustion on the client side).
ulimit -n 16384 2>/dev/null || true

TS="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$STRESS_DIR/reports/$TS"
mkdir -p "$RUN_DIR/raw"
echo "==> run dir: $RUN_DIR"

cat > "$RUN_DIR/meta.json" <<EOF
{"started":"$TS","intensity":"$INTENSITY","scenarios":"$SCENARIOS",
 "stack":"prompt-stress (isolated, keycloak 26.4.7)","catalog_count":252,
 "core_url":"$CORE_URL"}
EOF

echo "==> seed fixtures"
"$PY" "$STRESS_DIR/lib/fixtures.py" "$RUN_DIR" || { echo "fixtures failed"; exit 1; }

echo "==> mint tokens (post role-assignment)"
"$PY" "$STRESS_DIR/lib/auth.py" "$RUN_DIR/tokens.json" || { echo "tokens failed"; exit 1; }

run_k6 () {
  local scen="$1"; local file="$2"
  echo "==> k6: $scen (intensity=$INTENSITY)"
  STRESS_DIR="$STRESS_DIR" RUN_DIR="$RUN_DIR" INTENSITY="$INTENSITY" SCENARIO="$scen" \
    "$K6" run --quiet --no-usage-report --out "json=$RUN_DIR/raw/k6_${scen}.json" "$file" \
    > "$RUN_DIR/raw/k6_${scen}.log" 2>&1
  echo "    exit=$? log=$RUN_DIR/raw/k6_${scen}.log"
}

for scen in $SCENARIOS; do
  case "$scen" in
    smoke|load|spike|soak) run_k6 "$scen" "$STRESS_DIR/k6/scenario.js";;
    exhaustion) run_k6 "exhaustion" "$STRESS_DIR/k6/exhaustion.js";;
    *) echo "  skip unknown scenario: $scen";;
  esac
done

echo "==> re-mint tokens before fuzz (insurance against expiry during long k6 phase)"
"$PY" "$STRESS_DIR/lib/auth.py" "$RUN_DIR/tokens.json" >/dev/null 2>&1 && echo "    tokens refreshed"

echo "==> fuzz"
MB=4; [[ "$INTENSITY" == "brutal" ]] && MB=16; [[ "$INTENSITY" == "gentle" ]] && MB=1
"$PY" "$STRESS_DIR/fuzz/fuzz.py" "$RUN_DIR" --max-body-mb "$MB" || echo "  fuzz returned nonzero (continuing)"

echo "==> build report"
"$PY" "$STRESS_DIR/report/build_report.py" "$RUN_DIR"

if [[ $KEEP_FIXTURES -eq 0 ]]; then
  echo "==> teardown (delete this run's courses)"
  ADMIN=$("$PY" - "$RUN_DIR" <<'PYEOF'
import json,sys
print(json.load(open(sys.argv[1]+"/tokens.json"))["tokens"].get("admin",""))
PYEOF
)
  for cid in $("$PY" - "$RUN_DIR" <<'PYEOF'
import json,sys
fx=json.load(open(sys.argv[1]+"/fixtures.json"))
print(fx.get("course_a",{}).get("id",""))
print(fx.get("course_b",{}).get("id",""))
PYEOF
); do
    [[ -n "$cid" ]] && curl -s -o /dev/null -w "  deleted course $cid -> %{http_code}\n" \
      -X DELETE -H "Authorization: Bearer $ADMIN" "$CORE_URL/api/courses/$cid"
  done
fi

echo ""
echo "==================== DONE ===================="
echo "Report:   $RUN_DIR/report.md"
echo "Findings: $RUN_DIR/findings.json"
echo "HTML:     $RUN_DIR/report.html"
