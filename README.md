# MikroTik Sandbox CLI

A local CLI tool that generates a realistic, containerized MikroTik RouterOS (CHR) environment for testing ISP billing systems and Hotspot managers.

## Prerequisites

- Go 1.22+
- Docker
- QEMU (`qemu-system-x86_64`, `qemu-img`)
- Linux (for full sandbox; network namespaces required)

## CHR Base Image

Before `sandbox create`, download a MikroTik CHR image from [mikrotik.com/download/chr](https://mikrotik.com/download/chr) and place it in `~/.mikrotik-sandbox/base/`:

- `chr.qcow2` or `chr.img` (or any `.qcow2` / `.img` file)
- If you download the RAW `.img`, convert with: `qemu-img convert -f raw -O qcow2 chr-7.x.img chr.qcow2`

## Build

```bash
go build -o sandbox .
```

## Test

```bash
go test -v ./...
```

Or use the Makefile:

```bash
make build
make test
```

## Usage

```bash
sandbox create      # Initialize isp_small environment (requires base image)
sandbox start       # Boot the stack (Docker Compose + QEMU CHR)
sandbox reset       # Revert to post-create state (< 2s)
sandbox destroy     # Tear down
sandbox scale-users 300   # Scale active users
sandbox validate    # Run beta validation checks
```

Ports: API `8728`, Winbox `8291`, SSH `2222`

See [docs/CLI.md](docs/CLI.md) for full command reference.
