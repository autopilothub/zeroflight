# UoW-03 — Functional Design: Mission Upload

**Unit:** UoW-03  
**Status:** ✅ Implemented  
**Gate:** Construction — approved

---

## Scope

Upload YAML-defined waypoint missions to INAV via MAVLink mission protocol.

## Protocol flow

```text
GCS                          INAV
 │  MISSION_CLEAR_ALL    →    │
 │  MISSION_COUNT        →    │
 │  ← MISSION_REQUEST(0)      │
 │  MISSION_ITEM(0)      →    │
 │  ← MISSION_REQUEST(1)      │
 │  MISSION_ITEM(1)      →    │
 │  ← MISSION_ACK             │
```

## Mission file

`internal/mission/plan.go` — YAML loader with validation.

## CLI

```bash
zeroflight mission upload --file path.yaml
zeroflight mission clear
```

## Constraints

- Disarmed only (INAV rejects while armed)
- `MAV_FRAME_GLOBAL` per INAV wiki
- `MAV_CMD_NAV_WAYPOINT` for each item

## Files

| File | Purpose |
|------|---------|
| `internal/mission/plan.go` | YAML mission plan |
| `internal/inav/mission.go` | MAVLink upload/clear |
| `cmd/zeroflight/mission.go` | CLI |
| `configs/example-mission.yaml` | Example |

## Approval

- [x] **Approve and Continue** — UoW-03 closed
