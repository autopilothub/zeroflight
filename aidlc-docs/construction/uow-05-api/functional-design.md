# UoW-05 — Functional Design: REST API & Deployment

**Unit:** UoW-05  
**Status:** ✅ Implemented

---

## Scope

Long-lived MAVLink session, HTTP API, RPi deployment artifacts.

## Architecture

```text
zeroflight serve
  └── internal/service (persistent inav.Client)
        └── internal/api (net/http)
```

## API endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Liveness |
| GET | `/api/v1/status` | Vehicle telemetry JSON |
| GET | `/api/v1/preflight` | Checklist |
| POST | `/api/v1/goto` | `{"lat","lon","alt","force"}` |
| POST | `/api/v1/hover` | `{"alt","force"}` |
| POST | `/api/v1/mission/upload` | `{"waypoints":[...]}` |
| POST | `/api/v1/mission/clear` | Clear mission |

## Config

```yaml
api:
  listen: "127.0.0.1:8080"
```

## Deployment

- `deploy/zeroflight.service` — systemd
- `Makefile` — `build-pi`, `install-service`
- `.github/workflows/ci.yml` — test + arm64 build
- `docs/deployment.md` — RPi install guide

## Approval

- [x] **Approve and Continue** — MVP complete
