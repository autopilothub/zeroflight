# Technical Environment — ZeroFlight

**Status:** Approved  
**Gate:** Inception

---

## 1. Language & runtime

| Item | Value |
|------|-------|
| Language | Go |
| Version | 1.25+ (gomavlib v3 requirement) |
| Module | `github.com/autopilothub/zeroflight` |
| Target OS | Linux (Raspberry Pi OS, arm64) |
| Dev OS | macOS / Linux |

---

## 2. Dependencies (approved)

| Package | Purpose | Notes |
|---------|---------|-------|
| `github.com/bluenviron/gomavlib/v3` | MAVLink v2 | ardupilotmega dialect |
| `github.com/spf13/cobra` | CLI | subcommands |
| `go.bug.st/serial` | Serial I/O | `/dev/serial0` |
| `gopkg.in/yaml.v3` | Config | `configs/inav.yaml` |

---

## 3. Prohibited / avoid

| Avoid | Reason | Alternative |
|-------|--------|-------------|
| ArduPilot-only MAVLink APIs | INAV incompatible | INAV adapter layer |
| Direct motor control from RPi | Safety | MAVLink/INAV nav modes |
| `tinygo` for FC timing | Not needed on RPi | Standard Go on Pi |
| Hardcoded `/dev/ttyAMA0` | Pi model variance | `/dev/serial0` |

---

## 4. Project structure

```text
cmd/zeroflight/          # CLI entrypoint
internal/
  inav/                  # INAV MAVLink adapter
  config/                # YAML loader
pkg/geo/                 # Haversine, bearing
configs/inav.yaml        # Runtime config
docs/                    # User documentation
aidlc-docs/              # AI-DLC artifacts (not runtime)
```

---

## 5. Code patterns

### INAV client interface

```go
type Client struct { /* gomavlib node + state */ }
func (c *Client) State() VehicleState
func (c *Client) SendGoto(req GotoRequest) error
```

### Config loading

```go
cfg, _ := config.Load("configs/inav.yaml")
clientCfg, _ := cfg.INAVConfig(connectionOverride)
```

### CLI command structure

```go
root.AddCommand(newStatusCmd())
root.AddCommand(newGotoCmd())
```

---

## 6. Testing

| Type | Tool | Location |
|------|------|----------|
| Unit | `go test` | `pkg/geo`, `internal/inav` |
| Integration | Bench + FC | Manual, propellers off |
| Field | Outdoor GPS | Manual checklist |

---

## 7. Deployment model

| Stage | Target |
|-------|--------|
| Dev | Mac/Linux cross-compile |
| Test | RPi on bench (UART6) |
| Ops | RPi systemd service (UoW-05) |

No AWS cloud in MVP. Operations phase covers RPi edge deployment only.

---

## 8. Security baseline

- No secrets in git
- RC manual override always available
- INAV failsafe independent of RPi
- Geofence before autonomous commands (UoW-04)

---

## Approval

- [x] **Approve and Continue**
