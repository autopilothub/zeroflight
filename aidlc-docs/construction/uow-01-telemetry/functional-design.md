# UoW-01 — Functional Design: Telemetry & Status CLI

**Unit:** UoW-01  
**Status:** ✅ Implemented  
**Gate:** Construction — approved retroactively

---

## Scope

MAVLink v2 client for INAV; aggregate `VehicleState`; expose `zeroflight status`.

## Domain model

- `VehicleState` — aggregate root
- `Attitude`, `GPSFix`, `Battery`, `SensorHealth` — value objects
- `FlightMode` — enum from HEARTBEAT custom_mode

## MAVLink messages handled

| Message | Handler |
|---------|---------|
| HEARTBEAT | `applyHeartbeat` |
| ATTITUDE | `applyAttitude` |
| GPS_RAW_INT | `applyGPSRaw` |
| GLOBAL_POSITION_INT | `applyGlobalPosition` |
| SYS_STATUS | `applySysStatus` |

## CLI behavior

- Default refresh 500ms
- `--once` single snapshot
- Clear screen ANSI codes

## Files created

| File | Action |
|------|--------|
| `internal/inav/*.go` | create |
| `internal/config/config.go` | create |
| `cmd/zeroflight/main.go` | create |
| `pkg/geo/geo.go` | create |
| `configs/inav.yaml` | create |

## Build & test

```bash
go build -o zeroflight ./cmd/zeroflight
go test ./...
```

## Approval

- [x] **Approve and Continue** — UoW-01 closed
