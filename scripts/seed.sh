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
SEED_DB_PASSWORD="${SEED_DB_PASSWORD:-prompt-postgres}"
SEED_TIMEOUT="${SEED_TIMEOUT:-300}"

# The eight phase type names seed/core.sql resolves by name. Core creates them
# on startup: the first seven in InitCoursePhaseTypeModule, Application later in
# InitApplicationAdministrationModule, both before the router starts serving.
CORE_PHASE_TYPES="'Application','Interview','Matching','Team Allocation','Self Team Allocation','Assessment','Certificate','Presentation'"

# Every phase seed points at core-owned ids (course phases f……, participations
# cd……) that only seed/core.sql creates, and nothing enforces those links across
# databases. core.sql applies in one transaction, so the demo course is present
# exactly when the rest of what it owns is.
DEMO_COURSE="c0000002-0000-0000-0000-000000000002"

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

# Every database carries its own DB_<SERVICE>_* credentials in .env, so each
# connection setting is overridable per service and only falls back to the core
# values when nothing more specific is set.
svc_setting() {
    key="$(echo "$1" | tr '[:lower:]' '[:upper:]')"
    eval "printf '%s' \"\${SEED_${key}_$2:-\$3}\""
}

svc_host() { svc_setting "$1" HOST localhost; }
svc_port() { svc_setting "$1" PORT "$(default_port "$1")"; }
svc_user() { svc_setting "$1" USER "$SEED_DB_USER"; }
svc_password() { svc_setting "$1" PASSWORD "$SEED_DB_PASSWORD"; }
svc_name() { svc_setting "$1" NAME "$SEED_DB_NAME"; }

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
    PGPASSWORD="$(svc_password "$1")" psql -X -A -t -q \
        -h "$(svc_host "$1")" -p "$(svc_port "$1")" -U "$(svc_user "$1")" \
        -d "$(svc_name "$1")" -c "$2" 2>/dev/null || true
}

wait_for() {
    want="$(expected_version "$1")"
    waited=0
    while [ "$waited" -lt "$SEED_TIMEOUT" ]; do
        if [ "$(query "$1" "SELECT version FROM schema_migrations WHERE NOT dirty")" = "$want" ]; then
            if [ "$1" != core ]; then
                return 0
            fi
            if [ "$(query "$1" "SELECT count(*) FROM course_phase_type WHERE name IN ($CORE_PHASE_TYPES)")" = 8 ]; then
                return 0
            fi
        fi
        waited=$((waited + 2))
        sleep 2
    done
    die "$1 at $(svc_host "$1"):$(svc_port "$1") is not at migration $want with its phase types initialised after ${SEED_TIMEOUT}s.
     The servers own the schema - start them once (make servers) so their startup migrations run."
}

require_core_seeded() {
    [ "$(query core "SELECT count(*) FROM course WHERE id = '$DEMO_COURSE'")" = 1 ] ||
        die "core at $(svc_host core):$(svc_port core) does not hold the demo course.
     A phase seed references core-owned ids - seed core first (make seed, or add core to the targets)."
}

apply() {
    file="$SEED_SQL_ROOT/$1.sql"
    [ -f "$file" ] || die "no seed file at $file"
    PGPASSWORD="$(svc_password "$1")" psql -X -v ON_ERROR_STOP=1 --single-transaction -q \
        -h "$(svc_host "$1")" -p "$(svc_port "$1")" -U "$(svc_user "$1")" \
        -d "$(svc_name "$1")" -f "$file"
}

targets="${*:-$SERVICES}"
for svc in $targets; do
    default_port "$svc" >/dev/null
done

# core.sql creates what every other file references, so it is applied first
# whatever order the targets arrive in.
core=""
phases=""
for svc in $targets; do
    if [ "$svc" = core ]; then core="core"; else phases="$phases $svc"; fi
done
targets="$core$phases"

# Preflight every database before touching any of them: each file runs in its own
# transaction, so a late failure would leave some databases reseeded and others not.
echo "seed: waiting for schemas"
for svc in $targets; do
    wait_for "$svc"
done

[ -n "$core" ] || require_core_seeded

for svc in $targets; do
    echo "seed: $svc"
    apply "$svc"
done

echo "seed: done"
