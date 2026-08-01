# Application Design — ZeroFlight

**Status:** Approved  
**Gate:** Inception — Application Design (Mob Elaboration)

---

## 1. System context

```text
┌─────────────┐     UART6/MAVLink      ┌──────────────────┐
│ Raspberry Pi│◄──────────────────────►│ Mamba F405 MK2 │
│ ZeroFlight  │     /dev/serial0       │ INAV             │
│  (Go CLI)   │                        │ Gyro/Accel/Mag   │
└──────┬──────┘                        │ GPS, Motors      │
       │                               └────────┬─────────┘
       │ CLI / REST (future)                    │ RC
       ▼                                        ▼
   Operator                              Transmitter
```

---

## 2. Components

| Component | Responsibility | Package |
|-----------|----------------|---------|
| **CLI** | User commands, output | `cmd/zeroflight` |
| **Config** | YAML load, connection parse | `internal/config` |
| **INAV Client** | MAVLink connect, telemetry, commands | `internal/inav` |
| **Geo** | Distance, bearing, arrival | `pkg/geo` |
| **Safety** (future) | Geofence, preflight | `internal/safety` |
| **Mission** (future) | Waypoint builder, upload | `internal/mission` |
| **API** (future) | REST handlers | `internal/api` |

---

## 3. Data flow

### Telemetry (downlink)

```text
FC ──MAVLink──► gomavlib Node ──► applyFrame() ──► VehicleState ──► CLI display
```

Messages: `HEARTBEAT`, `ATTITUDE`, `GPS_RAW_INT`, `GLOBAL_POSITION_INT`, `SYS_STATUS`

### Goto (uplink)

```text
CLI goto ──► PreflightGoto() ──► SendGoto() ──► COMMAND_INT (DO_REPOSITION) ──► FC WP#255
```

Precondition: `GCSNavActive == true`, GPS fix ≥ 3, armed

---

## 4. Mode state machine (INAV)

```text
MANUAL/STABILIZE ──► POS_HOLD ──► GCS_NAV (GUIDED) ──► goto accepted
                      │                │
                      └──── RTH ◄──────┘
```

ZeroFlight reads `HEARTBEAT.custom_mode` → `FlightMode` enum.

---

## 5. Interface contracts

### VehicleState (read model)

Aggregated snapshot; thread-safe copy via `Client.State()`.

### GotoRequest (command)

```go
type GotoRequest struct {
    LatDeg, LonDeg float64
    AltM           float32
    YawDeg         *float32
}
```

### Config (file + CLI)

```yaml
mavlink.connection: "serial:/dev/serial0:115200"
safety.max_altitude_m: 120
```

---

## 6. Extension points

| Extension | Hook |
|-----------|------|
| MSP raw IMU | Second serial on UART3 or UDP |
| REST API | `internal/api` mounting same `inav.Client` |
| Mission file format | `internal/mission` YAML/JSON |

---

## Approval

- [x] **Approve and Continue**
