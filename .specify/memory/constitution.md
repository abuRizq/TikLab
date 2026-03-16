<!--
=== Sync Impact Report ===
Version change: 1.0.0 → 2.0.0

Modified principles:
  - I. Single-Command Delivery: Replaced `docker run` delivery contract with CLI binary contract.
    Explicitly documents `create` → `start` as mandatory two-step activation.
  - V. Docker-First Distribution → CLI-First Distribution: Docker image is now a CLI-managed
    runtime dependency, not a standalone entry point.

Added sections: N/A
Removed sections: N/A

Templates requiring updates:
  - .specify/templates/plan-template.md        ✅ compatible (Constitution Check section references principle names generically)
  - .specify/templates/spec-template.md         ✅ compatible
  - .specify/templates/tasks-template.md        ✅ compatible

Follow-up TODOs:
  - Update plan.md Constitution Check table for Principle V name and evidence
=== End Report ===
-->

# TikLab Constitution

## Core Principles

### I. Single-Command Delivery (NON-NEGOTIABLE)

The sandbox environment MUST be deployable and operable through minimal CLI commands. There MUST NOT be multi-step installation guides, prerequisite dependency chains, or manual configuration steps.

- The `tiklab` CLI binary is the delivery contract. Each lifecycle operation is a single `tiklab <verb>` invocation.
- `tiklab create` followed by `tiklab start` is the mandatory two-step activation sequence. Both are always required for a new sandbox.
- If a user needs to read documentation beyond command help to get a working environment, the onboarding has failed.

**Rationale**: The product's primary differentiator is zero-friction deployment. Any friction in setup directly undermines the value proposition and erodes trust with the target audience (SaaS developers who need fast iteration cycles).

### II. Simulation Fidelity

The sandbox MUST respond exactly like an independent physical MikroTik device. External integrations connecting to the sandbox MUST NOT be able to distinguish it from real hardware through normal API, Winbox, or SSH interactions.

- API access MUST match RouterOS API behavior and response formats.
- Winbox graphical access MUST be functional for manual configuration and monitoring.
- SSH terminal access MUST accept standard RouterOS commands.
- User behavior profiles MUST accurately reflect real-world consumption patterns: 40% idle, 45% standard browsing, 15% heavy usage.
- DHCP and Hotspot services MUST behave identically to their physical counterparts.

**Rationale**: The product exists so developers can trust that code tested against the sandbox will work against real hardware. Any behavioral divergence defeats the purpose and creates false confidence.

### III. Isolation & Stateless Reset

Every sandbox instance MUST be fully isolated from the host system and from other sandbox instances. The reset mechanism MUST restore a pristine state instantly without reinstallation or manual cleanup.

- Destructive operations inside the sandbox (user disconnection, configuration overrides, firewall changes) MUST never affect the host system or other sandboxes.
- `tiklab reset` MUST return the environment to its original clean state deterministically.
- `tiklab destroy` MUST remove all artifacts completely, leaving no orphaned resources.
- State management MUST support returning to a "clean slate" at any time.

**Rationale**: Safety is a core selling point. Developers MUST be able to run destructive or experimental commands with zero risk. If reset is slow or unreliable, developers lose trust and revert to physical hardware.

### IV. Go-Native Standalone Binary

The CLI and control unit MUST be written in Go. The delivered artifact MUST be a self-contained binary with zero runtime dependencies beyond Docker (for the simulated environment).

- No interpreted language runtimes, no JVMs, no external library dependencies at runtime.
- The binary MUST compile and run natively on Windows, Linux, and macOS without modification.
- Cross-compilation MUST be part of the release process.
- Resource consumption (CPU, memory) of the CLI itself MUST remain minimal; heavy work belongs inside the container.

**Rationale**: Go guarantees a single static binary with fast startup, low memory footprint, and native cross-platform support. This directly enables the single-command delivery principle and eliminates "works on my machine" failures.

### V. CLI-First Distribution

The `tiklab` CLI binary is the primary user interface, distributed via GitHub Releases for all supported platforms. The Docker image encapsulates the complete sandbox environment (RouterOS simulator, behavior engine, network configuration) but is NOT designed for standalone `docker run` usage — the CLI orchestrates the container through mandatory sequential steps (`create` → `start`).

- GitHub Releases is the primary distribution channel for the CLI binary.
- Docker Hub hosts the sandbox image as a runtime dependency, pulled automatically by `tiklab create`.
- The Docker image MUST be versioned and tagged following semantic versioning.
- Image size MUST be kept as small as practical; multi-stage builds are expected.

**Rationale**: The sandbox requires CLI orchestration (state tracking, RouterOS API configuration, behavior engine coordination) that cannot be replicated by a bare `docker run`. The CLI binary provides the zero-friction interface while Docker provides the cross-platform runtime guarantee.

## Technology Constraints

- **Core Language**: Go (Golang) — CLI, control unit, behavior engine orchestration.
- **Containerization**: Docker — primary runtime and distribution mechanism.
- **Simulated OS**: MikroTik RouterOS (CHR or equivalent virtualized image).
- **CLI Prefix**: `tiklab` — all user-facing commands use this prefix.
- **Beta Scope**: Synthetic mode only. Services limited to DHCP and Hotspot. Approximately 50 simulated active users.
- **Access Ports**: API, Winbox, and SSH MUST be exposed and mapped to the host.
- **Out of Scope (Beta)**: PPPoE, Mirror Mode, multiple traffic profiles, LAN integration, RADIUS, attack simulation, monitoring dashboards.

## Development Workflow

- **Beta-First**: Ship the smallest viable mode (Synthetic with DHCP + Hotspot) before adding complexity. Features from the roadmap (Section 6 of the product spec) MUST NOT leak into Beta scope.
- **CLI Contract Stability**: The five core commands (`create`, `start`, `reset`, `destroy`, `scale`) define the Beta CLI surface. New commands require explicit scope approval.
- **Integration Testing**: All CLI commands MUST be tested end-to-end against a running Docker environment. API responsiveness MUST be verified under simulated load (50 concurrent users at bandwidth saturation).
- **Cross-Platform CI**: Builds and tests MUST run on Windows, Linux, and macOS in CI. A platform-specific failure is a release blocker.
- **Commit Discipline**: Each commit MUST represent a single logical change. Commits MUST NOT mix CLI changes with behavior engine changes with Docker configuration changes.

## Governance

This constitution is the authoritative reference for all architectural and development decisions on TikLab. When a proposed change conflicts with a principle defined here, the constitution takes precedence unless formally amended.

- **Amendment Process**: Any principle change requires a documented proposal explaining (a) which principle is affected, (b) why the current formulation is insufficient, and (c) what the replacement text is. Amendments MUST be versioned.
- **Versioning Policy**: Constitution versions follow semantic versioning. MAJOR for principle removals or redefinitions, MINOR for new principles or material expansions, PATCH for clarifications and wording fixes.
- **Compliance Review**: All code reviews and design discussions MUST verify alignment with these principles. Complexity or scope additions MUST be justified against the constitution.

**Version**: 2.0.0 | **Ratified**: 2026-03-16 | **Last Amended**: 2026-03-16
