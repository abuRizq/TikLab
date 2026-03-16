# Feature Specification: TikLab Sandbox Beta

**Feature Branch**: `001-tiklab-sandbox-beta`
**Created**: 2026-03-16
**Status**: Draft
**Input**: MikroTik Sandbox CLI product overview — Beta scope covering Synthetic Mode with DHCP, Hotspot services, and realistic user traffic simulation.

## Clarifications

### Session 2026-03-16

- Q: Are `create` and `start` two separate commands or a single combined flow? → A: Two distinct commands. `create` provisions the environment, `start` activates it and begins traffic. Both are always required for a new sandbox.
- Q: Which RouterOS version does the sandbox simulate? → A: RouterOS v7 (current mainline, long-term support). v6 is out of scope for Beta.
- Q: What is the maximum supported simulated user count for Beta? → A: 500 users. The system MUST enforce this ceiling and report a clear error if exceeded.
- Q: How are simulated users identified (usernames, MACs)? → A: Randomized. Each create/reset generates fresh random usernames and MAC addresses for realism. Identities are not deterministic across sessions.
- Q: What does the CLI output during normal (non-error) operations? → A: Progress steps. Brief status lines for each phase (e.g., "Provisioning environment...", "Starting services...", "Ready."). No verbose log streaming by default.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create and Launch a Sandbox (Priority: P1)

A SaaS developer building a billing system wants to spin up a virtual MikroTik environment to test their integration. They run a single command and, within seconds, have a fully operational RouterOS instance with pre-configured network services, ready to accept connections via API, Winbox, or SSH.

**Why this priority**: Without a running sandbox, no other feature has value. This is the foundation of the entire product and the minimum viable deliverable.

**Independent Test**: Can be fully tested by running the create and start commands, then verifying that the sandbox is reachable via API, SSH, and Winbox ports on the host. Delivers immediate value: a working MikroTik test environment with zero hardware.

**Acceptance Scenarios**:

1. **Given** a host machine with the tool installed, **When** the developer runs the create command followed by the start command, **Then** a sandbox environment is running with API, Winbox, and SSH ports accessible on the host.
2. **Given** a running sandbox, **When** the developer runs the destroy command, **Then** all sandbox resources are removed and the ports are released.
3. **Given** a host machine without the tool's runtime prerequisites, **When** the developer attempts to create a sandbox, **Then** the system reports a clear, actionable error explaining what is missing.

---

### User Story 2 - Connect and Operate via Standard Protocols (Priority: P2)

A network automation engineer connects to the running sandbox using their existing tools (API client, Winbox application, SSH terminal) and executes standard RouterOS commands. The sandbox responds with the same data structures, error codes, and behaviors they would see on physical hardware. Their existing scripts and integrations work without modification.

**Why this priority**: Simulation fidelity is the core value proposition. If the sandbox behaves differently from real hardware, developers cannot trust their test results. This story validates that trust.

**Independent Test**: Can be tested by connecting to the sandbox via each protocol (API, SSH, Winbox), executing a set of standard RouterOS operations (query interfaces, list users, read DHCP leases), and comparing responses to documented RouterOS behavior.

**Acceptance Scenarios**:

1. **Given** a running sandbox, **When** a developer connects via API and queries the user list, **Then** the response format and data structure match RouterOS API specifications.
2. **Given** a running sandbox, **When** a developer connects via SSH and runs standard RouterOS CLI commands, **Then** the output matches the expected RouterOS terminal behavior.
3. *(Deferred to post-Beta)* **Given** a running sandbox, **When** a developer connects via Winbox, **Then** the graphical interface displays configuration and monitoring data consistent with a physical device. Beta scope verifies Winbox port reachability only; functional Winbox verification is on the post-Beta roadmap.
4. **Given** a running sandbox with Hotspot configured, **When** a developer queries active Hotspot sessions, **Then** the returned session data includes realistic user attributes (IP address, MAC address, uptime, bytes transferred).

---

### User Story 3 - Observe Realistic User Traffic (Priority: P3)

A developer building a Hotspot billing dashboard starts the sandbox and observes that simulated users are generating realistic network traffic. When they query the management interfaces, they see active DHCP leases, Hotspot sessions, bandwidth queue statistics, and consumption data that mirrors what a small ISP with ~50 subscribers would produce. The traffic distribution follows real-world patterns: most users are idle or browsing lightly, while a small percentage consumes heavy bandwidth.

**Why this priority**: Realistic data generation is what separates TikLab from a bare RouterOS instance. Without it, developers still face the "scaling gap" where code works with empty data but fails under realistic conditions.

**Independent Test**: Can be tested by starting a sandbox, waiting for traffic generation to stabilize, then querying Hotspot sessions, DHCP leases, and queue statistics. Verify that approximately 40% of users show minimal activity, 45% show moderate activity, and 15% show heavy consumption.

