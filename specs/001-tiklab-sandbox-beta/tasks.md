# Tasks: TikLab Sandbox Beta

**Input**: Design documents from `/specs/001-tiklab-sandbox-beta/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli-schema.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Go project initialization and dependency management

- [X] T001 Initialize Go module (`go mod init github.com/tiklab/tiklab`) and add dependencies: `github.com/spf13/cobra`, `github.com/docker/docker/client`, `github.com/go-routeros/routeros/v3` in go.mod
- [X] T002 Create project directory structure per plan.md: `cmd/tiklab/`, `cmd/tiklab-engine/`, `internal/cli/`, `internal/sandbox/`, `internal/docker/`, `internal/routeros/`, `internal/engine/`, `build/`, `build/scripts/`, `tests/integration/`, `tests/unit/`
- [X] T003 [P] Create Makefile with targets: `build` (compile `cmd/tiklab/main.go`), `build-engine` (compile behavior engine binary), `build-image` (docker build), `test` (go test ./...), `lint` (golangci-lint), and cross-compilation targets for `GOOS=linux,darwin,windows GOARCH=amd64,arm64`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Implement Docker client wrapper in internal/docker/client.go — connect to Docker daemon with `client.FromEnv` and `WithAPIVersionNegotiation()`, expose methods: `Connect() error`, `Close()`, `IsAvailable() bool`, `PullImage(tag string) error` with progress output
- [ ] T005 [P] Implement container lifecycle operations in internal/docker/container.go — methods: `CreateContainer(name, imageTag string, ports PortMapping) (containerID string, error)`, `StartContainer(id string) error`, `StopContainer(id string) error`, `RemoveContainer(id string, removeVolumes bool) error`, `ContainerExists(name string) (bool, error)`. Port mappings: SSH 2222→22, API 8728→8728, Winbox 8291→8291, Control 9090→9090. Add `NET_ADMIN` capability and `/dev/net/tun` device
- [ ] T006 [P] Implement SandboxState model and JSON persistence in internal/sandbox/state.go — struct fields: ContainerID, ContainerName, ImageTag, Status (created/running), CreatedAt, StartedAt, Ports, UserCount. Methods: `Load() (*SandboxState, error)`, `Save(state *SandboxState) error`, `Delete() error`. State file path: `~/.tiklab/state.json`. Create `~/.tiklab/` directory if not exists
- [ ] T007 Implement sandbox lifecycle manager in internal/sandbox/manager.go — state machine enforcing valid transitions per data-model.md: not existing→created (create), created→running (start), running→running (reset/scale), *→not existing (destroy). Methods: `Create()`, `Start()`, `Scale(n int)`, `Reset()`, `Destroy()`. Each method validates current state and returns descriptive errors for invalid transitions per contracts/cli-schema.md error messages
- [ ] T008 [P] Implement RouterOS API client wrapper in internal/routeros/client.go — methods: `Connect(host string, port int, user, pass string) error`, `Close()`, `Run(command string, args ...string) (*routeros.Reply, error)`, `WaitForReady(timeout time.Duration) error` (poll `/system/identity/print` every 3 seconds, up to 90 seconds total, log each retry, return error "RouterOS failed to boot within timeout" if exhausted). Use `github.com/go-routeros/routeros/v3`
- [ ] T009 Create CLI entry point and root command in cmd/tiklab/main.go and internal/cli/root.go — root command `tiklab` with `--version` and `--help` flags using cobra. Version string injected via `-ldflags` at build time. Register subcommands: create, start, scale, reset, destroy (stubs initially)
- [ ] T010 Create base Dockerfile in build/Dockerfile — runtime stage based on `evilfreelancer/docker-routeros:7`, copy `build/scripts/entrypoint.sh`. Expose ports 22, 8728, 8291, 9090. Engine build stage is added in Phase 5 after engine implementation is complete
- [ ] T011 [P] Create container entrypoint script in build/scripts/entrypoint.sh — start QEMU/RouterOS (delegate to base image entrypoint), wait for RouterOS API to become reachable on internal network, then start the behavior engine binary in the background. Accept environment variables for control API port (default 9090)
- [ ] T012 [P] Implement unit tests for sandbox state machine in tests/unit/sandbox_test.go — test valid transitions (created→running), invalid transitions (running→created), state persistence (save/load round-trip), error messages for conflicting operations per data-model.md
- [ ] T013 [P] Implement unit tests for behavior engine in tests/unit/engine_test.go — test profile assignment distribution (40/45/15 ±5%), user identity generation (uniqueness of username and MAC), GenerateUsers count accuracy
- [ ] T014 [P] Implement unit tests for RouterOS config generation in tests/unit/config_test.go — test DHCP, Hotspot, and queue configuration command generation, verify correct API command sequences

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 — Create and Launch a Sandbox (Priority: P1) MVP

**Goal**: Developer can create, start, and destroy a sandbox environment. Ports (SSH, API, Winbox) are accessible on the host after start.

**Independent Test**: Run `tiklab create && tiklab start`, verify ports 2222, 8728, 8291 respond on localhost, then run `tiklab destroy` and verify ports are released and state file is deleted.

### Implementation for User Story 1

- [ ] T015 [US1] Implement `tiklab create` command in internal/cli/create.go — pre-flight: verify Docker daemon is reachable (error "Docker is not running. Please start Docker and try again." to stderr, exit 1), check for port conflicts on 2222, 8728, 8291, 9090 (error "Port N is already in use" to stderr, exit 1). Then: check no sandbox exists (load state, error if found), pull image if needed (print "Pulling image..."), create container with port mappings, write state file (status: created), print port info and "Run `tiklab start` to activate." per CLI contract
- [ ] T016 [US1] Implement `tiklab start` command in internal/cli/start.go — load state (error if not found or already running), start container, call `routeros.WaitForReady()` with timeout (print "Waiting for RouterOS to boot..."), update state (status: running, set StartedAt), print connection info per CLI contract. Integrate with RouterOS config application (placeholder for Phase 4 wiring)
- [ ] T017 [P] [US1] Implement `tiklab destroy` command in internal/cli/destroy.go — load state (error if not found with "Nothing to destroy"), stop container if running (force), remove container with volumes, delete state file, print "Sandbox destroyed."
- [ ] T018 [US1] Implement integration test for full lifecycle in tests/integration/lifecycle_test.go — test create→start→destroy cycle: verify container creation, port reachability (SSH 2222, API 8728, Winbox 8291), state file updates, clean destroy. Assert create+start completes in under 2 minutes (SC-001)

**Checkpoint**: At this point, `tiklab create && tiklab start` boots a RouterOS v7 sandbox with accessible ports. `tiklab destroy` cleans up completely. MVP is functional.

---

## Phase 4: User Story 2 — Connect and Operate via Standard Protocols (Priority: P2)

**Goal**: RouterOS is configured with DHCP and Hotspot services. API and SSH connections work identically to physical hardware. Winbox port is reachable (functional verification deferred to post-Beta).

**Independent Test**: After `tiklab start`, connect via SSH (`ssh admin@localhost -p 2222`) and run `/ip/dhcp-server/print`, `/ip/hotspot/print`. Connect via API (port 8728) and query `/system/identity/print`. Verify responses match RouterOS v7 format.

### Implementation for User Story 2

- [ ] T019 [US2] Implement DHCP server configuration via RouterOS API in internal/routeros/config.go — function `ConfigureDHCP(client *Client) error`: create IP pool `10.10.0.10-10.10.1.254` (501 addresses, supports 500-user max), assign address `10.10.0.1/22` to bridge interface, create DHCP server on bridge with 1h lease time, set DHCP network gateway to `10.10.0.1` with DNS `8.8.8.8`. Use `/ip/pool/add`, `/ip/address/add`, `/ip/dhcp-server/add`, `/ip/dhcp-server/network/add`
- [ ] T020 [US2] Implement Hotspot server configuration via RouterOS API in internal/routeros/config.go — function `ConfigureHotspot(client *Client) error`: create Hotspot server on bridge interface, set address pool to DHCP pool, configure HTTP PAP login method, create default user profile with no bandwidth limit at Hotspot level. Use `/ip/hotspot/add`, `/ip/hotspot/profile/set`, `/ip/hotspot/user/profile/add`
- [ ] T021 [US2] Implement simple queue template configuration via RouterOS API in internal/routeros/config.go — function `ConfigureQueueTemplate(client *Client) error`: define queue parameters per behavior profile (idle: 256k/256k, standard: 2M/2M, heavy: 5M/5M max-limit). This creates the template; per-user queues are created by the behavior engine in Phase 5. Use `/queue/simple/add`
- [ ] T022 [US2] Implement combined initial configuration function in internal/routeros/config.go — function `ApplyInitialConfig(client *Client) error` that calls `ConfigureDHCP`, `ConfigureHotspot`, `ConfigureQueueTemplate` in sequence with progress output. Wire into `tiklab start` command flow in internal/cli/start.go after RouterOS boot completes
- [ ] T023 [US2] Implement integration test for protocol access in tests/integration/protocol_test.go — connect via RouterOS API (port 8728) and run `/system/identity/print`, connect via SSH (port 2222) and run `/ip/dhcp-server/print`, verify DHCP server and Hotspot are configured and return RouterOS v7 response format. Verify Winbox port 8291 is reachable via TCP connect (functional Winbox verification deferred to post-Beta) (SC-002)

**Checkpoint**: Sandbox now boots with DHCP + Hotspot fully configured. API and SSH connections return correct RouterOS v7 responses. Winbox port is reachable.

---

## Phase 5: User Story 3 — Observe Realistic User Traffic (Priority: P3)

**Goal**: Behavior engine generates 50 simulated users with realistic traffic patterns. Management interfaces show active DHCP leases, Hotspot sessions, and queue statistics matching the 40/45/15 profile distribution.

**Independent Test**: After `tiklab start`, query `/ip/dhcp-server/lease/print` via API and verify ~50 active leases. Query `/ip/hotspot/active/print` and verify active sessions. Query `/queue/simple/print` and verify per-user queues with non-zero byte counters. Check profile distribution is approximately 40% idle / 45% standard / 15% heavy.

### Implementation for User Story 3

- [ ] T024 [P] [US3] Define behavior profile constants and parameters in internal/engine/profiles.go — struct `ProfileConfig` with fields: Name, TrafficType, IntervalMin/Max, ThroughputTarget. Define three profiles per research.md R8: idle (ping 30s, DNS 60s, ~1KB/min), standard (HTTP 5-15s, ~50-200KB/min), heavy (continuous TCP, ~500KB-1MB/min). Export `AssignProfiles(count int) []Profile` distributing 40/45/15
- [ ] T025 [P] [US3] Implement simulated user identity generation in internal/engine/user.go — struct `SimulatedUser` per data-model.md. Function `GenerateUsers(count int, profiles []Profile) []SimulatedUser`: generate random username (format `guest_` + 6 hex chars), random locally-administered MAC address, assign profile from pre-computed distribution. Ensure uniqueness of username and MAC within the batch
- [ ] T026 [US3] Implement behavior engine HTTP control API in internal/engine/engine.go — `Engine` struct managing user lifecycle. HTTP server on port 9090 (configurable via env). Endpoints per contracts: `POST /start` (begin traffic with count), `POST /stop` (halt all traffic), `POST /scale` (adjust count), `GET /status` (return user counts by profile). JSON request/response bodies per cli-schema.md
- [ ] T027 [US3] Implement DHCP client simulation in internal/engine/traffic.go — function `SimulateDHCPClient(user *SimulatedUser) error`: create virtual network interface with user's MAC, send DHCP discover/request on the bridge network, store assigned IP in user struct. Use raw sockets or `github.com/insomniacslk/dhcp` library for DHCP client protocol
- [ ] T028 [US3] Implement Hotspot authentication simulation in internal/engine/traffic.go — function `AuthenticateHotspot(user *SimulatedUser) error`: perform HTTP POST to RouterOS Hotspot login page (typically `http://10.10.0.1/login`) with username from user struct. Mark `SessionActive = true` on success
- [ ] T029 [US3] Implement per-user Simple Queue creation in internal/engine/traffic.go — function `CreateUserQueue(user *SimulatedUser) error`: after DHCP IP assignment and Hotspot authentication, connect to RouterOS API (internal address `10.10.0.1:8728`) and create a Simple Queue via `/queue/simple/add` with target set to user's assigned IP, max-limit based on behavior profile (idle: 256k/256k, standard: 2M/2M, heavy: 5M/5M). On user teardown, remove the queue via `/queue/simple/remove`. Ensures per-user bandwidth limits and consumption metrics are visible in management interfaces
- [ ] T030 [P] [US3] Implement idle traffic generator in internal/engine/traffic.go — function `RunIdleTraffic(user *SimulatedUser, stop chan struct{})`: send ICMP ping to gateway (10.10.0.1) every 30 seconds, DNS query every 60 seconds. Run as goroutine per user. Respect stop channel for clean shutdown
- [ ] T031 [P] [US3] Implement standard traffic generator in internal/engine/traffic.go — function `RunStandardTraffic(user *SimulatedUser, stop chan struct{})`: HTTP GET requests to built-in HTTP server every 5-15 seconds (randomized interval), occasional 10-100KB file download. Run as goroutine per user
- [ ] T032 [P] [US3] Implement heavy traffic generator in internal/engine/traffic.go — function `RunHeavyTraffic(user *SimulatedUser, stop chan struct{})`: continuous TCP stream to/from built-in traffic sink. Maintain connection and stream data constantly. Run as goroutine per user
- [ ] T033 [US3] Implement engine orchestrator startup flow in internal/engine/engine.go — `Engine.Start(count int)`: call `GenerateUsers`, for each user call `SimulateDHCPClient` then `AuthenticateHotspot` then `CreateUserQueue`, then launch traffic goroutine matching user's profile. Track all goroutines for clean shutdown. Handle partial failures (log and continue if individual user setup fails)
- [ ] T034 [US3] Create behavior engine binary entry point in cmd/tiklab-engine/main.go — `main()` function that initializes the `Engine` from `internal/engine/engine.go`, reads configuration from environment variables (control API port, default user count), and starts the HTTP control server. This is the binary compiled into the Docker image
- [ ] T035 [US3] Update Dockerfile to compile and embed behavior engine binary in build/Dockerfile — add Go builder stage compiling `cmd/tiklab-engine/` to `/usr/local/bin/tiklab-engine`. Update entrypoint.sh to start engine binary after RouterOS is ready. Wire `tiklab start` command to call `POST /start {"count": 50}` on the control API after config is applied
- [ ] T036 [US3] Implement integration test for traffic verification in tests/integration/traffic_test.go — query `/ip/dhcp-server/lease/print` (verify ~50 leases), `/ip/hotspot/active/print` (verify sessions), `/queue/simple/print` (verify per-user queues with non-zero counters). Assert profile distribution within ±5% of 40/45/15 (SC-003). Verify API responsiveness under 50 users (SC-004)

