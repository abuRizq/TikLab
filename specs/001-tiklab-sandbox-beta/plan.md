# Implementation Plan: TikLab Sandbox Beta

**Branch**: `001-tiklab-sandbox-beta` | **Date**: 2026-03-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-tiklab-sandbox-beta/spec.md`

## Summary

TikLab Sandbox Beta is a Go CLI tool that provisions and manages a Docker container running MikroTik RouterOS v7 CHR (via QEMU), populated with simulated users generating realistic network traffic. The CLI (`tiklab`) provides five lifecycle commands (`create`, `start`, `reset`, `destroy`, `scale`) while a behavior engine inside the container generates DHCP client activity, Hotspot sessions, and traffic conforming to three profiles (40% idle, 45% standard, 15% heavy). The architecture uses the RouterOS API for configuration and monitoring, Docker SDK for container lifecycle, and an in-container behavior engine for traffic simulation.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: `github.com/spf13/cobra` (CLI), `github.com/docker/docker/client` (Docker SDK), `github.com/go-routeros/routeros/v3` (RouterOS API)
**Storage**: Docker volumes for RouterOS disk state; local JSON file (`~/.tiklab/state.json`) for tracking the active sandbox instance
**Testing**: `go test` — unit tests for core logic, integration tests against a live Docker container with RouterOS
**Target Platform**: Windows, Linux, macOS (cross-compiled Go binary via `GOOS`/`GOARCH`)
**Project Type**: CLI tool + Docker image
**Performance Goals**: Create+Start < 2 minutes, Reset < 30 seconds, Scale 50→200 users < 60 seconds, RouterOS API remains responsive under 50 concurrent simulated users
**Constraints**: Max 500 simulated users, single sandbox instance at a time, Docker required on host, RouterOS CHR free license (1 Mbps per interface — sufficient for simulation)
**Scale/Scope**: 1–500 simulated users per sandbox instance

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Single-Command Delivery | PASS | Each CLI operation is a single `tiklab <verb>` invocation. No multi-step setup. Docker image is pre-built and pulled automatically. |
| II. Simulation Fidelity | PASS | Uses real MikroTik RouterOS v7 CHR image (not a mock). API, Winbox, SSH ports exposed natively. DHCP and Hotspot are actual RouterOS services. |
| III. Isolation & Stateless Reset | PASS | Docker container provides full isolation. Reset uses RouterOS API to wipe and re-apply configuration without container restart (fast path). Destroy removes all Docker artifacts. |
| IV. Go-Native Standalone Binary | PASS | CLI written in Go with cobra. Single binary, zero runtime dependencies beyond Docker. Cross-compiled for Windows/Linux/macOS. |
| V. CLI-First Distribution | PASS | CLI binary is the primary interface distributed via GitHub Releases. Docker image is a runtime dependency pulled automatically by `tiklab create`. CLI orchestrates container lifecycle through `create` → `start` sequence. Image tagged with semver. |

No violations. Gate passed.

## Project Structure

### Documentation (this feature)

```text
specs/001-tiklab-sandbox-beta/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli-schema.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
├── tiklab/
│   └── main.go
└── tiklab-engine/
    └── main.go

internal/
├── cli/
│   ├── root.go           # Root command, version, global flags
│   ├── create.go         # tiklab create
│   ├── start.go          # tiklab start
│   ├── reset.go          # tiklab reset
│   ├── destroy.go        # tiklab destroy
│   └── scale.go          # tiklab scale <count>
├── sandbox/
│   ├── manager.go        # Sandbox lifecycle state machine
│   └── state.go          # State persistence (~/.tiklab/state.json)
├── docker/
│   ├── client.go         # Docker SDK wrapper (connect, pull, health checks)
│   └── container.go      # Container create/start/stop/remove, port mapping, volumes
├── routeros/
│   ├── client.go         # RouterOS API client wrapper (connect, run, close)
│   └── config.go         # Initial configuration: DHCP server, Hotspot, queues, users
└── engine/
    ├── engine.go          # Behavior engine orchestrator (runs inside container)
    ├── user.go            # Simulated user: random identity, DHCP request, Hotspot auth
    ├── profiles.go        # Profile definitions: idle, standard, heavy — traffic parameters
    └── traffic.go         # Traffic generators per profile type

build/
├── Dockerfile            # Multi-stage: QEMU + RouterOS CHR + behavior engine
└── scripts/
    └── entrypoint.sh     # Container entrypoint: boot QEMU, wait for RouterOS, start engine

tests/
├── integration/
│   ├── lifecycle_test.go # Create → start → reset → destroy full cycle
│   ├── protocol_test.go  # API, SSH, Winbox port reachability and basic operations
│   └── traffic_test.go   # Verify user count, profile distribution, queue stats
└── unit/
    ├── sandbox_test.go   # State machine transitions
    ├── engine_test.go    # User generation, profile assignment
    └── config_test.go    # RouterOS configuration generation
