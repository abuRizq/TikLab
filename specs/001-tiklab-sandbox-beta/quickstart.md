# Quickstart: TikLab Sandbox Beta

**Branch**: `001-tiklab-sandbox-beta` | **Date**: 2026-03-16

## Prerequisites

- Docker installed and running
- `tiklab` binary on your PATH

## 1. Create and Start

```bash
tiklab create
tiklab start
```

Wait for "Ready." — the sandbox is now running RouterOS v7 with 50 simulated users.

## 2. Connect

**SSH** (default credentials: `admin` / no password):
```bash
ssh admin@localhost -p 2222
```

**RouterOS API** (port 8728):
```bash
# From your billing system or API client, connect to:
# Host: localhost, Port: 8728, User: admin, Password: (empty)
```

**Winbox** (port 8291):
Open Winbox and connect to `localhost:8291`.

## 3. Explore

Once connected via SSH, try these RouterOS v7 commands:

```
# List active Hotspot users
/ip/hotspot/active/print

# View DHCP leases
/ip/dhcp-server/lease/print

# Check bandwidth queues
/queue/simple/print

# View interface traffic
/interface/print stats
```

## 4. Scale

Increase simulated users to stress-test your integration:

```bash
tiklab scale 200
```

Scale back down:

```bash
tiklab scale 50
```

## 5. Reset

Wipe all changes and start fresh (without recreating the environment):

```bash
tiklab reset
```

## 6. Clean Up

Remove the sandbox entirely:

```bash
tiklab destroy
```

## Port Reference

| Service | Host Port | Container Port |
|---------|-----------|---------------|
| SSH | 2222 | 22 |
| RouterOS API | 8728 | 8728 |
| Winbox | 8291 | 8291 |