**Checkpoint**: Sandbox now boots with 50 simulated users generating traffic. DHCP leases, Hotspot sessions, and per-user queue statistics are visible in management interfaces with correct profile distribution.

---

## Phase 6: User Story 4 — Scale Simulated Users Dynamically (Priority: P4)

**Goal**: Developer can adjust simulated user count during a running session using `tiklab scale <count>` without restarting the environment.

**Independent Test**: With sandbox running at 50 users, run `tiklab scale 200`, verify ~200 users in `/ip/hotspot/active/print`. Run `tiklab scale 50`, verify ~50 users. Check 40/45/15 distribution maintained at both counts.

### Implementation for User Story 4

- [ ] T037 [US4] Implement `tiklab scale` command in internal/cli/scale.go — parse count argument (validate integer, range 1-500), load state (error if not running), call behavior engine `POST /scale {"count": N}` on control API port, wait for response, update state file UserCount, print "Scaled to N users." per CLI contract. Error messages per cli-schema.md for invalid range and not-running state
- [ ] T038 [US4] Implement scale-up logic in internal/engine/engine.go — `Engine.ScaleTo(target int)`: if target > current count, generate additional users with `GenerateUsers(delta)`, run DHCP+Hotspot+CreateUserQueue+traffic for each new user. Maintain 40/45/15 profile distribution across the total user pool (not just the delta)
- [ ] T039 [US4] Implement scale-down logic in internal/engine/engine.go — if target < current count, select excess users (LIFO or random), stop their traffic goroutines via stop channel, remove user queue via RouterOS API, release DHCP lease, deauthenticate Hotspot session, remove virtual network interface. Update internal user list
- [ ] T040 [US4] Implement integration test for dynamic scaling in tests/integration/traffic_test.go — scale 50→200 users, verify ~200 active sessions, verify 40/45/15 distribution maintained, assert scaling completes in under 60 seconds (SC-007). Scale back to 50, verify ~50 sessions

