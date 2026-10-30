#!/bin/sh
# Print the number of this checkout's local stack.
# Usage: scripts/stack-number.sh [--list]
# Full documentation: docs/contributor/setup.md ("Run several stacks side by side")
#
# The main checkout is stack 0. A linked worktree gets the first free number in
# 1..STACK_MAX on first use, recorded as "<n> <worktree path>" lines in
# <git common dir>/prompt-stacks, so every worktree of the clone sees the same
# allocation. Entries whose worktree git no longer lists are freed on the next
# allocation. PROMPT_STACK=<n> overrides the allocation (and is not recorded).
#
# --list prints the recorded allocations instead.
set -eu

# Stack n shifts every host port by n * 100. Above 7 the shifted ports start to
# collide (stack 8's certificate server would take stack 0's SeaweedFS filer
# port 8888), so stop there.
STACK_MAX=7

die() {
    echo "stack-number: $*" >&2
    exit 1
}

if [ "${1:-}" = "--list" ]; then
    common_dir="$(git rev-parse --path-format=absolute --git-common-dir 2>/dev/null)" || exit 0
    [ ! -f "$common_dir/prompt-stacks" ] || cat "$common_dir/prompt-stacks"
    exit 0
fi

if [ -n "${PROMPT_STACK:-}" ]; then
    case "$PROMPT_STACK" in
        *[!0-9]*) die "PROMPT_STACK must be a number between 0 and $STACK_MAX (got: $PROMPT_STACK)" ;;
    esac
    [ "$PROMPT_STACK" -le "$STACK_MAX" ] \
        || die "PROMPT_STACK must be a number between 0 and $STACK_MAX (got: $PROMPT_STACK)"
    # Strip leading zeros, so 01 and 1 name the same stack.
    echo $((PROMPT_STACK + 0))
    exit 0
fi

# Outside a git checkout (an exported tarball, say) there is only one stack.
top="$(git rev-parse --show-toplevel 2>/dev/null)" || { echo 0; exit 0; }
git_dir="$(git rev-parse --path-format=absolute --git-dir)"
common_dir="$(git rev-parse --path-format=absolute --git-common-dir)"
registry="$common_dir/prompt-stacks"

# The main checkout's git dir is the common dir itself.
if [ "$git_dir" = "$common_dir" ]; then
    echo 0
    exit 0
fi

# Serialize allocations: parallel make runs in two fresh worktrees must not pick
# the same number. The lock is a symlink to the holder's pid: creating it is
# atomic and records the pid in the same step, so a lock always names its
# holder. A lock whose holder pid is gone is left over from a killed run and is
# broken.
lock="$registry.lock"
until ln -s "$$" "$lock" 2>/dev/null; do
    holder="$(readlink "$lock" 2>/dev/null || true)"
    if [ -n "$holder" ] && ! kill -0 "$holder" 2>/dev/null; then
        rm -f "$lock"
    fi
    sleep 0.1
done
trap 'rm -f "$lock"' EXIT INT TERM

live="$(git worktree list --porcelain | sed -n 's/^worktree //p')"
kept=""
assigned=""
if [ -f "$registry" ]; then
    while read -r n path; do
        [ -n "$n" ] || continue
        # Drop entries of removed worktrees, which frees their number.
        printf '%s\n' "$live" | grep -Fxq -- "$path" || continue
        kept="$kept$n $path
"
        [ "$path" = "$top" ] && assigned="$n"
    done < "$registry"
fi

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

if [ -z "$assigned" ]; then
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
        die "all stacks 1-$STACK_MAX are taken by other worktrees (see $registry).$hint Or remove a worktree you no longer need, or set PROMPT_STACK=<n>."
    fi
    assigned="$n"
    kept="$kept$n $top
"
fi

printf '%s' "$kept" > "$registry.tmp"
mv "$registry.tmp" "$registry"
echo "$assigned"
