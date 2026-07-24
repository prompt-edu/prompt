#!/usr/bin/env bash
#
# Scaffold a new PROMPT course phase from the example component + example server.
#
# Usage: scripts/new-course-phase.sh <name> <client_port> <server_port> [db_port]
#   name         snake_case phase name WITHOUT the _component suffix (e.g. peer_review)
#   client_port  unique dev port for the micro-frontend (e.g. 3011)
#   server_port  unique port for the Go service (e.g. 8089)
#   db_port      unique published Postgres port (default: next free 54xx in docker-compose.yml)
#
# Copies clients/example_component and servers/example_server, renames every identifier,
# and registers the phase in the workspaces, core client, docker-compose, env templates,
# Makefile, and CI matrices. Prints the remaining manual steps at the end.
# Full documentation: docs/contributor/new_course_phase.md
#
# Run on a clean working tree. If a step fails midway, revert with:
#   git checkout -- . && git clean -fd clients/<name>_component servers/<name>

set -euo pipefail

die() { echo "error: $*" >&2; exit 1; }

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"
[[ -d clients/example_component && -d servers/example_server ]] \
  || die "run from the PROMPT repo (example component/server not found)"

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

# Derived spellings
KEBAB="${NAME//_/-}"
UPPER="$(echo "$NAME" | tr '[:lower:]' '[:upper:]')"
PASCAL="$(echo "$NAME" | perl -pe 's/(^|_)([a-z])/\u$2/g')"
CAMEL="$(echo "$PASCAL" | perl -pe 's/^([A-Z])/\l$1/')"
READABLE="$(echo "$NAME" | perl -pe 's/_/ /g; s/(^| )([a-z])/$1\u$2/g')"

CLIENT_DIR="clients/${NAME}_component"
SERVER_DIR="servers/${NAME}"
[[ ! -e "$CLIENT_DIR" ]] || die "$CLIENT_DIR already exists"
[[ ! -e "$SERVER_DIR" ]] || die "$SERVER_DIR already exists"

grep -q "localhost:${CLIENT_PORT}\`" clients/core/rspack.config.mjs \
  && die "client port ${CLIENT_PORT} is already used by another remote"
grep -q "\"${SERVER_PORT}:8080\"" docker-compose.yml \
  && die "server port ${SERVER_PORT} is already used in docker-compose.yml"

if [[ -z "$DB_PORT" ]]; then
  DB_PORT="$(grep -o '"54[0-9][0-9]:5432"' docker-compose.yml | tr -d '"' | cut -d: -f1 | sort -n | tail -1)"
  DB_PORT=$((DB_PORT + 1))
fi
grep -q "\"${DB_PORT}:5432\"" docker-compose.yml \
  && die "db port ${DB_PORT} is already used in docker-compose.yml"

echo "Scaffolding phase '${NAME}' (client :${CLIENT_PORT}, server :${SERVER_PORT}, db :${DB_PORT})"

# rename_tokens <file...> — applies the example->NAME substitutions, most specific first
rename_tokens() {
  perl -pi -e "
    s/example_component/${NAME}_component/g;
    s/example_server/${NAME}/g;
    s/example_table/${NAME}_table/g;
    s/ExampleTable/${PASCAL}Table/g;
    s/example-service/${KEBAB}-service/g;
    s/example-root/${KEBAB}-root/g;
    s/EXAMPLE_SERVER/${UPPER}_SERVER/g;
    s/EXAMPLE_HOST/${UPPER}_HOST/g;
    s/DB_EXAMPLE_/DB_${UPPER}_/g;
    s/InitExampleModule/Init${PASCAL}Module/g;
    s/setupExampleRouter/setup${PASCAL}Router/g;
    s/ExampleService/${PASCAL}Service/g;
    s/ExampleServerConfigHandler/${PASCAL}ConfigHandler/g;
    s/ExampleServerCopyHandler/${PASCAL}CopyHandler/g;
    s/helloExampleServer/hello${PASCAL}Server/g;
    s/GetExampleInfo/Get${PASCAL}Info/g;
    s/getExampleInfo/get${PASCAL}Info/g;
    s/ExampleInfo/${PASCAL}Info/g;
    s/exampleInfo/${CAMEL}Info/g;
    s/exampleRouter/${CAMEL}Router/g;
    s/exampleServer/${CAMEL}Server/g;
    s/exampleAxiosInstance/${CAMEL}AxiosInstance/g;
    s/ExampleRoutes/${PASCAL}Routes/g;
    s/ExampleSidebar/${PASCAL}Sidebar/g;
    s/ExampleComponent/${PASCAL}Component/g;
    s/Example Component/${READABLE} Component/g;
    s/Example Server/${READABLE} Server/g;
    s/PROMPT Example API/PROMPT ${READABLE} API/g;
    s/\bExample\b/${PASCAL}/g;
    s/\bexample\b/${NAME}/g;
  " "$@"
}

