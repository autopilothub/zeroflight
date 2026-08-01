# ZeroFlight

Raspberry Pi companion app for autonomous flight with **Mamba F405 MK2** running **INAV** over **MAVLink**.

## Features (MVP)

- MAVLink telemetry + `status` CLI
- `goto` / `hover` (GCS NAV + `DO_REPOSITION`)
- `mission upload` / `clear`
- Geofence, `preflight`, CSV `log telemetry`
- REST API + **Web GCS dashboard** (`serve`)
- MSP raw IMU (`imu`), orbit path (`orbit`)

## Hardware

| Item | Setting |
|------|---------|
| FC | Mamba F405 MK2 |
| Firmware | INAV 8.0+ |
| Link | UART6 (TX6/RX6) ↔ RPi GPIO14/15 |
| Baud | 115200 |
| INAV Port | UART6 → MAVLink ON |

### INAV CLI (recommended)

```bash
set mavlink_extra1_rate = 20
set mavlink_pos_rate = 10
set mavlink_version = 2
save
```

Enable **GCS NAV** mode on a transmitter switch before using `goto`.

## Build

```bash
# on Raspberry Pi
go build -o zeroflight ./cmd/zeroflight

# cross-compile from Mac/Linux
make build-pi
```

## Documentation

상세 사용법은 [docs/](docs/) 디렉터리를 참고하세요.

- [사용법](docs/usage.md) — 설치, CLI, 비행 절차
- [하드웨어 설정](docs/hardware.md) — 배선, INAV 설정
- [설정 파일](docs/configuration.md) — `inav.yaml` 항목
- [배포](docs/deployment.md) — RPi, systemd, REST API

개발 산출물 및 진행 상태: [aidlc-docs/aidlc-state.md](aidlc-docs/aidlc-state.md)

## Usage

```bash
# live telemetry
./zeroflight status

# one-shot snapshot
./zeroflight status --once

# override connection
./zeroflight --connection serial:/dev/serial0:115200 status

# send goto (GCS NAV must be active, GPS 3D fix, armed)
./zeroflight goto --lat 37.5665000 --lon 126.9780000 --alt 15 --wait

# hold current position, change altitude
./zeroflight hover --alt 15 --wait

# upload waypoint mission (disarmed)
./zeroflight mission upload -f configs/example-mission.yaml

# preflight checklist
./zeroflight preflight

# log telemetry to CSV
./zeroflight log telemetry -o logs/telemetry.csv --duration 5m

# HTTP API server
./zeroflight serve
curl http://127.0.0.1:8080/api/v1/status
```

## INAV goto prerequisites

1. GPS 3D fix (8+ satellites recommended)
2. Armed
3. POS HOLD
4. **GCS NAV** switch ON (MAVLink `custom_mode` shows as GUIDED)

## Project layout

```
cmd/zeroflight/       CLI + serve
internal/inav/        MAVLink adapter
internal/api/         REST API + dashboard routes
internal/web/         embedded GCS UI
internal/service/     long-lived session
internal/safety/      geofence, preflight
configs/inav.yaml     settings
deploy/               systemd unit
```

## Limitations

INAV MAVLink is **partial** — no parameter API over MAVLink. Raw IMU via optional MSP UART. Web GCS at `http://<host>:8080/` when running `serve`.

## Safety

- Configure INAV failsafe (RX loss → RTH)
- Test on the bench without propellers first
- Keep RC manual override available at all times
