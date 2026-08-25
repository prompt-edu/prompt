#!/bin/sh
# Load the deterministic demo course into every PROMPT service database.
# Usage: scripts/seed.sh [service ...]
# Full documentation: docs/contributor/guide/seeding.md
#
# POSIX sh on purpose (the rest of scripts/ is bash): this also runs inside the
# e2e `seed` container, whose only shell is busybox ash.
set -eu

SERVICES="core interview team_allocation self_team_allocation assessment certificate presentation example_server"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SEED_MIGRATIONS_ROOT="${SEED_MIGRATIONS_ROOT:-$REPO_ROOT/servers}"
SEED_SQL_ROOT="${SEED_SQL_ROOT:-$REPO_ROOT/seed}"
SEED_DB_NAME="${SEED_DB_NAME:-prompt}"
SEED_DB_USER="${SEED_DB_USER:-prompt-postgres}"
SEED_TIMEOUT="${SEED_TIMEOUT:-300}"
PGPASSWORD="${SEED_DB_PASSWORD:-prompt-postgres}"
export PGPASSWORD

# The eight phase type names seed/core.sql resolves by name. Core creates them
# on startup: the first seven in InitCoursePhaseTypeModule, Application later in
# InitApplicationAdministrationModule, both before the router starts serving.
CORE_PHASE_TYPES="'Application','Interview','Matching','Team Allocation','Self Team Allocation','Assessment','Certificate','Presentation'"

die() { echo "seed: error: $*" >&2; exit 1; }

default_port() {
    case "$1" in
        core) echo 5432 ;;
        team_allocation) echo 5434 ;;
        assessment) echo 5435 ;;
        self_team_allocation) echo 5436 ;;
        example_server) echo 5437 ;;
        interview) echo 5438 ;;
        certificate) echo 5439 ;;
        presentation) echo 5440 ;;
        *) die "unknown service '$1'" ;;
    esac
}

svc_host() {
    key="$(echo "$1" | tr '[:lower:]' '[:upper:]')"
    eval "printf '%s' \"\${SEED_${key}_HOST:-localhost}\""
}

svc_port() {
    key="$(echo "$1" | tr '[:lower:]' '[:upper:]')"
    eval "printf '%s' \"\${SEED_${key}_PORT:-$(default_port "$1")}\""
}

# golang-migrate records the numeric prefix of the highest applied migration,
# unpadded (0028_x.up.sql -> 28). Deriving it from the source tree keeps the
# readiness gate exact without a committed version manifest that would drift.
expected_version() {
    dir="$SEED_MIGRATIONS_ROOT/$1/db/migration"
    [ -d "$dir" ] || die "no migration directory at $dir"
    version=""
    for file in "$dir"/*.up.sql; do
        [ -f "$file" ] || continue
        candidate="${file##*/}"
        candidate="${candidate%%_*}"
        candidate="$(printf '%s' "$candidate" | sed 's/^0*//')"
        [ -n "$candidate" ] || candidate=0
        if [ -z "$version" ] || [ "$candidate" -gt "$version" ]; then
            version="$candidate"
        fi
    done
    [ -n "$version" ] || die "no .up.sql migrations in $dir"
    echo "$version"
}

query() {
    psql -X -A -t -q -h "$1" -p "$2" -U "$SEED_DB_USER" -d "$SEED_DB_NAME" -c "$3" 2>/dev/null || true
}

wait_for() {
    svc="$1"; host="$2"; port="$3"
    want="$(expected_version "$svc")"
    waited=0
    while [ "$waited" -lt "$SEED_TIMEOUT" ]; do
        if [ "$(query "$host" "$port" "SELECT version FROM schema_migrations WHERE NOT dirty")" = "$want" ]; then
            if [ "$svc" != core ]; then
                return 0
            fi
            if [ "$(query "$host" "$port" "SELECT count(*) FROM course_phase_type WHERE name IN ($CORE_PHASE_TYPES)")" = 8 ]; then
                return 0
            fi
        fi
        waited=$((waited + 2))
        sleep 2
    done
    die "$svc at $host:$port is not at migration $want with its phase types initialised after ${SEED_TIMEOUT}s.
     The servers own the schema - start them once (make servers) so their startup migrations run."
}

apply() {
    file="$SEED_SQL_ROOT/$1.sql"
    [ -f "$file" ] || die "no seed file at $file"
    psql -X -v ON_ERROR_STOP=1 --single-transaction -q \
        -h "$2" -p "$3" -U "$SEED_DB_USER" -d "$SEED_DB_NAME" -f "$file"
}

targets="${*:-$SERVICES}"
for svc in $targets; do
    default_port "$svc" >/dev/null
done

# Preflight every database before touching any of them: each file runs in its own
# transaction, so a late failure would leave some databases reseeded and others not.
echo "seed: waiting for schemas"
for svc in $targets; do
    wait_for "$svc" "$(svc_host "$svc")" "$(svc_port "$svc")"
done

for svc in $targets; do
    echo "seed: $svc"
    apply "$svc" "$(svc_host "$svc")" "$(svc_port "$svc")"
done

echo "seed: done"
