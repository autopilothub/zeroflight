# Requirements — ZeroFlight

**Status:** Approved  
**Gate:** Inception — Requirements Analysis

---

## Functional requirements

### FR-01 Telemetry

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-01.1 | MAVLink HEARTBEAT 수신 (armed, mode) | Must | ✅ |
| FR-01.2 | ATTITUDE 수신 (roll/pitch/yaw + rates) | Must | ✅ |
| FR-01.3 | GPS_RAW_INT, GLOBAL_POSITION_INT 수신 | Must | ✅ |
| FR-01.4 | SYS_STATUS (battery, sensors) 수신 | Should | ✅ |
| FR-01.5 | CLI `status` 실시간 표시 | Must | ✅ |
| FR-01.6 | 텔레메트리 CSV/JSON 로깅 | Could | ⏳ UoW-04 |

### FR-02 Navigation commands

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-02.1 | `goto` — MAV_CMD_DO_REPOSITION | Must | ✅ partial |
| FR-02.2 | GCS NAV preflight 검증 | Must | ✅ |
| FR-02.3 | 도착 판정 (`--wait`) | Must | ✅ |
| FR-02.4 | yaw 지정 (`--set-yaw`) | Should | ✅ |
| FR-02.5 | 미션 업로드 (MISSION_ITEM) | Must | ⏳ UoW-03 |
| FR-02.6 | `hover` (현재 위치 reposition) | Should | ⏳ UoW-02 |
| FR-02.7 | `rtl` 소프트 트리거 | Could | ⏳ backlog |

### FR-03 Configuration

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-03.1 | YAML config file | Must | ✅ |
| FR-03.2 | `--connection` CLI override | Must | ✅ |
| FR-03.3 | serial `/dev/serial0` default | Must | ✅ |

---

## Non-functional requirements

### NFR-01 Performance

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-01.1 | Telemetry latency | < 500ms |
| NFR-01.2 | ATTITUDE rate | ≥ 10 Hz (INAV CLI tune) |
| NFR-01.3 | goto command latency | < 200ms send |

### NFR-02 Safety

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-02.1 | Max altitude guard | config `max_altitude_m` |
| NFR-02.2 | Geofence radius | config `max_radius_m` (UoW-04) |
| NFR-02.3 | No arm via software alone | RC arm required |
| NFR-02.4 | Link timeout → no orphan goto | UoW-04 |

### NFR-03 Reliability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-03.1 | FC failsafe independent | INAV RTH on RX loss |
| NFR-03.2 | RPi crash safe | FC continues POS HOLD |
| NFR-03.3 | systemd auto-restart | UoW-05 |

### NFR-04 Maintainability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-04.1 | INAV adapter isolated | `internal/inav/` |
| NFR-04.2 | Unit tests for geo/modes | `go test ./...` |
| NFR-04.3 | AI-DLC traceable artifacts | `aidlc-docs/` |

---

## INAV-specific constraints

1. GCS NAV must be active for goto (MAVLink GUIDED mode)
2. Mission upload rejected while armed
3. No RAW_IMU over MAVLink — use ATTITUDE rates only
4. No PARAM API — tune via Configurator

---

## User stories (summary)

| Story | As a… | I want… | So that… |
|-------|-------|---------|----------|
| US-01 | developer | see live telemetry | verify FC connection |
| US-02 | operator | send goto to lat/lon/alt | fly autonomously to a point |
| US-03 | operator | wait for arrival | know when mission step completes |
| US-04 | developer | upload waypoints | fly multi-point missions |
| US-05 | operator | geofence limits | prevent flyaway |

---

## Approval

- [x] **Approve and Continue**
