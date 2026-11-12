#!/bin/sh
# Stop the servers and client dev servers that `make servers` and `make clients`
# started for a stack.
# Usage: scripts/stack-stop.sh <port offset> <stack 0 port>...
# Full documentation: docs/contributor/setup.md ("Run several stacks side by side")
#
# Both targets run their processes in the background, so Ctrl-C doesn't reach
# them. Whatever listens on one of the stack's server or client dev ports is
# stopped, except the port forwarding of the container runtime, which holds the
# ports a container of the stack publishes.
set -eu

[ $# -ge 1 ] || { echo "usage: $0 <port offset> <stack 0 port>..." >&2; exit 2; }
offset="$1"
shift

if ! command -v lsof >/dev/null 2>&1; then
    echo "stack-stop: lsof is not installed, so the stack's processes can't be found by their ports." >&2
    exit 1
fi

ports=""
for port in "$@"; do
    ports="${ports:+$ports,}$((port + offset))"
done

# -F pcn prints each process as p<pid> and c<command>, then n<address:port>
# for every listening socket it holds; awk turns that into "<pid> <port> <command>".
listeners="$(lsof -nP -iTCP:"$ports" -sTCP:LISTEN -F pcn 2>/dev/null || true)"
pids=""
while read -r pid port name; do
    [ -n "$pid" ] || continue
    case "$name" in
        *[Dd]ocker* | *vpnkit* | *OrbStack* | *limactl* | *ssh*) continue ;;
    esac
    case " $pids " in
        *" $pid "*) continue ;;
    esac
    echo "stopping $name (pid $pid) on port $port"
    pids="$pids $pid"
done <<EOF
$(printf '%s\n' "$listeners" | awk '
    /^p/ { pid = substr($0, 2) }
    /^c/ { name = substr($0, 2) }
    /^n/ { port = $0; sub(/.*:/, "", port); print pid, port, name }')
EOF

if [ -z "$pids" ]; then
    echo "No server or client dev server of this stack is running."
    exit 0
fi
# shellcheck disable=SC2086 # one argument per pid
kill $pids