**Acceptance Scenarios**:

1. **Given** a started sandbox, **When** traffic generation stabilizes, **Then** approximately 50 simulated users appear as active sessions in the management interfaces.
2. **Given** active simulated traffic, **When** the developer queries DHCP lease information, **Then** each simulated user has a dynamically assigned IP address.
3. **Given** active simulated traffic, **When** the developer queries bandwidth queue statistics, **Then** per-user consumption data reflects the assigned behavior profile (idle users show minimal bytes, heavy users show high throughput).
4. **Given** active simulated traffic, **When** the developer queries Hotspot session data, **Then** user sessions include login timestamps, session durations, and cumulative data transfer values.
5. **Given** active simulated traffic, **When** the developer examines the distribution of user activity levels, **Then** the ratio approximates 40% idle / 45% standard / 15% heavy.

---

### User Story 4 - Scale Simulated Users Dynamically (Priority: P4)

A developer wants to stress-test their billing system under increasing load. While the sandbox is running, they issue a scale command to increase the number of active simulated users from 50 to 200. The additional users appear in the management interfaces and generate traffic according to the same behavior profile distribution. Later, they scale back down to verify their system handles user departures gracefully.

**Why this priority**: Dynamic scaling enables load testing without restarting the environment, which is critical for identifying performance bottlenecks in billing and automation systems. However, it builds on the foundation of Stories 1-3.

**Independent Test**: Can be tested by starting a sandbox with the default user count, issuing a scale-up command, verifying the new user count in management interfaces, then scaling back down and verifying the reduced count.

**Acceptance Scenarios**:

1. **Given** a running sandbox with 50 active users, **When** the developer issues a scale command to 200 users, **Then** the management interfaces reflect approximately 200 active users within 60 seconds.
2. **Given** a sandbox scaled to 200 users, **When** the developer issues a scale command back to 50 users, **Then** excess user sessions are terminated and the management interfaces reflect approximately 50 active users.
3. **Given** a scaled sandbox, **When** the developer queries user behavior distribution, **Then** the 40/45/15 ratio is maintained regardless of total user count.

---

### User Story 5 - Reset to Clean State (Priority: P5)

A developer has been running destructive tests — deleting users, changing firewall rules, modifying queue configurations. They want to start a fresh test cycle without tearing down and recreating the entire environment. They issue a reset command, and within seconds the sandbox reverts to its original state with default services, fresh user data, and no trace of the previous test session.

**Why this priority**: Fast iteration cycles are essential for developer productivity. Without reliable reset, developers must destroy and recreate the environment for each test cycle, which is slower and more disruptive.

**Independent Test**: Can be tested by starting a sandbox, making deliberate configuration changes (add/delete users, modify queues, change firewall rules), issuing the reset command, then verifying all changes are reverted and the sandbox matches its original post-creation state.

**Acceptance Scenarios**:

1. **Given** a sandbox with modified configurations (users deleted, queues changed, firewall rules added), **When** the developer issues the reset command, **Then** the sandbox returns to its original state with default services, users, and configurations.
2. **Given** a reset sandbox, **When** the developer queries the management interfaces, **Then** the data is identical to a freshly created and started sandbox.
3. **Given** a running sandbox with active billing system connections, **When** the developer issues the reset command, **Then** the system completes the reset and existing connections are terminated cleanly.

---

### Edge Cases

- What happens when the host machine's required ports (API, Winbox, SSH) are already occupied by another application?
- What happens when the developer attempts to scale beyond the 500-user Beta maximum?
- What happens when the developer issues `destroy` while the sandbox is still running with active external connections?
- What happens when `create` is called while a sandbox instance already exists?
- What happens when the developer issues `start` on a sandbox that is already running?
- What happens when `reset` is called on a stopped (but not destroyed) sandbox?
- What happens when network connectivity to the container runtime is lost mid-operation?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST create a complete sandbox environment from a single command invocation.
- **FR-002**: System MUST start the sandbox and begin generating simulated user traffic from a single command invocation.
- **FR-003**: System MUST expose API (host port 8728), Winbox (host port 8291), and SSH (host port 2222 → container port 22) access points on the host machine when the sandbox is running.
- **FR-004**: System MUST configure a DHCP service that dynamically allocates IP addresses to simulated users.
- **FR-005**: System MUST configure a Hotspot service with user authentication capabilities.
- **FR-006**: System MUST generate approximately 50 active simulated users upon environment start (Synthetic Mode default).
- **FR-007**: System MUST distribute simulated users across three behavior profiles: approximately 40% idle, 45% standard browsing, and 15% heavy usage.
- **FR-008**: Idle simulated users MUST generate minimal network activity (pings, DNS queries) sufficient to maintain session presence.
- **FR-009**: Standard browsing simulated users MUST generate traffic patterns resembling typical web surfing and small file downloads.
- **FR-010**: Heavy simulated users MUST generate continuous high-bandwidth traffic resembling large downloads or streaming.
- **FR-011**: System MUST implement bandwidth queues that enforce per-user limits and track consumption data.
- **FR-012**: System MUST support dynamic adjustment of the active simulated user count during a running session without restart. The supported range is 1 to 500 users. Requests exceeding 500 MUST be rejected with a clear error message.
- **FR-013**: System MUST restore the sandbox to its original post-creation state with a single reset command, removing all user-made changes.
- **FR-014**: System MUST completely remove all sandbox artifacts (containers, volumes, network resources) with a single destroy command.
- **FR-015**: System MUST produce identical behavior when deployed on Windows, Linux, and macOS.
- **FR-016**: Management interfaces (API, SSH, Winbox) MUST accurately reflect current simulated user data (sessions, leases, queue statistics, consumption).
- **FR-017**: System MUST report clear, actionable error messages when operations fail (e.g., port conflicts, missing prerequisites, invalid commands).
- **FR-018**: System MUST prevent conflicting lifecycle operations (e.g., creating a sandbox when one already exists, starting an already-running sandbox) and report the conflict to the user.
- **FR-019**: System MUST display brief progress status lines during multi-step operations (e.g., "Provisioning environment...", "Starting services...", "Ready."). Output MUST go to stdout on success and stderr on failure. No verbose log streaming by default.

