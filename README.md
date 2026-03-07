# MikroTik Sandbox CLI

A local CLI tool that generates a realistic, containerized MikroTik RouterOS (CHR) environment for testing ISP billing systems and Hotspot managers.

## Quick Start (Docker – one command)

Works on any machine with Docker. No Go, QEMU, or Linux required on the host.

```bash
git clone https://github.com/abuRizq/TikLab.git
cd TikLab
docker compose run --rm tiklab up
```

**First run:** CHR image (~60MB) downloads automatically from MikroTik. Ports: API `8728`, Winbox `8291`, SSH `2222`

## Prerequisites

**Docker mode (recommended):** Docker only. Everything runs inside containers.

**Native mode:** Go 1.22+, Docker, `qemu-img`, Linux (for full sandbox; network namespaces required)

## CHR Base Image

Download a MikroTik CHR image from [mikrotik.com/download/chr](https://mikrotik.com/download/chr):

- `chr.qcow2` or `chr.img` (or any `.qcow2` / `.img` file)
- **Docker:** place in `./data/base/`
- **Native:** place in `~/.mikrotik-sandbox/base/`
- If you download the RAW `.img`, convert with: `qemu-img convert -f raw -O qcow2 chr-7.x.img chr.qcow2`

## Build (native)

```bash
go build -o sandbox .
```

## Test

```bash
go test -v ./...
```

Or use the Makefile: `make build` / `make test`

## Usage

**Docker (one command):**
```bash
docker compose run --rm tiklab up      # Create + start
docker compose run --rm tiklab start   # Start only
docker compose run --rm tiklab reset  # Reset
docker compose run --rm tiklab destroy # Tear down
```

**Native:**
```bash
sandbox create      # Initialize isp_small environment (requires base image)
sandbox start       # Boot the stack (Docker Compose + QEMU CHR)
sandbox up          # Create (if needed) + start in one command
sandbox reset       # Revert to post-create state (< 2s)
sandbox destroy     # Tear down
sandbox scale-users 300   # Scale active users
sandbox validate    # Run beta validation checks
```

Ports: API `8728`, Winbox `8291`, SSH `2222`

See [docs/CLI.md](docs/CLI.md) for full command reference.
