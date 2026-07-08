#!/usr/bin/env bash
#
# Initialize a course phase created from prompt-course-phase-template.
#
# Usage: ./init.sh <name> [client_port] [server_port]
#   name         snake_case phase name (e.g. peer_review)
#   client_port  dev port for the micro-frontend (default 3011)
#   server_port  port for the Go service (default 8089)
#
# Renames every `example` placeholder in client/ and server/ to your phase name,
# then verifies the result builds. Run once, right after "Use this template".
#
# Afterwards, register the phase in the PROMPT core (one small PR to
# prompt-edu/prompt) — see docs/contributor/new_course_phase.md there.

set -euo pipefail

die() { echo "error: $*" >&2; exit 1; }

cd "$(dirname "${BASH_SOURCE[0]}")"
[[ -d client && -d server ]] || die "run from the repository root (client/ and server/ not found)"

NAME="${1:-}"
CLIENT_PORT="${2:-3011}"
SERVER_PORT="${3:-8089}"

[[ -n "$NAME" ]] || die "usage: $0 <name> [client_port] [server_port]"
[[ "$NAME" =~ ^[a-z][a-z0-9_]*$ ]] || die "name must be snake_case (got: $NAME)"
[[ "$NAME" != "example" ]] || die "pick a name other than 'example'"

KEBAB="${NAME//_/-}"
UPPER="$(echo "$NAME" | tr '[:lower:]' '[:upper:]')"
PASCAL="$(echo "$NAME" | perl -pe 's/(^|_)([a-z])/\u$2/g')"
CAMEL="$(echo "$PASCAL" | perl -pe 's/^([A-Z])/\l$1/')"
READABLE="$(echo "$NAME" | perl -pe 's/_/ /g; s/(^| )([a-z])/$1\u$2/g')"

echo "Initializing phase '${NAME}' (client :${CLIENT_PORT}, server :${SERVER_PORT})"

rename_tokens() {
  perl -pi -e "
    s/example_component/${NAME}_component/g;
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
    s/exampleServerConfig/${CAMEL}ServerConfig/g;
    s/exampleServer/${CAMEL}Server/g;
    s/exampleAxiosInstance/${CAMEL}AxiosInstance/g;
    s/ExampleComponent/${PASCAL}Component/g;
    s/Example Component/${READABLE} Component/g;
    s/Example Server/${READABLE} Server/g;
    s/PROMPT Example API/PROMPT ${READABLE} API/g;
    s/\bExample\b/${PASCAL}/g;
    s/\bexample\b/${NAME}/g;
  " "$@"
}

# Client
mv client/src/example_component "client/src/${NAME}_component"
mv "client/src/${NAME}_component/network/exampleServerConfig.ts" \
   "client/src/${NAME}_component/network/${CAMEL}ServerConfig.ts"
mv "client/src/${NAME}_component/network/queries/getExampleInfo.ts" \
   "client/src/${NAME}_component/network/queries/get${PASCAL}Info.ts"
find client -type f \( -name '*.ts' -o -name '*.tsx' -o -name '*.mjs' -o -name '*.js' \
  -o -name '*.json' -o -name '*.html' -o -name 'Dockerfile' -o -name '*.conf' \) \
  -not -path '*/node_modules/*' -not -path '*/build/*' -print0 \
  | while IFS= read -r -d '' f; do rename_tokens "$f"; done
perl -pi -e "s/COMPONENT_DEV_PORT = 3001/COMPONENT_DEV_PORT = ${CLIENT_PORT}/" client/rspack.config.mjs

# Server
mv server/example "server/${NAME}"
find server -type f \( -name '*.go' -o -name '*.sql' -o -name '*.yaml' -o -name '*.json' \
  -o -name '*.mod' -o -name 'Dockerfile' \) -print0 \
  | while IFS= read -r -d '' f; do rename_tokens "$f"; done
find server -type f \( -name '*.go' -o -name '*.json' -o -name '*.yaml' \) -print0 \
  | xargs -0 perl -pi -e "s/localhost:8086/localhost:${SERVER_PORT}/g"

# Root files
rename_tokens docker-compose.yml docker-compose.prod.yml .env.template README.md 2>/dev/null || true
perl -pi -e "s/\"8086:8080\"/\"${SERVER_PORT}:8080\"/; s/\"3001:80\"/\"${CLIENT_PORT}:80\"/" \
  docker-compose.yml docker-compose.prod.yml 2>/dev/null || true

# Verify
echo "-> verifying"
(cd server && go mod tidy && go build ./...) || die "go build failed"
if command -v yarn >/dev/null; then
  (cd client && yarn install && yarn typecheck) || die "client typecheck failed"
else
  echo "warning: yarn not found — run 'cd client && yarn install && yarn typecheck' yourself"
fi

cat <<EOF

Done. Phase '${NAME}' initialized.

Next steps:
  1. Delete this script and commit.
  2. Implement your schema (server/db/migration, server/db/query, sqlc generate) and logic.
  3. Register the phase in PROMPT core — see docs/contributor/new_course_phase.md in
     https://github.com/prompt-edu/prompt (remote URL, ${UPPER}_HOST, phase mappings, phase type).
EOF
