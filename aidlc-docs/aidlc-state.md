# AI-DLC State — ZeroFlight

> Workflow state tracker. Update after each gate approval.

**Project:** ZeroFlight  
**Methodology:** [AWS AI-DLC](https://aws.amazon.com/blogs/devops/ai-driven-development-life-cycle/)  
**Profile:** Brownfield (Phase 1 code exists)  
**Last updated:** 2026-08-01 (UoW-06, UoW-07)

---

## Current phase

| Phase | Status | Notes |
|-------|--------|-------|
| Inception | ✅ Complete | Retroactive documentation from planning + Phase 1 |
| Construction | 🔄 In progress | UoW-06..07 done; UoW-08 backlog |
| Operations | ✅ Complete | systemd, CI, deployment docs |

---

## Inception gates

- [x] Vision document approved — `inception/vision.md`
- [x] Technical environment approved — `inception/technical-environment.md`
- [x] Requirements analysis approved — `inception/requirements.md`
- [x] Application design approved — `inception/application-design.md`
- [x] Units of work approved — `inception/units-of-work.md`

---

## Construction — Units of Work

| Unit | Name | Status | Design | Code | Test |
|------|------|--------|--------|------|------|
| UoW-01 | MAVLink telemetry & status CLI | ✅ Done | `construction/uow-01-telemetry/` | merged | `go test ./...` |
| UoW-02 | GCS NAV goto command | ✅ Done | `construction/uow-02-goto/` | merged | `go test ./...` |
| UoW-03 | Mission upload & fly | ✅ Done | `construction/uow-03-mission/` | merged | bench test |
| UoW-04 | Safety & geofence | ✅ Done | `construction/uow-04-safety/` | merged | `go test ./...` |
| UoW-05 | REST API & deployment | ✅ Done | `construction/uow-05-api/` | merged | `go test ./...` |
| UoW-06 | MSP RAW_IMU secondary serial | ✅ Done | `construction/uow-06-msp/` | merged | `go test ./...` |
| UoW-07 | Orbit path generator | ✅ Done | `construction/uow-07-orbit/` | merged | `go test ./...` |
| UoW-08 | Web GCS dashboard | ⏳ Backlog | — | — | — |

---

## Operations gates

- [x] Infrastructure design — `operations/infrastructure.md`
- [x] RPi systemd unit — `deploy/zeroflight.service`
- [x] CI pipeline — `.github/workflows/ci.yml`
- [x] Deployment guide — `docs/deployment.md`

---

## Active bolt

UoW-06 (MSP IMU) and UoW-07 (orbit) complete. Backlog: web GCS dashboard (UoW-08).

---

## Resume instruction

```text
Go to aidlc-docs/aidlc-state.md, find the first unchecked item,
then go to the corresponding plan file and resume from that point.
```
