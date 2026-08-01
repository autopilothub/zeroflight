# UoW-02 — Functional Design: GCS NAV Goto

**Unit:** UoW-02  
**Status:** 🔄 In progress  
**Gate:** Construction — pending approval

---

## Scope

Harden `goto` command for field use; add `hover`; improve operator feedback.

## Current implementation

| Feature | File | Status |
|---------|------|--------|
| `SendGoto` (DO_REPOSITION) | `internal/inav/commands.go` | ✅ |
| `PreflightGoto` | `internal/inav/commands.go` | ✅ |
| CLI `goto` | `cmd/zeroflight/main.go` | ✅ |
| Arrival `--wait` | `cmd/zeroflight/main.go` | ✅ |

## Planned additions

### hover command

Send goto to current GPS position with optional altitude change.

```bash
zeroflight hover --alt 15
```

### Enhanced preflight

| Check | Action if fail |
|-------|----------------|
| Mode != GCS_NAV | Print "Enable GCS NAV switch" |
| Mode == POS_HOLD but not GCS_NAV | Hint two-step activation |
| HDOP > 2.0 | Warn, optional `--force` |

### Goto audit log

Append JSON lines to `logs/goto.jsonl`:

```json
{"ts":"...","lat":37.56,"lon":126.97,"alt":15,"mode":"GCS_NAV"}
```

## INAV protocol notes

- `MAV_CMD_DO_REPOSITION`, frame `MAV_FRAME_GLOBAL`
- WP#255, requires `BOXGCSNAV`
- param4 yaw: 0 = keep heading

## Files to modify

| File | Change |
|------|--------|
| `cmd/zeroflight/main.go` | add `hover` subcommand |
| `internal/inav/commands.go` | optional force flag, ack wait |
| `internal/inav/log.go` | goto audit log (new) |
| `docs/usage.md` | hover, troubleshooting |

## Questions

See `questions.md` if gate requires clarification.

## Approval

- [ ] **Request Changes**
- [ ] **Approve and Continue** — proceed to implementation plan