**Checkpoint**: `tiklab scale` adjusts user count in both directions. Profile distribution is maintained.

---

## Phase 7: User Story 5 — Reset to Clean State (Priority: P5)

**Goal**: Developer can reset the sandbox to its original state with `tiklab reset`, wiping all configuration changes and regenerating users with fresh random identities. No container restart.

**Independent Test**: Start sandbox, make changes via SSH (add firewall rule, delete a user, modify a queue), run `tiklab reset`, verify via API that config matches a fresh sandbox (no extra firewall rules, default users regenerated, queues reset). Verify reset completes in under 30 seconds.

### Implementation for User Story 5

- [ ] T041 [US5] Implement RouterOS configuration wipe via API in internal/routeros/config.go — function `WipeConfig(client *Client) error`: remove all simple queues (`/queue/simple/remove`), remove all Hotspot users and active sessions (`/ip/hotspot/user/remove`, `/ip/hotspot/active/remove`), remove Hotspot server, remove DHCP leases and server, remove IP addresses, remove extra firewall rules. Order matters: remove dependents before parents
- [ ] T042 [US5] Implement `tiklab reset` command in internal/cli/reset.go — load state (error if not running), call behavior engine `POST /stop` to halt traffic, call `WipeConfig` via RouterOS API (print "Clearing configuration..."), call `ApplyInitialConfig` to re-apply DHCP/Hotspot/queues (print "Reapplying initial setup..."), call engine `POST /start {"count": 50}` to regenerate users with fresh identities (print "Regenerating users..."), print "Reset complete." State remains running, UserCount reverts to 50
- [ ] T043 [US5] Implement engine full restart in internal/engine/engine.go — `Engine.Stop()` must cleanly shut down all user goroutines (signal stop channels, wait for completion), remove all user queues via RouterOS API, release all network interfaces, clear internal user list. Subsequent `Engine.Start(count)` generates entirely new users with fresh random identities
- [ ] T044 [US5] Implement integration test for reset in tests/integration/lifecycle_test.go — make config changes via RouterOS API (add firewall rule, delete user), run `tiklab reset`, verify clean state matches fresh sandbox via API queries, assert reset completes in under 30 seconds (SC-005)

