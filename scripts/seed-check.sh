#!/bin/sh
# Verify the cross-database references in seed/ resolve.
# Usage: scripts/seed-check.sh
# Full documentation: docs/contributor/guide/seeding.md
#
# Each service has its own database and the links between them are bare UUID
# values that no foreign key enforces, so a typo in a course_phase_id only shows
# up as an empty screen. This checks statically that every core-owned id a phase
# seed references is one seed/core.sql actually creates.
set -eu

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SEED_DIR="$REPO_ROOT/seed"
CORE="$SEED_DIR/core.sql"

# Demo course phases (f……), participations (cd……) and the shared students that
# the phase seeds point at.
CORE_OWNED='(f[0-9a-f]{7}|cd[0-9a-f]{6}|e[01][0-9a-f]{6}|a5000007)-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'

[ -f "$CORE" ] || { echo "seed-check: no $CORE" >&2; exit 1; }

status=0
for file in "$SEED_DIR"/*.sql; do
    [ "$file" = "$CORE" ] && continue
    for id in $(grep -Eo "$CORE_OWNED" "$file" | sort -u); do
        if ! grep -q "$id" "$CORE"; then
            echo "seed-check: $(basename "$file") references $id, which seed/core.sql does not create" >&2
            status=1
        fi
    done
done

[ "$status" -eq 0 ] && echo "seed-check: cross-database references OK"
exit "$status"
