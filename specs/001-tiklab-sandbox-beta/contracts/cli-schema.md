# CLI Contract: TikLab Sandbox Beta

**Branch**: `001-tiklab-sandbox-beta` | **Date**: 2026-03-16

## Binary

**Name**: `tiklab`
**Global flags**: `--version`, `--help`

## Commands

### `tiklab create`

Provisions a new sandbox environment. Does not start it.

**Arguments**: None
**Flags**: None (Beta scope)

**stdout** (success):
```
Pulling image tiklab/sandbox:0.1.0...
Creating sandbox...
Sandbox created.

  SSH:    localhost:2222
  API:    localhost:8728
  Winbox: localhost:8291

Run `tiklab start` to activate.
```

**stderr** (errors):
```
Error: Sandbox already exists. Run `tiklab destroy` first.
Error: Docker is not running. Please start Docker and try again.
Error: Port 8728 is already in use. Free the port or configure an alternative.
```

**Exit codes**: `0` success, `1` error

---

### `tiklab start`

Activates a created sandbox: boots RouterOS, applies configuration, starts traffic generation.

**Arguments**: None
**Flags**: None (Beta scope)

**stdout** (success):
```
Starting sandbox...
Waiting for RouterOS to boot...
Configuring DHCP server...
Configuring Hotspot...
Starting traffic generation (50 users)...
Ready.

  SSH:    ssh admin@localhost -p 2222
  API:    localhost:8728
  Winbox: localhost:8291
```

**stderr** (errors):
```
Error: No sandbox found. Run `tiklab create` first.
Error: Sandbox is already running.
Error: RouterOS failed to boot within timeout. Run `tiklab destroy` and try again.
```

**Exit codes**: `0` success, `1` error

---

### `tiklab scale <count>`

Adjusts the number of active simulated users in a running sandbox.

**Arguments**:
| Arg | Type | Required | Description |
|-----|------|----------|-------------|
| count | int | yes | Target user count (1–500) |

**Flags**: None

**stdout** (success):
```
Scaling to 200 users...
Scaled to 200 users.
```

**stderr** (errors):
```
Error: Sandbox is not running. Run `tiklab start` first.
Error: Maximum user count is 500.
Error: Minimum user count is 1.
Error: Invalid user count. Provide a number between 1 and 500.
```

**Exit codes**: `0` success, `1` error

---

### `tiklab reset`

Resets the sandbox to its original post-creation state. Wipes all user-made changes, regenerates simulated users with fresh random identities. Does not restart the container.

**Arguments**: None
**Flags**: None

**stdout** (success):
```
Resetting sandbox...
Clearing configuration...
Reapplying initial setup...
Regenerating users (50 users)...
Reset complete.
```

**stderr** (errors):
```
Error: Sandbox is not running. Run `tiklab start` first.
Error: No sandbox found. Run `tiklab create` first.
```

**Exit codes**: `0` success, `1` error

---

### `tiklab destroy`

Completely removes the sandbox environment. Stops the container if running, removes all Docker artifacts.

**Arguments**: None
**Flags**: None

**stdout** (success):
```
Destroying sandbox...
Sandbox destroyed.
```

**stderr** (errors):
```
Error: No sandbox found. Nothing to destroy.
```

**Exit codes**: `0` success, `1` error

---

## Behavior Engine Control API

Internal HTTP API exposed on container port 9090. Used by the CLI to control the behavior engine. Not part of the user-facing contract. Intentionally unauthenticated — the control API is local to the Docker container network and not exposed to external networks.

### `POST /scale`

```json
{ "count": 200 }
```

**Response** (200):
```json
{ "status": "ok", "activeUsers": 200 }
```

### `POST /stop`

Stops all traffic generation.

**Response** (200):
```json
{ "status": "stopped", "activeUsers": 0 }
```

### `POST /start`

Starts traffic generation with the specified user count.

```json
{ "count": 50 }
```

**Response** (200):
```json
{ "status": "running", "activeUsers": 50 }
```

### `GET /status`

**Response** (200):
```json
{
  "status": "running",
  "activeUsers": 50,
  "profiles": {
    "idle": 20,
    "standard": 23,
    "heavy": 7
  }
}
```