**Checkpoint**: `tiklab reset` restores the sandbox to a clean state in under 30 seconds without restarting the container.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T045 [P] Add cross-platform CI build script in Makefile — targets for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`. Produce binaries named `tiklab-{os}-{arch}` in `dist/`
- [ ] T046 [P] Add Docker image versioning in build/Dockerfile — accept `VERSION` build arg, tag image as `tiklab/sandbox:${VERSION}` and `tiklab/sandbox:latest`. Update CLI create command to use version-matched image tag
- [ ] T047 [P] Add graceful error handling for Docker connectivity loss in internal/docker/client.go — detect Docker socket errors mid-operation, wrap with user-friendly messages ("Docker connection lost. Is Docker running?"), ensure state file consistency on partial failures (write state before Docker ops, clean up on failure), apply to all commands. Cover edge case: network/Docker connectivity lost during create/start/reset/scale operations
- [ ] T048 Run full lifecycle validation: `tiklab create` → `tiklab start` → verify 50 users via API → `tiklab scale 200` → verify 200 users → `tiklab reset` → verify clean state with 50 fresh users → `tiklab destroy` → verify no artifacts remain
- [ ] T049 [P] Validate quickstart.md instructions end-to-end — execute each step in specs/001-tiklab-sandbox-beta/quickstart.md on a clean machine, confirm all commands and expected outputs match
- [ ] T050 [P] Create CI pipeline in .github/workflows/ci.yml — matrix builds for linux/amd64, darwin/amd64, darwin/arm64, windows/amd64. Steps: checkout, setup Go, lint (`golangci-lint`), run unit tests, cross-compile CLI and engine binaries. Integration tests on all platforms with Docker: verify CLI↔Docker daemon interaction (container create/start/destroy), host-to-container port reachability (SSH, API). Container-internal traffic tests on linux/amd64 only. Satisfies constitution cross-platform CI mandate

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion
- **User Story 2 (Phase 4)**: Depends on US1 (needs working create/start)
- **User Story 3 (Phase 5)**: Depends on US2 (needs DHCP/Hotspot configured)
- **User Story 4 (Phase 6)**: Depends on US3 (needs behavior engine running)
- **User Story 5 (Phase 7)**: Depends on US2 + US3 (needs config + engine to reset)
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Foundational (Phase 2) — No dependencies on other stories
- **US2 (P2)**: Depends on US1 — needs create/start commands to boot RouterOS
- **US3 (P3)**: Depends on US2 — needs DHCP/Hotspot configured for traffic to flow
- **US4 (P4)**: Depends on US3 — needs behavior engine running to scale
- **US5 (P5)**: Depends on US2 + US3 — needs config to wipe and engine to restart

### Within Each User Story

- Models/data structures before services
- Services before CLI commands
- Core implementation before error handling polish
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: T003 can run in parallel with T001/T002
- **Phase 2**: T004+T005 (Docker), T006 (state), T008 (RouterOS client), T011 (entrypoint) — all target different files and can run in parallel after T002. T012+T013+T014 (unit tests) can run in parallel after their target implementations complete
- **Phase 3**: T017 (destroy) can run in parallel with T015/T016 (create/start)
- **Phase 5**: T024 (profiles), T025 (user gen), T030+T031+T032 (traffic generators) — all target different files
- **Phase 8**: T045, T046, T047, T049, T050 — all independent

---

## Parallel Example: Phase 2 (Foundational)

```bash
# These target different files and can run simultaneously:
Task: "Implement Docker client wrapper in internal/docker/client.go" (T004)
Task: "Implement container lifecycle in internal/docker/container.go" (T005)
Task: "Implement SandboxState model in internal/sandbox/state.go" (T006)
Task: "Implement RouterOS API client in internal/routeros/client.go" (T008)
Task: "Create entrypoint script in build/scripts/entrypoint.sh" (T011)

