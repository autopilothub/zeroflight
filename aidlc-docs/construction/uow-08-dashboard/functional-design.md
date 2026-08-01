# UoW-08 — Web GCS dashboard

## Goal

Browser-based ground control station served by `zeroflight serve`, using the existing REST API.

## Architecture

```text
zeroflight serve
  └── internal/api
        ├── /api/v1/*   REST API
        └── /           embedded dashboard (internal/web)
```

## Features

- Live telemetry polling (`/api/v1/status`)
- Preflight checklist (`/api/v1/preflight`)
- Goto / hover commands
- Attitude horizon, relative map, MSP raw IMU panel

## Acceptance

- [x] Embedded static UI (`go:embed`)
- [x] Served at `/` with assets under `/assets/`
- [x] API tests for dashboard routes
- [x] Documented in usage/deployment guides
