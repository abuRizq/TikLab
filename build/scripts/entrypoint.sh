#!/bin/sh
# TikLab sandbox container entrypoint
# Delegates to base image for QEMU/RouterOS, then starts behavior engine.
# Env: TIKLAB_CONTROL_PORT (default 9090)

set -e

CONTROL_PORT="${TIKLAB_CONTROL_PORT:-9090}"

# Start QEMU/RouterOS in background (base image CMD passed as args)
"$@" &

# Wait for RouterOS to boot; CLI also polls
sleep 30

# Start behavior engine when available (Phase 5)
if command -v tiklab-engine >/dev/null 2>&1; then
    export TIKLAB_CONTROL_PORT="$CONTROL_PORT"
    tiklab-engine &
fi

wait
