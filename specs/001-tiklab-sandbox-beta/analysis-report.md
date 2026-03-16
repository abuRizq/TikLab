## Specification Analysis Report

**Scope**: `specs/001-tiklab-sandbox-beta/` | **Artifacts Analyzed**: spec.md, plan.md, tasks.md, constitution.md, data-model.md, contracts/cli-schema.md  
**Last Updated**: 2026-03-16 (post-refinement)

### Resolved Issues (Previous Report)

The following findings from the prior analysis have been addressed by strategic refinements:

| ID | Resolution |
|----|------------|
| **C1** | Constitution Principle V renamed to **CLI-First Distribution**. Docker image is no longer required to be standalone; CLI orchestration via `create` → `start` is the mandatory flow. |
| **H1** | `.rsc` scripts removed from plan.md project structure. API-based configuration (T019–T022) is the sole source of truth. |
| **H2** | Winbox functional verification deferred to post-Beta. T023 verifies Winbox port 8291 reachability only. Spec.md updated accordingly. |
| **M1** | Tasks renumbered to continuous T001–T050 sequence. |
| **M2** | T050 (CI pipeline) updated: integration tests run on all platforms (CLI↔Docker daemon, host-to-container ports). Container-internal traffic tests on linux/amd64 only. |
| **M4** | Engine entry point moved to Phase 5 (T034), after core engine logic (T024–T033). Eliminates compilation gap. |
| **M7** | T029 added: per-user Simple Queue creation via RouterOS API. T033 orchestrator includes `CreateUserQueue` in user onboarding flow. |

---

### Current Findings

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| M3 | Underspecification | MEDIUM | tasks T010/T035/T046, constitution V | No task publishes the Docker image to a registry. Constitution (CLI-First) states "Docker Hub hosts the sandbox image as a runtime dependency, pulled automatically by `tiklab create`." Build/push workflow is undefined. Acceptable for local development; blocks beta release. | Add a task for Docker image publishing (manual `docker push` in quickstart.md or automated via CI in T050). Specify registry name and tag convention. |
| M5 | Inconsistency | MEDIUM | cli-schema.md:L152-154, tasks T005 | cli-schema.md states the control API is "local to the Docker container network and not exposed to external networks." T005 maps container port 9090 to host port 9090 (required for CLI communication). Port is accessible on host localhost. | Update cli-schema.md: "Exposed on host localhost:9090 for CLI communication only; not intended for end-user interaction." |
| M6 | Coverage | MEDIUM | spec.md:L99 (US5 scenario 3), tasks T042 | US5 acceptance scenario 3: "existing connections are terminated cleanly" during reset. T042 implements reset but does not explicitly address graceful termination of active external SSH/API/Winbox connections. | Document that external connections may experience transient errors during reset (acceptable for beta), or add pre-reset notification. Add test case to T044 for active connections during reset. |
| L1 | Ambiguity | LOW | tasks T027 | T027 says "Use raw sockets or `github.com/insomniacslk/dhcp` library" — implementation choice left unresolved. | Decide during implementation. `github.com/insomniacslk/dhcp` recommended (higher-level, cross-platform). |

---

### Coverage Summary

| Requirement | Has Task? | Task IDs | Notes |
|-------------|-----------|----------|-------|
| FR-001 (single-command create) | Yes | T015 | |
| FR-002 (single-command start + traffic) | Yes | T016, T019–T022, T033–T035 | |
| FR-003 (expose SSH/API/Winbox ports) | Yes | T005, T015, T018 | |
| FR-004 (DHCP service) | Yes | T019 | |
| FR-005 (Hotspot service) | Yes | T020 | |
| FR-006 (~50 users on start) | Yes | T025, T033, T035 | |
| FR-007 (40/45/15 distribution) | Yes | T024, T025, T013 | |
| FR-008 (idle traffic) | Yes | T030 | |
| FR-009 (standard traffic) | Yes | T031 | |
| FR-010 (heavy traffic) | Yes | T032 | |
| FR-011 (bandwidth queues) | Yes | T021, T029 | Per-user queues via T029 |
| FR-012 (dynamic scaling 1–500) | Yes | T037–T039 | |
| FR-013 (reset to original state) | Yes | T041–T043 | |
| FR-014 (destroy all artifacts) | Yes | T017 | |
| FR-015 (cross-platform identical) | Yes | T045, T050 | CI runs integration tests on all platforms |
| FR-016 (interfaces reflect data) | Partial | T023, T036 | API + SSH tested; Winbox port reachability only (deferred) |
| FR-017 (clear error messages) | Yes | T007, T015, T037, T042, T047 | |
| FR-018 (prevent conflicts) | Yes | T007, T012 | |
| FR-019 (progress output) | Yes | T015, T016, T042 | |

---

### Constitution Alignment

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Single-Command Delivery | PASS | Each CLI operation is a single `tiklab <verb>`. `create` → `start` is mandatory two-step activation. |
| II. Simulation Fidelity | PASS | Real RouterOS CHR via QEMU. API/SSH/Winbox are native RouterOS protocols. |
| III. Isolation & Stateless Reset | PASS | Docker isolation. API-based reset. Clean destroy. |
| IV. Go-Native Standalone Binary | PASS | Go + cobra. Single binary. Cross-compiled. |
| V. CLI-First Distribution | PASS | CLI binary is primary interface. Docker image is runtime dependency pulled by `tiklab create`. |

---

### Unmapped Tasks

All tasks map to at least one requirement or user story. Infrastructure tasks:

- **Setup**: T001–T003
- **Foundational**: T004–T014
- **Polish**: T045–T050

No orphaned tasks.

---

### Metrics

| Metric | Value |
|--------|-------|
| Total Functional Requirements | 19 |
| Total Tasks | 50 |
| Full Coverage (requirement has implementing task) | 18/19 (95%) |
| Partial Coverage | 1/19 (FR-016: Winbox deferred) |
| Zero Coverage | 0/19 |
| User Stories | 5 |
| Acceptance Scenarios | 18 |
| Edge Cases Defined | 7 |
| Critical Issues | 0 |
| High Issues | 0 |
| Medium Issues | 3 |
| Low Issues | 1 |
| Total Open Findings | 4 |

---

### Next Actions

**No CRITICAL issues.** Proceed with `/speckit.implement` when ready.

**Recommended improvements (non-blocking):**

1. **M3** — Define Docker image publishing workflow before beta release.
2. **M5** — Clarify control API port documentation in cli-schema.md.
3. **M6** — Document reset behavior for active external connections; add test case to T044.
4. **L1** — Choose DHCP library during implementation.
