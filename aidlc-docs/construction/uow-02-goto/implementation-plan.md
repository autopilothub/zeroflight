# UoW-02 — Implementation Plan

**Unit:** UoW-02  
**Bolt:** B-02  
**Status:** Ready for gate approval

---

## Checklist

- [x] 1. Add `internal/inav/log.go` — JSONL goto audit logger
- [x] 2. Add `hover` subcommand in `cmd/zeroflight/main.go`
- [x] 3. Extend `PreflightGoto` with HDOP warning (non-blocking)
- [x] 4. Add `--force` flag to skip HDOP warning on goto/hover
- [x] 5. Update `docs/usage.md` with hover examples
- [x] 6. Update `aidlc-docs/aidlc-state.md` — mark UoW-02 complete
- [x] 7. `go test ./...` && `go build`

**Status:** ✅ Complete

---

## File plan

| # | File | Action |
|---|------|--------|
| 1 | `internal/inav/log.go` | create |
| 2 | `cmd/zeroflight/main.go` | modify — hover, --force |
| 3 | `internal/inav/commands.go` | modify — PreflightGotoWithOptions |
| 4 | `docs/usage.md` | modify |
| 5 | `aidlc-docs/aidlc-state.md` | modify |

---

## Test plan

| Test | Method |
|------|--------|
| Unit | `go test ./...` |
| Bench | `./zeroflight status --once` with FC connected |
| Goto | GCS NAV on, `./zeroflight goto --lat ... --wait` |
| Hover | `./zeroflight hover --alt 10` |

---

## Approval (required before code generation)

- [ ] **Request Changes**
- [ ] **Approve and Continue** — execute checklist
