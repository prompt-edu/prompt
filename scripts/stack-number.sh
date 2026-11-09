#!/bin/sh
# Print the number of this checkout's local stack.
# Usage: scripts/stack-number.sh [--peek | --list | --release <n>]
# Full documentation: docs/contributor/setup.md ("Run several stacks side by side")
#
# The main checkout is stack 0. A linked worktree gets the first free number in
# 1..STACK_MAX on first use, recorded as "<n> <worktree path>" lines in
# <git common dir>/prompt-stacks, so every worktree of the clone sees the same
# allocation. Entries whose worktree git no longer lists are freed on the next
# allocation, and `make stack-down` frees its stack's number with --release.
#
# PROMPT_STACK=<n> overrides the allocation (and is not recorded). It refuses a
# number that another live worktree holds, unless PROMPT_STACK_FORCE=1.
#
# --peek prints the number without allocating one, or "none" if this worktree
# has none yet.
# --list prints the allocations, and the numbers whose removed worktree left
# containers or volumes behind.
# --release <n> frees number n.
set -eu

# Stack n shifts every host port by n * 100. Above 7 the shifted ports start to
# collide (stack 8's certificate server would take stack 0's SeaweedFS filer
# port 8888), so stop there.
STACK_MAX=7

# How long to wait for another run's lock, in tenths of a second.
LOCK_TIMEOUT=100

die() {
    echo "stack-number: $*" >&2
    exit 1
}

mode="${1:-}"
case "$mode" in
    "" | --peek | --list) ;;
    --release)
        case "${2:-}" in
            "" | *[!0-9]*) die "--release needs a stack number (got: ${2:-nothing})" ;;
        esac
        ;;
    *) die "unknown option $mode (expected --peek, --list or --release <n>)" ;;
esac

# Outside a git checkout (an exported tarball, say) there is only one stack.
if top="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    git_dir="$(git rev-parse --path-format=absolute --git-dir)"
    common_dir="$(git rev-parse --path-format=absolute --git-common-dir)"
    registry="$common_dir/prompt-stacks"
    lock="$registry.lock"
    worktrees="$(git worktree list --porcelain | sed -n 's/^worktree //p')"
    main_checkout="$(printf '%s\n' "$worktrees" | head -n 1)"
else
    top=""
fi

# The registry entries whose worktree still exists.
live_entries() {
    [ -f "$registry" ] || return 0
    while read -r n path; do
        [ -n "$n" ] || continue
        if printf '%s\n' "$worktrees" | grep -Fxq -- "$path"; then
            printf '%s %s\n' "$n" "$path"
        fi
    done < "$registry"
}

# The live worktree that holds stack number $1, if any.
holder() {
    if [ "$1" = 0 ]; then
        printf '%s\n' "$main_checkout"
        return
    fi
    live_entries | awk -v n="$1" '$1 == n { sub(/^[^ ]+ /, ""); print; exit }'
}

# A removed worktree's containers and volumes outlive it until someone runs
# `make stack-down PROMPT_STACK=<n>`. Handing its number to a new worktree would
# start that worktree on another branch's data, so such a number stays taken.
leftovers() {
    command -v docker >/dev/null 2>&1 || return 1
    for project in "prompt-$1" "prompt-$1-e2e"; do
        [ -n "$(docker ps -aq --filter "label=com.docker.compose.project=$project" 2>/dev/null)" ] && return 0
        [ -n "$(docker volume ls -q --filter "label=com.docker.compose.project=$project" 2>/dev/null)" ] && return 0
    done
    return 1
}

docker_running() {
    command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1
}

# Serialize registry writes: parallel make runs in two fresh worktrees must not
# pick the same number. The lock is a symlink to the holder's pid: creating it
# is atomic and records the pid in the same step, so a lock always names its
# holder. A lock whose holder pid is gone is left over from a killed run and is
# broken.
take_lock() {
    [ -w "$common_dir" ] \
        || die "can't write the stack registry in $common_dir. Set PROMPT_STACK=<n> to pick a number without it."
    tries=0
    until ln -s "$$" "$lock" 2>/dev/null; do
        lock_pid="$(readlink "$lock" 2>/dev/null || true)"
        if [ -n "$lock_pid" ] && ! kill -0 "$lock_pid" 2>/dev/null; then
            rm -f "$lock"
        fi
        tries=$((tries + 1))
        [ "$tries" -lt "$LOCK_TIMEOUT" ] \
            || die "timed out waiting for $lock (held by pid ${lock_pid:-unknown}). If no other make is running, remove that file."
        sleep 0.1
    done
    trap 'rm -f "$lock"' EXIT INT TERM
}