# insert_after <file> <perl-regex matching a full line> <replacement text appended as new line>
# Duplicates nothing; adds one new line directly after the first matching line.
insert_after() {
  local file="$1" pattern="$2" newline="$3"
  MATCH_PATTERN="$pattern" NEW_LINE="$newline" perl -pi -e '
    BEGIN { $done = 0 }
    if (!$done && /$ENV{MATCH_PATTERN}/) { $_ .= "$ENV{NEW_LINE}\n"; $done = 1 }
  ' "$file"
  grep -qF "$newline" "$file" || die "failed to insert into $file (anchor: $pattern)"
}

# copy_block <file> <anchor-regex> — copies the block starting at the anchor line and
# ending before the next blank line, applies rename_tokens + port substitutions to the
# copy, and inserts it (preceded by a blank line) directly after the original block.
copy_block() {
  local file="$1" anchor="$2" tmp
  tmp="$(mktemp)"
  awk -v anchor="$anchor" '
    $0 ~ anchor { inblock = 1 }
    inblock && /^[[:space:]]*$/ { exit }
    inblock { print }
  ' "$file" > "$tmp"
  [[ -s "$tmp" ]] || die "block anchor not found in $file: $anchor"
  rename_tokens "$tmp"
  perl -pi -e "
    s/\"8086:8080\"/\"${SERVER_PORT}:8080\"/;
    s/\"3001:80\"/\"${CLIENT_PORT}:80\"/;
    s/\"5437:5432\"/\"${DB_PORT}:5432\"/;
  " "$tmp"
  BLOCK_FILE="$tmp" awk -v anchor="$anchor" '
    { print }
    $0 ~ anchor { inblock = 1 }
    inblock && /^[[:space:]]*$/ && !done {
      while ((getline line < ENVIRON["BLOCK_FILE"]) > 0) print line
      print ""
      done = 1; inblock = 0
    }
  ' "$file" > "${file}.new" && mv "${file}.new" "$file"
  rm -f "$tmp"
}

# ---------------------------------------------------------------- 1. client
echo "-> ${CLIENT_DIR}"
cp -R clients/example_component "$CLIENT_DIR"
rm -rf "$CLIENT_DIR/node_modules" "$CLIENT_DIR/build" "$CLIENT_DIR/README.md"
mv "$CLIENT_DIR/src/example_component" "$CLIENT_DIR/src/${NAME}_component"
mv "$CLIENT_DIR/src/${NAME}_component/network/exampleServerConfig.ts" \
   "$CLIENT_DIR/src/${NAME}_component/network/${CAMEL}ServerConfig.ts"
mv "$CLIENT_DIR/src/${NAME}_component/network/queries/getExampleInfo.ts" \
   "$CLIENT_DIR/src/${NAME}_component/network/queries/get${PASCAL}Info.ts"
find "$CLIENT_DIR" -type f \( -name '*.ts' -o -name '*.tsx' -o -name '*.mjs' \
  -o -name '*.json' -o -name '*.js' -o -name '*.html' -o -name 'Dockerfile' \) \
  -exec perl -0777 -pi -e 's/exampleServerConfig/'"${CAMEL}"'ServerConfig/g' {} + 2>/dev/null || true
find "$CLIENT_DIR" -type f \( -name '*.ts' -o -name '*.tsx' -o -name '*.mjs' \
  -o -name '*.json' -o -name '*.js' -o -name '*.html' -o -name 'Dockerfile' \) -print0 \
  | xargs -0 perl -pi -e "s/COMPONENT_DEV_PORT = 3001/COMPONENT_DEV_PORT = ${CLIENT_PORT}/"
find "$CLIENT_DIR" -type f \( -name '*.ts' -o -name '*.tsx' -o -name '*.mjs' \
  -o -name '*.json' -o -name '*.js' -o -name '*.html' -o -name 'Dockerfile' \) -print0 \
  | while IFS= read -r -d '' f; do rename_tokens "$f"; done

