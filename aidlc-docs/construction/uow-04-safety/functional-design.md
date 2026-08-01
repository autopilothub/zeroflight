# UoW-04 — Functional Design: Safety & Geofence

**Unit:** UoW-04  
**Status:** ✅ Implemented

---

## Scope

Geofence validation, preflight checklist, link timeout, CSV telemetry logging.

## Components

| Package | Responsibility |
|---------|----------------|
| `internal/safety` | Geofence, preflight, link stale check |
| `internal/log` | CSV telemetry writer |
| `inav` | `HomePosition` from `GPS_GLOBAL_ORIGIN` |

## Geofence rules

- Home from INAV `GPS_GLOBAL_ORIGIN`
- `max_radius_m` — horizontal distance from home
- `max_altitude_m` — relative altitude
- Applied to `goto`, `hover`, `mission upload`

## CLI

```bash
zeroflight preflight [--require-pass]
zeroflight log telemetry -o logs/telemetry.csv [--interval 1s] [--duration 5m]
```

## Config

```yaml
safety:
  max_radius_m: 500
  link_timeout_sec: 3
```

## Approval

- [x] **Approve and Continue** — UoW-04 closed
