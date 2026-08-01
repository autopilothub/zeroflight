# UoW-07 — Orbit path generator

## Goal

Fly a circular path by generating waypoints around a center and issuing sequential `goto` commands.

## Components

| Package | Role |
|---------|------|
| `pkg/geo/orbit.go` | `CirclePoints` great-circle waypoint generator |
| `cmd/zeroflight orbit` | CLI command |

## Acceptance

- [x] `geo.CirclePoints` unit test
- [x] `zeroflight orbit` with `--radius`, `--points`, `--alt`, `--wait`
