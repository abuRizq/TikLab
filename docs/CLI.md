# MikroTik Sandbox CLI - Command Reference (Beta)

## Overview

The sandbox CLI provides five commands for managing a local MikroTik CHR environment with synthetic traffic.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Enable verbose output |
| `--skip-prereqs` | | Skip prerequisite checks (for testing only) |

## Commands

### `sandbox create`

Initializes the default `isp_small` synthetic environment.

**Inputs:** None

**Flags:**
- `--force` — Overwrite existing sandbox (stops containers, tears down, recreates)

**Outputs:** Success message when the environment is created.

**Failure states:**
- Sandbox already exists (use `--force` to overwrite) → exit 1
- Missing prerequisites (Docker, QEMU, Linux) → exit 1
- Work directory creation failure → exit 1

---

### `sandbox start`

Boots the Docker Compose stack, QEMU CHR, and Traffic Engine.

**Inputs:** None (requires prior `sandbox create`)

**Outputs:** Success message with access ports (API: 8728, Winbox: 8291, SSH: 2222)

**Failure states:**
- Missing prerequisites → exit 1
- Environment not created → exit 1
- Docker/Compose failure → exit 1

---

### `sandbox reset`

Stops CHR, drops and recreates the QCOW2 overlay, then restarts. Target: < 2 seconds.

**Inputs:** None

**Outputs:** Success message.

**Failure states:**
- Missing prerequisites → exit 1
- Overlay disk access failure → exit 1
- Compose up failure after overlay reset → exit 1

---

### `sandbox destroy`

Tears down the environment and cleans up all resources.

**Inputs:** None

**Flags:**
- `--all` — Also remove workdir (base image and all state)

**Outputs:** Success message when teardown completes.

**Failure states:**
- Missing prerequisites → exit 1
- Partial teardown (containers, namespaces) → exit 1

---

### `sandbox scale-users <count>`

Dynamically adds or removes active simulation namespaces (Linux only).

**Inputs:**
- `count` (required): Non-negative integer, target number of active users

**Outputs:** Success message with new user count.

**Failure states:**
- Invalid count (non-integer or negative) → exit 1
- Missing prerequisites → exit 1
- Not Linux (network namespaces required) → exit 1

---

## Prerequisites

- **Docker:** Installed and running
- **QEMU:** `qemu-system-x86_64` in PATH
- **Linux:** Required for network namespaces (full sandbox)

On non-Linux hosts, use `--skip-prereqs` for CLI testing only; the full sandbox will not run.

---

### `sandbox validate`

Runs beta validation checks: port connectivity, API CRUD, reset performance.

**Inputs:** None (requires running sandbox for full validation)

**Flags:**
- `--skip-reset` — Skip reset performance measurement (avoids resetting the sandbox)

**Outputs:** Port status, API status, reset timing.

**Failure states:**
- Port checks fail (sandbox not running) → exit 1