# ---------------------------------------------------------------- 2. server
echo "-> ${SERVER_DIR}"
cp -R servers/example_server "$SERVER_DIR"
rm -f "$SERVER_DIR/README.md"
mv "$SERVER_DIR/example" "$SERVER_DIR/${NAME}"
find "$SERVER_DIR" -type f \( -name '*.go' -o -name '*.sql' -o -name '*.yaml' \
  -o -name '*.json' -o -name '*.mod' -o -name '*.sum' \) -print0 \
  | while IFS= read -r -d '' f; do rename_tokens "$f"; done
perl -pi -e "s/localhost:8086/localhost:${SERVER_PORT}/g" "$SERVER_DIR/main.go" \
  "$SERVER_DIR/docs/docs.go" "$SERVER_DIR/docs/swagger.json" "$SERVER_DIR/docs/swagger.yaml"

# ------------------------------------------------------- 3. yarn workspaces
echo "-> workspaces"
perl -pi -e "s/^(\s*)\"example_component\",\$/\$1\"example_component\",\n\$1\"${NAME}_component\",/" \
  clients/package.json clients/lerna.json

# ------------------------------------------------------------- 4. core client
echo "-> core client registration"
RSPACK=clients/core/rspack.config.mjs
insert_after "$RSPACK" \
  'const exampleURL = ' \
  "  const ${CAMEL}URL = IS_DEV ? \`http://localhost:${CLIENT_PORT}\` : \`/${KEBAB}\`"
insert_after "$RSPACK" \
  'example_component: `example_component@' \
  "          ${NAME}_component: \`${NAME}_component@\${${CAMEL}URL}/remoteEntry.js?\${Date.now()}\`,"

PM=clients/core/src/managementConsole/PhaseMapping
for f in ExternalRoutes/ExampleRoutes.tsx ExternalSidebars/ExampleSidebar.tsx; do
  dst="$PM/${f/Example/$PASCAL}"
  cp "$PM/$f" "$dst"
  rename_tokens "$dst"
done
insert_after "$PM/PhaseRouterMapping.tsx" \
  "import \\{ ExampleRoutes \\}" \
  "import { ${PASCAL}Routes } from './ExternalRoutes/${PASCAL}Routes'"
insert_after "$PM/PhaseRouterMapping.tsx" \
  '  example_component: ExampleRoutes,' \
  "  ${NAME}_component: ${PASCAL}Routes,"
insert_after "$PM/PhaseSidebarMapping.tsx" \
  "import \\{ ExampleSidebar \\}" \
  "import { ${PASCAL}Sidebar } from './ExternalSidebars/${PASCAL}Sidebar'"
insert_after "$PM/PhaseSidebarMapping.tsx" \
  '  example_component: ExampleSidebar,' \
  "  ${NAME}_component: ${PASCAL}Sidebar,"
insert_after "$PM/PhaseStudentDetailMapping.tsx" \
  '  example_component: Fallback,' \
  "  ${NAME}_component: Fallback,"

APP=clients/core/src/App.tsx
insert_after "$APP" \
  "import \\{ ExampleRoutes \\}" \
  "import { ${PASCAL}Routes } from './managementConsole/PhaseMapping/ExternalRoutes/${PASCAL}Routes'"
ROUTE_BLOCK="$(perl -0777 -ne "print \$1 if /(\n            <Route\n              path='\/management\/course\/:courseId\/example_component\/\*'.*?\n            \/>)/s" "$APP")"
[[ -n "$ROUTE_BLOCK" ]] || die "example route block not found in $APP"
NEW_ROUTE="$(echo "$ROUTE_BLOCK" | perl -pe "s/example_component/${NAME}_component/; s/ExampleRoutes/${PASCAL}Routes/")"
ROUTE_OLD="$ROUTE_BLOCK" ROUTE_NEW="$NEW_ROUTE" perl -0777 -pi -e \
  's/\Q$ENV{ROUTE_OLD}\E/$ENV{ROUTE_OLD}$ENV{ROUTE_NEW}/' "$APP"

insert_after clients/core/public/env.js \
  "  EXAMPLE_HOST: 'http://localhost:8086'," \
  "  ${UPPER}_HOST: 'http://localhost:${SERVER_PORT}',"
insert_after clients/core/public/env.template.js \
  "  EXAMPLE_HOST: '\\\$EXAMPLE_HOST'," \
  "  ${UPPER}_HOST: '\$${UPPER}_HOST',"