if [ "$mode" = "--list" ]; then
    [ -n "$top" ] || exit 0
    entries="$(live_entries)"
    [ -z "$entries" ] || printf '%s\n' "$entries"
    docker_running || exit 0
    n=1
    while [ "$n" -le "$STACK_MAX" ]; do
        if ! printf '%s\n' "$entries" | cut -d' ' -f1 | grep -Fxq "$n" && leftovers "$n"; then
            echo "$n (worktree removed, containers or volumes left: make stack-down PROMPT_STACK=$n)"
        fi
        n=$((n + 1))
    done
    exit 0
fi

if [ "$mode" = "--release" ]; then
    { [ -n "$top" ] && [ -f "$registry" ]; } || exit 0
    take_lock
    awk -v n="$2" '$1 != n' "$registry" > "$registry.tmp"
    mv "$registry.tmp" "$registry"
    exit 0
fi

if [ -n "${PROMPT_STACK:-}" ]; then
    case "$PROMPT_STACK" in
        *[!0-9]*) die "PROMPT_STACK must be a number between 0 and $STACK_MAX (got: $PROMPT_STACK)" ;;
    esac
    [ "$PROMPT_STACK" -le "$STACK_MAX" ] \
        || die "PROMPT_STACK must be a number between 0 and $STACK_MAX (got: $PROMPT_STACK)"
    # Strip leading zeros, so 01 and 1 name the same stack.
    n=$((PROMPT_STACK + 0))
    if [ -n "$top" ] && [ "${PROMPT_STACK_FORCE:-}" != 1 ]; then
        owner="$(holder "$n")"
        if [ -n "$owner" ] && [ "$owner" != "$top" ]; then
            die "stack $n belongs to the worktree $owner. Run make there, or add PROMPT_STACK_FORCE=1 to use that stack from here anyway."
        fi
    fi
    echo "$n"
    exit 0
fi

[ -n "$top" ] || { echo 0; exit 0; }

# The main checkout's git dir is the common dir itself.
if [ "$git_dir" = "$common_dir" ]; then
    echo 0
    exit 0
fi

if [ "$mode" = "--peek" ]; then
    assigned="$(live_entries | awk -v top="$top" '{ n = $1; sub(/^[^ ]+ /, "") } $0 == top { print n; exit }')"
    echo "${assigned:-none}"
    exit 0
fi

take_lock

kept=""
assigned=""
if [ -f "$registry" ]; then
    while read -r n path; do
        [ -n "$n" ] || continue
        # Drop entries of removed worktrees, which frees their number.
        printf '%s\n' "$worktrees" | grep -Fxq -- "$path" || continue
        kept="$kept$n $path
"
        [ "$path" = "$top" ] && assigned="$n"
    done < "$registry"
fi

if [ -z "$assigned" ]; then
    # Without a running daemon, docker ps and docker volume ls print nothing, so
    # leftovers() would report every number as free.
    if command -v docker >/dev/null 2>&1 && ! docker_running; then
        die "Docker is not running, so free stack numbers can't be checked for leftovers. Start Docker, or set PROMPT_STACK=<n>."
    fi
    n=1
    skipped=""
    while [ "$n" -le "$STACK_MAX" ]; do
        if ! printf '%s' "$kept" | cut -d' ' -f1 | grep -Fxq "$n"; then
            leftovers "$n" || break
            skipped="$skipped $n"
        fi
        n=$((n + 1))
    done
    if [ "$n" -gt "$STACK_MAX" ]; then
        hint=""
        [ -n "$skipped" ] && hint=" Stacks$skipped still have containers or volumes of a removed worktree: free one with 'make stack-down PROMPT_STACK=<n>'."
        die "all stacks 1-$STACK_MAX are taken by other worktrees (see make stack).$hint Or run 'make stack-down' in a worktree whose stack you no longer need."
    fi
    assigned="$n"
    kept="$kept$n $top
"
fi

printf '%s' "$kept" > "$registry.tmp"
mv "$registry.tmp" "$registry"
echo "$assigned"
