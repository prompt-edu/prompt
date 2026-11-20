#!/usr/bin/env bash
#
# Scaffold a new PROMPT course phase from the prompt-example-phase repository.
#
# Usage: scripts/new-course-phase.sh <name> <client_port> <server_port> [db_port]
#   name         snake_case phase name WITHOUT the _component suffix (e.g. peer_review)
#   client_port  unique dev port for the micro-frontend (e.g. 3014)
#   server_port  unique port for the Go service (e.g. 8093)
#   db_port      unique published Postgres port (default: next free 54xx in docker-compose.yml)
#
# Clones https://github.com/prompt-edu/prompt-example-phase, renames every identifier with
# its init.sh, converts it to the workspace layout of this repository, and registers the
# phase in the workspaces, core client, docker-compose (plus its stack overlay), env
# templates, Makefile, and CI matrices. Prints the remaining manual steps at the end.
# Full documentation: docs/contributor/new_course_phase.md
#
# Environment:
#   PHASE_TEMPLATE_SOURCE  git URL or local path to clone instead of the GitHub repository
#   PHASE_TEMPLATE_REF     branch or tag to clone (default: main)
#
# Run on a clean working tree. If a step fails midway, revert with:
#   git checkout -- . && git clean -fd clients/<name>_component servers/<name> \
#     clients/core/src/managementConsole/PhaseMapping

set -euo pipefail

die() { echo "error: $*" >&2; exit 1; }

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATES="$REPO_ROOT/scripts/new-course-phase"
cd "$REPO_ROOT"
[[ -f clients/core/remotes.config.mjs && -d servers/core ]] || die "run from the PROMPT repository"

SOURCE="${PHASE_TEMPLATE_SOURCE:-https://github.com/prompt-edu/prompt-example-phase.git}"
REF="${PHASE_TEMPLATE_REF:-main}"

NAME="${1:-}"
CLIENT_PORT="${2:-}"
SERVER_PORT="${3:-}"
DB_PORT="${4:-}"

[[ -n "$NAME" && -n "$CLIENT_PORT" && -n "$SERVER_PORT" ]] \
  || die "usage: $0 <name> <client_port> <server_port> [db_port]"
[[ "$NAME" =~ ^[a-z][a-z0-9_]*$ ]] || die "name must be snake_case (got: $NAME)"
[[ "$NAME" != *_component && "$NAME" != *_server ]] \
  || die "pass the bare phase name without _component/_server suffix"
[[ "$NAME" != "example" ]] || die "pick a name other than 'example'"
[[ "$CLIENT_PORT" =~ ^[0-9]+$ && "$SERVER_PORT" =~ ^[0-9]+$ ]] || die "ports must be numeric"

KEBAB="${NAME//_/-}"
UPPER="$(echo "$NAME" | tr '[:lower:]' '[:upper:]')"
PASCAL="$(echo "$NAME" | perl -pe 's/(^|_)([a-z])/\u$2/g')"
READABLE="$(echo "$NAME" | perl -pe 's/_/ /g; s/(^| )([a-z])/$1\u$2/g')"
READABLE_LOWER="${NAME//_/ }"

CLIENT_DIR="clients/${NAME}_component"
SERVER_DIR="servers/${NAME}"
[[ ! -e "$CLIENT_DIR" ]] || die "$CLIENT_DIR already exists"
[[ ! -e "$SERVER_DIR" ]] || die "$SERVER_DIR already exists"

grep -q "devPort: ${CLIENT_PORT}," clients/core/remotes.config.mjs \
  && die "client port ${CLIENT_PORT} is already used by another remote"
# Host ports in docker-compose.yml read "${PROMPT_PORT_<port>:-<port>}:<container port>"
# so a parallel stack can shift them (docs/contributor/setup.md).
grep -qF ":-${CLIENT_PORT}}:80\"" docker-compose.yml \
  && die "client port ${CLIENT_PORT} is already used in docker-compose.yml"
grep -qF ":-${SERVER_PORT}}:8080\"" docker-compose.yml \
  && die "server port ${SERVER_PORT} is already used in docker-compose.yml"

if [[ -z "$DB_PORT" ]]; then
  DB_PORT="$(grep -oE ':-54[0-9]{2}\}:5432"' docker-compose.yml | grep -oE '54[0-9]{2}' | sort -n | tail -1)"
  DB_PORT=$((DB_PORT + 1))
fi
grep -qF ":-${DB_PORT}}:5432\"" docker-compose.yml \
  && die "db port ${DB_PORT} is already used in docker-compose.yml"

