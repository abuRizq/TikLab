#!/bin/sh
# TikLab sandbox container entrypoint
# Delegates to base image for QEMU/RouterOS, then starts behavior engine.
# Env: TIKLAB_CONTROL_PORT (default 9090)

set -e

CONTROL_PORT="${TIKLAB_CONTROL_PORT:-9090}"

# Create a dummy eth1 so the base entrypoint detects two NICs and enables
# QEMU user-mode networking with hostfwd for port forwarding.
# We do NOT use a Docker network for this because multi-network containers
# break Docker Desktop's port forwarding on WSL2.
if ! ip link show eth1 >/dev/null 2>&1; then
    ip link add eth1 type dummy
    ip link set eth1 up
    ip addr add 192.168.88.1/24 dev eth1
fi

# Start QEMU/RouterOS via base image entrypoint
/routeros_source/entrypoint.sh &

# Kill udhcpd to prevent DHCP race between udhcpd and QEMU's SLiRP DHCP.
# With two NICs, RouterOS bridges them and its DHCP client sees both servers.
sleep 3
killall udhcpd 2>/dev/null || true

# Wait for RouterOS to boot; CLI also polls
sleep 27

# Start behavior engine when available (Phase 5)
if command -v tiklab-engine >/dev/null 2>&1; then
    export TIKLAB_CONTROL_PORT="$CONTROL_PORT"
    tiklab-engine &
fi

wait
