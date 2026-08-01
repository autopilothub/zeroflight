# Units of Work — ZeroFlight

**Status:** Approved  
**Gate:** Inception — Units Generation

> AI-DLC terminology: **Units of Work** replace Epics; **Bolts** replace Sprints (hours–days).

---

## Overview

| Unit | Name | Bolt estimate | Depends on |
|------|------|---------------|------------|
| **UoW-01** | MAVLink telemetry & status CLI | 1–2 days | — |
| **UoW-02** | GCS NAV goto hardening | 2–3 days | UoW-01 |
| **UoW-03** | Mission upload & execution | 3–4 days | UoW-02 |
| **UoW-04** | Safety & geofence | 2 days | UoW-02 |
| **UoW-05** | REST API & RPi deployment | 2 days | UoW-03 |

---

## UoW-01: MAVLink telemetry & status CLI ✅

**Stories:** US-01  
**Status:** Complete

### Deliverables

- [x] `internal/inav` client (gomavlib)
- [x] Telemetry parsers (attitude, GPS, heartbeat)
- [x] `zeroflight status` CLI
- [x] `configs/inav.yaml`
- [x] Unit tests (geo, modes)
- [x] `docs/` usage documentation

### Construction artifacts

`aidlc-docs/construction/uow-01-telemetry/`

---

## UoW-02: GCS NAV goto hardening ⏳ NEXT

**Stories:** US-02, US-03  
**Status:** Partial (basic goto exists)

### Deliverables

- [ ] `hover` subcommand (current position)
- [ ] Improved preflight messages (mode-specific)
- [ ] COMMAND_ACK handling (if INAV sends)
- [ ] Integration test script (bench checklist)
- [ ] Field test runbook update

### Bolt B-02 tasks

1. Add `hover` command
2. Log goto commands to file
3. Validate GPS fix degradation during `--wait`
4. Document GCS NAV activation sequence

### Construction artifacts

`aidlc-docs/construction/uow-02-goto/`

---

## UoW-03: Mission upload & execution

**Stories:** US-04

### Deliverables

- [ ] `internal/mission` package
- [ ] MISSION_COUNT / MISSION_ITEM upload
- [ ] `zeroflight mission upload` / `start`
- [ ] Disarm-only upload guard
- [ ] INAV-compatible waypoint format

### Risks

- INAV partial mission protocol
- Heading/POI not via MAVLink

---

## UoW-04: Safety & geofence

**Stories:** US-05, NFR-02

### Deliverables

- [ ] `internal/safety` geofence check
- [ ] Home-relative radius limit
- [ ] MAVLink link heartbeat timeout
- [ ] Preflight checklist command
- [ ] Telemetry CSV logger

---

## UoW-05: REST API & deployment

### Deliverables

- [ ] `internal/api` REST (`/status`, `/goto`, `/mission`)
- [ ] systemd unit file
- [ ] Cross-compile Makefile or script
- [ ] GitHub Actions CI (`go test`, `go build`)

---

## Backlog (future units)

| Unit | Description |
|------|-------------|
| UoW-06 | MSP RAW_IMU secondary channel |
| UoW-07 | Orbit path generator |
| UoW-08 | Web GCS dashboard |

---

## Parallelization

- UoW-04 (safety) can start in parallel with UoW-03 after UoW-02 gate
- UoW-05 depends on stable command surface from UoW-03

---

## Approval

- [x] **Approve and Continue** — proceed to Construction UoW-02