### Key Entities

- **Sandbox Instance**: The complete virtual MikroTik environment running RouterOS v7. Has a lifecycle: not existing → (create) → created → (start) → running → (destroy) → not existing. The created state means the environment is provisioned but not generating traffic. The running state means services are active and simulated users are generating traffic. Only one instance exists at a time in the Beta scope.
- **Simulated User**: A virtual network subscriber with a randomly generated username and MAC address, an assigned behavior profile, a dynamically allocated IP address (via DHCP), and an optional Hotspot authentication session. Generates network traffic according to its profile. Identities are regenerated on each create or reset — they are not deterministic across sessions.
- **Behavior Profile**: Defines the network consumption pattern for a category of simulated users. Three profiles exist: Idle (minimal pings/DNS), Standard (web browsing, small downloads), Heavy (continuous high-bandwidth consumption). Determines traffic type, volume, and session activity level.
- **Network Service**: A configured service running inside the sandbox that simulated users interact with. Beta scope includes DHCP (address allocation) and Hotspot (access management with authentication).
- **Bandwidth Queue**: A traffic control rule linked to a simulated user or group, enforcing bandwidth limits and recording consumption metrics (bytes in, bytes out, current rate).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer with no prior experience provisions and launches a fully functional sandbox environment in under 2 minutes using two commands (`create` then `start`).
- **SC-002**: Management tools connected to the sandbox (API clients, SSH terminals) can execute standard RouterOS v7 operations without modification compared to physical hardware workflows. Winbox port (8291) is exposed and reachable; functional Winbox verification is deferred to post-Beta.
- **SC-003**: User consumption ratios (40% idle / 45% standard / 15% heavy) are accurately reflected in the sandbox's built-in accounting and monitoring interfaces, within a +/-5% tolerance.
- **SC-004**: The sandbox handles 50 concurrent simulated users while management access (API queries, SSH commands) remains responsive with no observable degradation.
- **SC-005**: The reset command returns the environment to a clean state in under 30 seconds, regardless of how many changes were made during the test session.
- **SC-006**: The identical deployment command produces functionally equivalent behavior on Windows, Linux, and macOS without platform-specific adjustments or instructions.
- **SC-007**: Scaling from 50 to 200 simulated users completes and stabilizes in under 60 seconds, with all new users visible in management interfaces.

### Assumptions

- Docker (or a compatible container runtime) is available on the host machine. The tool does not install Docker itself.
- The host machine has sufficient resources (CPU, memory) to run the sandbox container and simulate the default 50 users. Minimum resource requirements will be documented.
- The user has basic familiarity with command-line tools but does not need MikroTik-specific knowledge to deploy the sandbox.
- Beta scope is limited to a single sandbox instance at a time. Multi-instance support is deferred to a future release.
- The Hotspot login portal is functional within the sandbox but is not exposed to external browsers on the host (simulated users authenticate internally).
- A `tiklab status` command for inspecting sandbox state is deferred to post-Beta. Users verify sandbox state by connecting to the exposed ports.
- The Docker container image is not designed for standalone `docker run` usage. The `tiklab` CLI orchestrates container lifecycle, RouterOS configuration, and behavior engine coordination through mandatory sequential steps (`create` → `start`).
- Functional Winbox verification is deferred to post-Beta. The Winbox port (8291) is exposed and Winbox connections work (real RouterOS), but Beta testing validates API and SSH protocols only.
