# ZeroFlight

Raspberry Pi companion app for autonomous flight with **Mamba F405 MK2** running **INAV** over **MAVLink**.

## Features (Phase 1)

- MAVLink telemetry: `ATTITUDE`, `GPS_RAW_INT`, `GLOBAL_POSITION_INT`, `HEARTBEAT`, `SYS_STATUS`
- CLI `status` — live telemetry dashboard
- CLI `goto` — `MAV_CMD_DO_REPOSITION` for INAV GCS NAV mode

## Hardware

| Item | Setting |
|------|---------|
| FC | Mamba F405 MK2 |
| Firmware | INAV 8.0+ |
| Link | UART3 (TX3/RX3) ↔ RPi GPIO14/15 |
| Baud | 115200 |
| INAV Port | UART3 → MAVLink ON |

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
GOOS=linux GOARCH=arm64 go build -o zeroflight ./cmd/zeroflight
```

## Documentation

상세 사용법은 [docs/](docs/) 디렉터리를 참고하세요.

- [사용법](docs/usage.md) — 설치, CLI, 비행 절차
- [하드웨어 설정](docs/hardware.md) — 배선, INAV 설정
- [설정 파일](docs/configuration.md) — `inav.yaml` 항목

## Usage

```bash
# live telemetry
./zeroflight status

# one-shot snapshot
./zeroflight status --once

# override connection
./zeroflight --connection serial:/dev/ttyAMA0:115200 status

# send goto (GCS NAV must be active, GPS 3D fix, armed)
./zeroflight goto --lat 37.5665000 --lon 126.9780000 --alt 15 --wait
```

## INAV goto prerequisites

1. GPS 3D fix (8+ satellites recommended)
2. Armed
3. POS HOLD
4. **GCS NAV** switch ON (MAVLink `custom_mode` shows as GUIDED)

## Project layout

```
cmd/zeroflight/     CLI entrypoint
internal/inav/      INAV MAVLink adapter
internal/config/    YAML configuration
pkg/geo/            distance / bearing helpers
configs/inav.yaml   default settings
```

## Limitations

INAV MAVLink is **partial** — no raw IMU over MAVLink, no parameter API, mission upload is limited. See project planning docs for Phase 2+ (MSP raw IMU, mission manager, REST API).

## Safety

- Configure INAV failsafe (RX loss → RTH)
- Test on the bench without propellers first
- Keep RC manual override available at all times