json_value() { KEY="$2" perl -ne 'print $1 and exit if /"\Q$ENV{KEY}\E": "([^"]+)"/' "$1"; }
VERSION="$(json_value clients/package.json version)"
RSPACK_VERSION="$(json_value clients/package.json '@rspack/cli')"
PACKAGE_MANAGER="$(json_value clients/package.json packageManager)"
[[ -n "$VERSION" && -n "$RSPACK_VERSION" && -n "$PACKAGE_MANAGER" ]] \
  || die "could not read version, @rspack/cli, or packageManager from clients/package.json"

# render <template> — prints a template from scripts/new-course-phase with placeholders filled
render() {
  T_NAME="$NAME" T_KEBAB="$KEBAB" T_UPPER="$UPPER" T_PASCAL="$PASCAL" T_READABLE="$READABLE" \
  T_READABLE_LOWER="$READABLE_LOWER" T_CLIENT_PORT="$CLIENT_PORT" T_SERVER_PORT="$SERVER_PORT" \
  T_DB_PORT="$DB_PORT" T_VERSION="$VERSION" T_RSPACK_VERSION="$RSPACK_VERSION" \
  T_PACKAGE_MANAGER="$PACKAGE_MANAGER" \
    perl -pe 's/\{\{(\w+)\}\}/exists $ENV{"T_\U$1"} ? $ENV{"T_\U$1"} : die("unknown placeholder $1\n")/ge' \
    "$TEMPLATES/$1"
}

# insert_after <file> <perl-regex matching a full line> <text appended as a new line>
insert_after() {
  local file="$1" pattern="$2" newline="$3"
  MATCH_PATTERN="$pattern" NEW_LINE="$newline" perl -pi -e '
    BEGIN { $done = 0 }
    if (!$done && /$ENV{MATCH_PATTERN}/) { $_ .= "$ENV{NEW_LINE}\n"; $done = 1 }
  ' "$file"
  grep -qF "$newline" "$file" || die "failed to insert into $file (anchor: $pattern)"
}

# insert_template_before <file> <perl-regex matching a full line> <template>
insert_template_before() {
  local file="$1" pattern="$2" block
  block="$(mktemp)"
  render "$3" > "$block"
  MATCH_PATTERN="$pattern" BLOCK_FILE="$block" perl -pi -e '
    BEGIN { $done = 0; local $/; open my $fh, "<", $ENV{BLOCK_FILE} or die; $text = <$fh> }
    if (!$done && /$ENV{MATCH_PATTERN}/) { $_ = $text . $_; $done = 1 }
  ' "$file"
  grep -qF "$(head -1 "$block")" "$file" || die "failed to insert into $file (anchor: $pattern)"
  rm -f "$block"
}

echo "Scaffolding phase '${NAME}' (client :${CLIENT_PORT}, server :${SERVER_PORT}, db :${DB_PORT})"

# ------------------------------------------------------- 1. fetch and rename
echo "-> fetching ${SOURCE} (${REF})"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
git -c advice.detachedHead=false clone --quiet --depth 1 --branch "$REF" "$SOURCE" "$WORK/phase" 2>/dev/null \
  || die "could not clone ${SOURCE} (${REF})"
PHASE_GO_MODULE="github.com/prompt-edu/prompt/servers/${NAME}" \
  bash "$WORK/phase/init.sh" --no-verify "$NAME" "$CLIENT_PORT" "$SERVER_PORT" >/dev/null \
  || die "init.sh of the example phase failed"

# ---------------------------------------------------------------- 2. client
echo "-> ${CLIENT_DIR}"
SRC_CLIENT="$WORK/phase/client"
rm -rf "$SRC_CLIENT/shared" "$SRC_CLIENT/node_modules" "$SRC_CLIENT/build" "$SRC_CLIENT/.yarn" \
  "$SRC_CLIENT/yarn.lock" "$SRC_CLIENT/nginx.conf" "$SRC_CLIENT/README.md"
perl -pi -e "s#'\./shared/#'../shared/#" "$SRC_CLIENT/rspack.config.mjs"
perl -pi -e "s#'\.\./shared/#'../../shared/#" "$SRC_CLIENT/src/bootstrap.tsx"
grep -q "'\.\./shared/" "$SRC_CLIENT/rspack.config.mjs" \
  || die "shared import not found in rspack.config.mjs of the example phase"
perl -pi -e 's/"rootDir": "\."/"rootDir": ".."/' "$SRC_CLIENT/tsconfig.json"
render client.package.json.tmpl > "$SRC_CLIENT/package.json"
render client.biome.json.tmpl > "$SRC_CLIENT/biome.json"
render client.Dockerfile.tmpl > "$SRC_CLIENT/Dockerfile"
cp clients/core/.dockerignore "$SRC_CLIENT/.dockerignore"
mv "$SRC_CLIENT" "$CLIENT_DIR"