# ------------------------------------------------------------ 5. docker-compose
echo "-> docker-compose.yml"
copy_block docker-compose.yml '^  server-example:$'
copy_block docker-compose.yml '^  client-example-component:$'
copy_block docker-compose.yml '^  db-example-server:$'
insert_after docker-compose.yml '^      - EXAMPLE_HOST$' "      - ${UPPER}_HOST"

# ------------------------------------------------------------- 6. env templates
echo "-> env templates"
cat >> .env.template <<EOF

# ============================================================================
# ${UPPER} SERVER DATABASE CONFIGURATION
# ============================================================================

DB_HOST_${UPPER}_SERVER=db-${KEBAB}
DB_PORT_${UPPER}_SERVER=${DB_PORT}
DB_${UPPER}_NAME=prompt
DB_${UPPER}_USER=prompt-postgres
DB_${UPPER}_PASSWORD=prompt-postgres
${UPPER}_HOST=http://localhost:${SERVER_PORT}
SENTRY_DSN_${UPPER}_SERVER=
EOF
cat >> .env.dev.template <<EOF

# ${READABLE}
DB_HOST_${UPPER}_SERVER=localhost
DB_PORT_${UPPER}_SERVER=${DB_PORT}
EOF

# ------------------------------------------------------------------ 7. Makefile
echo "-> Makefile"
perl -pi -e "s/\bserver-example\b/server-example server-${KEBAB}/ if \$. < 15" Makefile
perl -pi -e "s/\btest-example\b/test-example test-${KEBAB}/ if \$. < 15" Makefile
perl -pi -e "s/\bsqlc-example\b/sqlc-example sqlc-${KEBAB}/ if \$. < 15" Makefile
insert_after Makefile '^\t@\$\(MAKE\) server-example &$' "	@\$(MAKE) server-${KEBAB} &"
insert_after Makefile '^\tcd servers/example_server && go vet' "	cd servers/${NAME} && go vet ./..."
perl -pi -e "s/^(test: .*test-example)/\$1 test-${KEBAB}/" Makefile
perl -pi -e "s/^(sqlc: .*sqlc-example)/\$1 sqlc-${KEBAB}/" Makefile
printf '\nserver-%s: ## Start %s server (port %s)\n\tcd servers/%s && go run main.go\n' \
  "$KEBAB" "$NAME" "$SERVER_PORT" "$NAME" >> Makefile
printf '\ntest-%s: ## Run %s server tests\n\tcd servers/%s && go test ./...\n' \
  "$KEBAB" "$NAME" "$NAME" >> Makefile
printf '\nsqlc-%s: ## Generate sqlc code for %s server\n\tcd servers/%s && sqlc generate\n' \
  "$KEBAB" "$NAME" "$NAME" >> Makefile

# ------------------------------------------------------------------------ 8. CI
echo "-> CI matrices"
insert_after .github/workflows/quality-clients.yml '^          - example_component$' \
  "          - ${NAME}_component"
insert_after .github/workflows/quality-servers.yml '^          - example_server$' \
  "          - ${NAME}"
insert_after .github/workflows/test-servers.yml '^          - example_server$' \
  "          - ${NAME}"

# ---------------------------------------------------------------- 9. verify
echo "-> verifying"
(cd "$SERVER_DIR" && go build ./...) || die "go build failed in $SERVER_DIR"
(cd clients && yarn install --mode update-lockfile >/dev/null 2>&1 && yarn install >/dev/null 2>&1) \
  || die "yarn install failed"
(cd clients && yarn tsc -p "${NAME}_component/tsconfig.json" --noEmit --pretty false) \
  || die "typecheck failed for ${NAME}_component"
(cd clients && yarn biome check --write "${NAME}_component" core >/dev/null 2>&1) || true
(cd clients && yarn biome check --diagnostic-level=error "${NAME}_component" >/dev/null) \
  || die "biome check failed for ${NAME}_component"

cat <<EOF

Done. Phase '${NAME}' scaffolded and registered.

Remaining manual steps (see docs/contributor/new_course_phase.md):
  1. Course phase type: add init${PASCAL}() in
     servers/core/coursePhaseType/initializeTypes.go (name key: '${NAME}_component').
  2. Deployment: docker-compose.prod.yml service + traefik labels, and the
     build-and-push-clients.yml / deploy-docker.yml / dev.yml / prod.yml wiring.
  3. Optional: add ${UPPER}_HOST to the EnvType in @tumaet/prompt-shared-state.
  4. Implement your schema (db/migration, db/query, 'make sqlc-${KEBAB}') and logic.

Try it: make db && make server-${KEBAB}   and   cd clients && yarn dev
EOF
