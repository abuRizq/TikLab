# Data Model: TikLab Sandbox Beta

**Branch**: `001-tiklab-sandbox-beta` | **Date**: 2026-03-16

## Entities

### SandboxState

Persisted to `~/.tiklab/state.json`. Tracks the single active sandbox instance.

| Field | Type | Description |
|-------|------|-------------|
| ContainerID | string | Docker container ID |
| ContainerName | string | Docker container name (e.g., `tiklab-sandbox`) |
| ImageTag | string | Docker image tag used (e.g., `tiklab/sandbox:0.1.0`) |
| Status | enum | `created`, `running` |
| CreatedAt | timestamp | ISO 8601 creation time |
| StartedAt | timestamp | ISO 8601 start time (null if not started) |
| Ports | PortMapping | Mapped host ports for SSH, API, Winbox, Control |
| UserCount | int | Current target number of simulated users |

**Validation rules**:
- Only one SandboxState may exist at a time. Creating a new sandbox while one exists is an error.
- Status transitions: `created → running`. No backward transitions. Reset does not change status (remains `running`). Destroy deletes the state file entirely rather than setting a terminal status.
- ContainerID must be non-empty and correspond to an existing Docker container.

### PortMapping

Embedded in SandboxState. Defines host-to-container port mappings.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| SSH | int | 2222 | Host port mapped to container SSH (22) |
| API | int | 8728 | Host port mapped to container RouterOS API (8728) |
| Winbox | int | 8291 | Host port mapped to container Winbox (8291) |
| Control | int | 9090 | Host port mapped to behavior engine control API (9090) |

### SimulatedUser

Managed by the behavior engine inside the container. Not persisted to disk — regenerated on each start/reset.

| Field | Type | Description |
|-------|------|-------------|
| ID | int | Sequential index (1–N) for internal tracking |
| Username | string | Randomly generated Hotspot username (e.g., `guest_a7f3b2`) |
| MACAddress | string | Randomly generated MAC address (locally administered bit set) |
| IPAddress | string | Dynamically assigned by RouterOS DHCP server |
| Profile | enum | `idle`, `standard`, `heavy` |
| SessionActive | bool | Whether the user has an active Hotspot session |
| BytesIn | int64 | Cumulative bytes received (as reported by RouterOS) |
| BytesOut | int64 | Cumulative bytes sent (as reported by RouterOS) |

**Validation rules**:
- Username must be unique within the current session.
- MACAddress must be unique within the current session and use the locally administered bit (second hex digit is 2, 6, A, or E).
- Profile assignment follows distribution: 40% idle, 45% standard, 15% heavy (±5% tolerance per spec SC-003).
- IPAddress is assigned by RouterOS DHCP server — not controlled by the behavior engine.

### BehaviorProfile

Static configuration compiled into the behavior engine binary.

| Profile | Traffic Type | Activity Interval | Throughput Target | Description |
|---------|-------------|-------------------|-------------------|-------------|
| idle | ICMP ping, DNS query | Ping every 30s, DNS every 60s | ~1 KB/min | Keeps session alive with minimal footprint |
| standard | HTTP GET, small downloads | Request every 5–15s | ~50–200 KB/min | Simulates web browsing and light usage |
| heavy | Continuous TCP stream | Constant | ~500 KB–1 MB/min | Simulates streaming or large downloads |

### RouterOSConfig

Initial configuration applied by the CLI during `tiklab start`. Represented as RouterOS API commands, not persisted separately.

| Service | Configuration |
|---------|--------------|
| DHCP Server | Pool: `10.10.0.10–10.10.1.254` (501 addresses, supports 500-user max), Gateway: `10.10.0.1/22`, Lease time: 1h, Interface: internal bridge |
| Hotspot | Server on internal bridge, Login method: HTTP PAP, User profile: `default` (no bandwidth limit at Hotspot level), Address pool: shared with DHCP |
| Simple Queues | One queue per simulated user, target: user's IP, max-limit based on behavior profile |
| Bridge | Internal bridge connecting QEMU virtual NIC to behavior engine virtual NICs |

## State Transitions

### Sandbox Lifecycle

```text
[not existing] ──create──► [created] ──start──► [running] ──destroy──► [not existing]
                                                    │  ▲
                                                    │  │
                                                    └──┘
                                                   reset
                                                  (stays running)
```

| Transition | Trigger | Side Effects |
|-----------|---------|--------------|
| → created | `tiklab create` | Docker container created, ports allocated, state file written |
| created → running | `tiklab start` | Container started, RouterOS booted, config applied, traffic started |
| running → running | `tiklab reset` | Config wiped via API, re-applied, users regenerated (new random identities) |
| running → running | `tiklab scale N` | Behavior engine adjusts user count (add/remove simulated users) |
| * → not existing | `tiklab destroy` | Container stopped and removed, volumes/networks cleaned, state file deleted |

### Invalid Transitions (error responses)

| Attempted | Current State | Error |
|-----------|--------------|-------|
| create | created or running | "Sandbox already exists. Run `tiklab destroy` first." |
| start | running | "Sandbox is already running." |
| start | not existing | "No sandbox found. Run `tiklab create` first." |
| reset | created | "Sandbox is not running. Run `tiklab start` first." |
| reset | not existing | "No sandbox found. Run `tiklab create` first." |
| scale | not running | "Sandbox is not running. Run `tiklab start` first." |
| scale N (N > 500) | any | "Maximum user count is 500." |
| scale N (N < 1) | any | "Minimum user count is 1." |