# ---------------------------------------------------------------- 3. server
echo "-> ${SERVER_DIR}"
rm -f "$WORK/phase/server/Dockerfile" "$WORK/phase/server/README.md"
mv "$WORK/phase/server" "$SERVER_DIR"

# ------------------------------------------------------- 4. yarn workspaces
echo "-> workspaces"
perl -pi -e "s/^(\s*)\"core\",\$/\$1\"core\",\n\$1\"${NAME}_component\",/" \
  clients/package.json clients/lerna.json
grep -q "\"${NAME}_component\"," clients/package.json || die "workspace not added to clients/package.json"
grep -q "\"${NAME}_component\"," clients/lerna.json || die "workspace not added to clients/lerna.json"
insert_after clients/Dockerfile '^COPY core/package.json ' \
  "$(printf 'COPY %-43s %s' "${NAME}_component/package.json" "${NAME}_component/package.json")"

# ------------------------------------------------------------- 5. core client
echo "-> core client registration"
insert_after clients/core/remotes.config.mjs \
  'export const REMOTES = \[' \
  "  { name: '${NAME}_component', phaseTypeName: '${NAME}_component', devPort: ${CLIENT_PORT}, prodPath: '/${KEBAB}' },"

PM=clients/core/src/managementConsole/PhaseMapping
render Routes.tsx.tmpl > "$PM/ExternalRoutes/${PASCAL}Routes.tsx"
render Sidebar.tsx.tmpl > "$PM/ExternalSidebars/${PASCAL}Sidebar.tsx"
insert_after "$PM/PhaseRouterMapping.tsx" \
  '^import \{ ApplicationRoutes \}' \
  "import { ${PASCAL}Routes } from './ExternalRoutes/${PASCAL}Routes'"
insert_after "$PM/PhaseRouterMapping.tsx" \
  '^const PhaseRouter: .* = \{$' \
  "  ${NAME}_component: ${PASCAL}Routes,"
insert_after "$PM/PhaseSidebarMapping.tsx" \
  '^import \{ ApplicationSidebar \}' \
  "import { ${PASCAL}Sidebar } from './ExternalSidebars/${PASCAL}Sidebar'"
insert_after "$PM/PhaseSidebarMapping.tsx" \
  '^\} = \{$' \
  "  ${NAME}_component: ${PASCAL}Sidebar,"
insert_after "$PM/PhaseStudentDetailMapping.tsx" \
  '^\} = \{$' \
  "  ${NAME}_component: Fallback,"

insert_after clients/core/public/env.js \
  "^  CORE_HOST: " \
  "  ${UPPER}_HOST: 'http://localhost:${SERVER_PORT}',"
insert_after clients/core/public/env.template.js \
  "^  CORE_HOST: " \
  "  ${UPPER}_HOST: '\$${UPPER}_HOST',"

# ------------------------------------------------------------ 6. docker-compose
echo "-> docker-compose.yml, docker-compose.stack.yml"
insert_template_before docker-compose.yml '^  client-core:$' compose.server.yml.tmpl
insert_template_before docker-compose.yml '^  clients-base:$' compose.client.yml.tmpl
insert_template_before docker-compose.yml '^  keycloak-db:$' compose.db.yml.tmpl
insert_after docker-compose.yml '^      - CORE_HOST=\$\{CORE_API_HOST\}$' "      - ${UPPER}_HOST"
# A parallel stack keeps the new database in a named volume, like the others.
insert_after docker-compose.stack.yml '^      - postgres_data:/var/lib/postgresql/data$' \
  "  db-${KEBAB}-server:
    volumes:
      - postgres_${NAME}_data:/var/lib/postgresql/data"
insert_after docker-compose.stack.yml '^  postgres_data:$' "  postgres_${NAME}_data:"

# ------------------------------------------------------------- 7. env templates
echo "-> env templates"
cat >> .env.template <<EOF

# ============================================================================
# ${UPPER} SERVER DATABASE CONFIGURATION
# ============================================================================

DB_HOST_${UPPER}_SERVER=db-${KEBAB}-server
DB_PORT_${UPPER}_SERVER=5432
DB_${UPPER}_NAME=prompt
DB_${UPPER}_USER=prompt-postgres
DB_${UPPER}_PASSWORD=prompt-postgres
${UPPER}_HOST=http://localhost:${SERVER_PORT}
SENTRY_DSN_${UPPER}_SERVER=
AUDIT_INGEST_KEY_${UPPER}_SERVER=
EOF
cat >> .env.dev.template <<EOF

# ${READABLE}
DB_HOST_${UPPER}_SERVER=localhost
DB_PORT_${UPPER}_SERVER=${DB_PORT}
EOF