```

**Structure Decision**: Single Go project at repository root. The `cmd/tiklab/` binary is the host-side CLI. The `cmd/tiklab-engine/` binary is the behavior engine entry point compiled and embedded in the Docker image via `build/Dockerfile`. All `internal/` packages except `internal/engine/` are used by the CLI; `internal/engine/` is shared by the engine binary. Tests are split between `unit/` (no Docker required) and `integration/` (requires running Docker).

## Architecture

### Component Overview

```text
┌─────────────────────────────────────────────────────────────┐
│  Host Machine                                               │
│                                                             │
│  ┌──────────────┐        ┌────────────────────────────────┐ │
│  │  tiklab CLI   │◄──────►│  Docker Engine                 │ │
│  │  (Go binary)  │  SDK   │                                │ │
│  └──────┬───────┘        │  ┌──────────────────────────┐  │ │
│         │                │  │  tiklab-sandbox container │  │ │
│         │ RouterOS API   │  │                          │  │ │
│         │ (port 8728)    │  │  ┌──────────┐            │  │ │
│         ├────────────────┼──┼─►│ QEMU     │            │  │ │
│         │                │  │  │┌────────┐│            │  │ │
│         │                │  │  ││RouterOS││ ◄─┐        │  │ │
│         │                │  │  ││  v7    ││   │        │  │ │
│         │                │  │  │└────────┘│   │        │  │ │
│         │                │  │  └──────────┘   │        │  │ │
│         │                │  │                 │ vNIC   │  │ │
│         │                │  │  ┌──────────────┴──────┐ │  │ │
│         │  Control API   │  │  │  Behavior Engine    │ │  │ │
│         ├────────────────┼──┼─►│  (Go binary)        │ │  │ │
│         │  (port 9090)   │  │  │  - DHCP clients     │ │  │ │
│         │                │  │  │  - Hotspot auth      │ │  │ │
│         │                │  │  │  - Traffic gen       │ │  │ │
│         │                │  │  └─────────────────────┘ │  │ │
│         │                │  └──────────────────────────┘  │ │
│         │                └────────────────────────────────┘ │
│  Exposed ports:                                             │
│    SSH (22), API (8728), Winbox (8291), Control (9090)      │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **RouterOS CHR via QEMU**: The Docker container runs QEMU with the official RouterOS v7 CHR disk image. This guarantees 100% API/protocol fidelity since it is the real RouterOS, not a simulation. The `evilfreelancer/docker-routeros:7` community image provides the base, customized with the behavior engine.

2. **In-Container Behavior Engine**: The traffic simulation engine runs inside the Docker container on a virtual network bridge connected to the QEMU RouterOS instance. This allows it to act as actual DHCP clients and Hotspot users on the same L2 network — generating real traffic that RouterOS processes natively.

3. **Control API**: The behavior engine exposes a minimal HTTP API on port 9090 for the host CLI to send commands (scale up/down, stop/start traffic). This avoids the complexity of `docker exec` and provides structured request/response.

4. **API-Based Reset (Fast Path)**: Instead of destroying and recreating the container, `tiklab reset` uses the RouterOS API to wipe all user-created configuration and re-apply the initial setup. The behavior engine regenerates users with fresh random identities. This keeps reset under 30 seconds.

5. **State File**: The CLI persists sandbox state to `~/.tiklab/state.json` (container ID, creation time, current status). This enables idempotent commands and meaningful error messages for conflicting operations (e.g., "sandbox already exists").

### Command Flow

| Command | Steps |
|---------|-------|
| `tiklab create` | (1) Check no sandbox exists → (2) Pull Docker image if needed → (3) Create container with port mappings → (4) Write state file (status: created) → (5) Print "Sandbox created." |
| `tiklab start` | (1) Check sandbox in created state → (2) Start container → (3) Wait for RouterOS boot (API health check) → (4) Apply initial config (DHCP, Hotspot, queues) → (5) Signal behavior engine to generate 50 users → (6) Wait for traffic stabilization → (7) Update state (status: running) → (8) Print connection info |
| `tiklab scale N` | (1) Check sandbox running → (2) Validate 1 ≤ N ≤ 500 → (3) Signal behavior engine with new count → (4) Wait for convergence → (5) Print "Scaled to N users." |
| `tiklab reset` | (1) Check sandbox running → (2) Signal behavior engine to stop traffic → (3) Wipe RouterOS config via API → (4) Re-apply initial config → (5) Signal engine to regenerate users (new random identities) → (6) Wait for stabilization → (7) Print "Reset complete." |
| `tiklab destroy` | (1) Check sandbox exists → (2) Stop container (force if running) → (3) Remove container, volumes, network → (4) Delete state file → (5) Print "Sandbox destroyed." |

## Complexity Tracking

No constitution violations to justify.
