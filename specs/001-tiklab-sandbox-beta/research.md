# Research: TikLab Sandbox Beta

**Branch**: `001-tiklab-sandbox-beta` | **Date**: 2026-03-16

## R1: RouterOS CHR in Docker

**Decision**: Use `evilfreelancer/docker-routeros:7` as the base image, customized with the TikLab behavior engine.

**Rationale**: This is the most widely used community image for running RouterOS in Docker for development and testing (500+ GitHub stars). It wraps RouterOS CHR in QEMU inside a Docker container, exposing all standard management ports. The image supports RouterOS v7 tags, which aligns with our version target.

**Alternatives considered**:
- `mikrotik/chr` (official): Smaller image, no QEMU layer. Requires `--privileged` and custom bridge networking. Less community documentation for automated setups. Better for production routing, less flexible for embedding a behavior engine.
- `vrnetlab/mikrotik_routeros`: Designed for ContainerLab network topologies. Much larger (~985 MB). Overkill for single-router sandbox use.
- Custom Dockerfile from scratch: Maximum control but significant effort to replicate QEMU+network bridge setup. Not justified for Beta.

**Key details**:
- Image size: ~86 MB
- Boot time: 30–60 seconds for RouterOS to become API-reachable after container start
- Requires: `NET_ADMIN` capability, `/dev/net/tun` device, `/dev/kvm` (optional, for hardware acceleration on Linux amd64)
- Exposed ports: SSH (22), API (8728), API-SSL (8729), Winbox (8291)

## R2: RouterOS CHR Licensing

**Decision**: Use the free CHR license tier (permanent, no registration required).

**Rationale**: The free tier limits throughput to 1 Mbps per interface. Since TikLab's simulated traffic does not need to reach the internet and the purpose is API/management testing, 1 Mbps is sufficient. The traffic needs to be visible in RouterOS's accounting and queue statistics, not to achieve high throughput.

**Alternatives considered**:
- P1 license ($45, 1 Gbps): Unnecessary for Beta scope. Can be offered as an option in future releases for high-throughput testing.
- 60-day trial: Requires MikroTik.com account registration, which violates the zero-friction deployment principle.

## R3: Go RouterOS API Library

**Decision**: Use `github.com/go-routeros/routeros/v3`.

**Rationale**: Most actively maintained Go library for the RouterOS API protocol. v3.0.1 released February 2025, supports Go 1.21+. Provides `Run` (execute commands), `Listen` (subscribe to events), and `Tab` (read tabular data). MIT licensed. Can configure all RouterOS services: DHCP, Hotspot, simple queues, user management.

**Alternatives considered**:
- `github.com/jda/routeros-api-go`: Inactive since 2016. Not recommended.
- `github.com/swoga/go-routeros`: Fork with some activity but smaller community.
- Direct API protocol implementation: Unnecessary given the mature library.

## R4: Docker SDK for Go

**Decision**: Use `github.com/docker/docker/client` with API version negotiation.

**Rationale**: The standard Docker SDK for Go. Provides all container lifecycle operations needed: `ImagePull`, `ContainerCreate`, `ContainerStart`, `ContainerStop`, `ContainerRemove`, `ContainerList`, `ContainerInspect`. Used by Docker CLI itself, Kubernetes, and most Go-based container tools.

**Key pattern**:
```go
cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
```

## R5: CLI Framework

**Decision**: Use `github.com/spf13/cobra`.

**Rationale**: De facto standard Go CLI framework. Used by Docker, Kubernetes, GitHub CLI, Hugo. Provides command/subcommand structure, flag parsing, help generation, and shell completion. Well-documented with extensive ecosystem support.

## R6: Behavior Engine Architecture

**Decision**: The behavior engine runs as a separate Go binary inside the Docker container, connected to the RouterOS QEMU instance via an internal virtual network bridge. The host CLI communicates with it via a lightweight HTTP control API on port 9090.

**Rationale**: Running the engine inside the container allows it to operate on the same L2 network as RouterOS, which is required for acting as real DHCP clients and Hotspot users. A control API is cleaner than `docker exec` for structured commands (scale up/down, stop, status).

**Alternatives considered**:
- Engine on host connecting via exposed ports: Cannot act as DHCP clients on the RouterOS internal network. Would require complex NAT/bridge configurations.
- Engine as a separate sidecar container: Adds Docker Compose dependency and networking complexity. Not worth it for Beta.
- Communication via `docker exec`: Fragile, platform-dependent behavior, no structured responses.

## R7: Reset Mechanism

**Decision**: API-based reset (fast path). The CLI uses the RouterOS API to wipe configuration and re-apply the initial setup, then signals the behavior engine to regenerate users. No container restart required.

**Rationale**: Container restart with QEMU reboot takes 30–60 seconds (RouterOS boot time alone). API-based config wipe and re-apply completes in seconds. The behavior engine regenerates users with fresh random identities. Total reset time well under 30 seconds.

**Alternatives considered**:
- Container destroy + recreate: Too slow (image boot time). Violates SC-005 (< 30 seconds).
- Docker checkpoint/restore: Not widely supported, especially on Windows/macOS Docker Desktop.
- RouterOS `/system reset-configuration`: Reboots the router, adding boot time. Slower than API-based wipe.

## R8: Traffic Generation Approach

**Decision**: The behavior engine creates virtual network interfaces (one per simulated user) on the internal bridge, performs DHCP requests, authenticates via Hotspot HTTP portal, then generates traffic according to the assigned profile.

**Traffic patterns by profile**:
- **Idle (40%)**: Periodic ICMP ping (every 30s) + DNS query (every 60s). ~1 KB/min.
- **Standard (45%)**: HTTP GET requests to a built-in HTTP server inside the container, simulating small page loads. Periodic small file downloads (10–100 KB). ~50–200 KB/min.
- **Heavy (15%)**: Continuous TCP stream to/from a built-in traffic sink, simulating downloads or streaming. ~500 KB–1 MB/min (constrained by queue limits).

**Rationale**: Real L2 traffic ensures RouterOS processes it through its actual DHCP server, Hotspot authentication, firewall, and queue system. This makes management interface data (leases, sessions, queue stats) genuinely accurate — not injected or faked.