# ------------------------------------------------------------------ 8. Makefile
echo "-> Makefile"
perl -pi -e "s/\bserver-core\b/server-core server-${KEBAB}/ if \$. < 20" Makefile
perl -pi -e "s/\btest-core\b/test-core test-${KEBAB}/ if \$. < 20" Makefile
perl -pi -e "s/\bsqlc-core\b/sqlc-core sqlc-${KEBAB}/ if \$. < 20" Makefile
insert_after Makefile '^\t@\$\(MAKE\) server-core &$' "	@\$(MAKE) server-${KEBAB} &"
# The host-run server reaches its database on the port of this stack.
insert_after Makefile '^export DB_CORE_PORT := ' \
  "export DB_PORT_${UPPER}_SERVER := \$(PROMPT_PORT_${DB_PORT})"
insert_after Makefile '^\tcd servers/core && go vet' "	cd servers/${NAME} && go vet ./..."
perl -pi -e "s/^(test: test-clients test-core)/\$1 test-${KEBAB}/" Makefile
perl -pi -e "s/^(sqlc: sqlc-core)/\$1 sqlc-${KEBAB}/" Makefile
# make db starts the databases listed here, in every stack.
perl -pi -e "s/^(DB_SERVICES = db )/\$1db-${KEBAB}-server /" Makefile
grep -q "^test: .*test-${KEBAB}" Makefile || die "test-${KEBAB} not added to the test target"
grep -q "^sqlc: .*sqlc-${KEBAB}" Makefile || die "sqlc-${KEBAB} not added to the sqlc target"
grep -q "^DB_SERVICES = .*db-${KEBAB}-server" Makefile || die "db-${KEBAB}-server not added to DB_SERVICES"
printf '\nserver-%s: ## Start %s server (port %s)\n\t$(call run-server,%s,$(PROMPT_PORT_%s))\n' \
  "$KEBAB" "$NAME" "$SERVER_PORT" "$NAME" "$SERVER_PORT" >> Makefile
printf '\ntest-%s: ## Run %s server tests\n\tcd servers/%s && go test ./...\n' \
  "$KEBAB" "$NAME" "$NAME" >> Makefile
printf '\nsqlc-%s: ## Generate sqlc code for %s server\n\tcd servers/%s && sqlc generate\n' \
  "$KEBAB" "$NAME" "$NAME" >> Makefile

# ------------------------------------------------------------------------ 9. CI
echo "-> CI matrices"
insert_after .github/workflows/quality-clients.yml '^          - core$' \
  "          - ${NAME}_component"
insert_after .github/workflows/quality-servers.yml '^          - core$' \
  "          - ${NAME}"
insert_after .github/workflows/test-servers.yml '^          - core$' \
  "          - ${NAME}"

# ---------------------------------------------------------------- 10. verify
echo "-> verifying"
(cd "$SERVER_DIR" && go build ./...) || die "go build failed in $SERVER_DIR"
(cd clients && yarn install --mode update-lockfile >/dev/null 2>&1 && yarn install >/dev/null 2>&1) \
  || die "yarn install failed"
(cd clients && yarn tsc -p "${NAME}_component/tsconfig.json" --noEmit --pretty false) \
  || die "typecheck failed for ${NAME}_component"
CORE_TOUCHED=(core/remotes.config.mjs core/src/managementConsole/PhaseMapping)
(cd clients && yarn biome check --write "${NAME}_component" "${CORE_TOUCHED[@]}" >/dev/null 2>&1) || true
(cd clients && yarn biome check --diagnostic-level=error "${NAME}_component" "${CORE_TOUCHED[@]}" >/dev/null) \
  || die "biome check failed for ${NAME}_component or the touched core files"

cat <<EOF

Done. Phase '${NAME}' scaffolded and registered.

Remaining manual steps (see docs/contributor/new_course_phase.md):
  1. Course phase type: add init${PASCAL}() in
     servers/core/coursePhaseType/initializeTypes.go (name key: '${NAME}_component',
     dev BaseUrl: devServiceURL(${SERVER_PORT}, "/${KEBAB}-service/api")).
  2. Deployment: docker-compose.prod.yml service + traefik labels, and the
     build-and-push-clients.yml (pass MF_BUILD_VERSION like the other remotes) /
     deploy-docker.yml / dev.yml / prod.yml wiring.
     For the audit log, add an AUDIT_INGEST_KEY_${UPPER}_SERVER secret to
     deploy-docker.yml and list the service in core's AUDIT_INGEST_KEYS.
  3. Optional: add ${UPPER}_HOST to the EnvType in @tumaet/prompt-shared-state.
  4. Implement your schema (db/migration, db/query, 'make sqlc-${KEBAB}') and logic.

Try it: make db && make server-${KEBAB}   and   make clients
EOF