# Then sequentially:
Task: "Implement sandbox lifecycle manager in internal/sandbox/manager.go" (T007) — depends on T006
Task: "Create CLI root command in cmd/tiklab/main.go" (T009) — depends on T007
```

## Parallel Example: Phase 5 (User Story 3 — Traffic)

```bash
# Profile definitions and user generation can run simultaneously:
Task: "Define behavior profiles in internal/engine/profiles.go" (T024)
Task: "Implement user identity generation in internal/engine/user.go" (T025)

# All three traffic generators target different functions in the same file but are independent:
Task: "Implement idle traffic generator in internal/engine/traffic.go" (T030)
Task: "Implement standard traffic generator in internal/engine/traffic.go" (T031)
Task: "Implement heavy traffic generator in internal/engine/traffic.go" (T032)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: `tiklab create && tiklab start` → verify SSH/API/Winbox ports respond → `tiklab destroy`
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → **MVP!** (bootable sandbox)
3. Add User Story 2 → Test independently → DHCP + Hotspot configured
4. Add User Story 3 → Test independently → realistic traffic flowing
5. Add User Story 4 → Test independently → dynamic scaling works
6. Add User Story 5 → Test independently → instant reset works
7. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- User stories are sequential (US2 depends on US1, US3 on US2, etc.) due to architectural layering
