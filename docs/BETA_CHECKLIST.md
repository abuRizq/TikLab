# MikroTik Sandbox Beta Acceptance Checklist

## Prerequisites

- [ ] Go 1.22+
- [ ] Docker installed and running
- [ ] QEMU (`qemu-system-x86_64`, `qemu-img`) in PATH
- [ ] Linux host (for full network simulation)
- [ ] CHR base image in `~/.mikrotik-sandbox/base/` (chr.qcow2 or chr.img)

## 1. API Compatibility

- [ ] Run `sandbox create` (with base image present)
- [ ] Run `sandbox start`
- [ ] Wait for CHR to boot (~30–60s)
- [ ] Run `sandbox validate` (or `sandbox validate --skip-reset`)
- [ ] API port 8728: OK
- [ ] Execute CRUD via MikroTik API (e.g. `/system/resource/print`)
- [ ] Verify API client can connect with admin/empty password

## 2. Port Connectivity

- [ ] API: `localhost:8728` — TCP connect succeeds
- [ ] Winbox: `localhost:8291` — TCP connect succeeds
- [ ] SSH: `localhost:2222` — TCP connect succeeds (maps to guest :22)

## 3. Traffic Realism

- [ ] Run `sandbox scale-users 300` (Linux only)
- [ ] Traffic container picks up namespaces (check logs)
- [ ] 40/45/15 distribution: idle (ping+DNS), browsing (k6/curl), heavy (iperf3)
- [ ] Traffic visible in RouterOS interfaces (if CHR on bridge)
- [ ] Queue behavior observable (if queues configured)

## 4. DHCP & Hotspot (Manual)

- [ ] Configure DHCP server on CHR (192.168.88.0/24 pool)
- [ ] Configure Hotspot on CHR
- [ ] Run `sandbox scale-users N` and verify DHCP leases
- [ ] Verify Hotspot login flow (MAC-cookie, HTTP-PAP)

## 5. Stability

- [ ] Sandbox sustains 300 concurrent users (isp_small)
- [ ] No container crashes under load
- [ ] Network namespaces stable

## 6. Reset Performance

- [ ] Run `sandbox reset`
- [ ] Reset completes in **< 2 seconds**
- [ ] CHR returns to post-create state
- [ ] Run `sandbox validate` to measure reset timing

## 7. Lifecycle

- [ ] `sandbox create` — succeeds
- [ ] `sandbox start` — succeeds
- [ ] `sandbox reset` — succeeds, < 2s
- [ ] `sandbox destroy` — succeeds
- [ ] `sandbox destroy --all` — removes workdir
- [ ] `sandbox create --force` — overwrites existing

## 8. Scale

- [ ] `sandbox scale-users 0` — removes all namespaces
- [ ] `sandbox scale-users 100` — adds 100 namespaces
- [ ] `sandbox scale-users 300` — scales to 300
- [ ] Traffic container restarts and picks up new namespaces

## Quick Validation Command

```bash
sandbox validate
```

Runs: port checks, API CRUD, reset performance (if not `--skip-reset`).

## Success Criteria Summary

| Criterion | Target |
|-----------|--------|
| API compatibility | CRUD works via API |
| Ports | 8728, 8291, 2222 reachable |
| Traffic | 40/45/15 distribution |
| Stability | 300 users sustained |
| Reset | < 2 seconds |
