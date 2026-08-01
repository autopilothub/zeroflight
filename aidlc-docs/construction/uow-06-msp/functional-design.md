# UoW-06 — MSP RAW_IMU

## Goal

Provide raw accelerometer, gyro, and magnetometer data via a secondary MSP serial channel, since INAV MAVLink does not expose raw IMU.

## Components

| Package | Role |
|---------|------|
| `internal/msp` | MSP v1 encode/decode, serial client, poller |
| `internal/session` | Merge MAVLink + MSP telemetry |
| `cmd/zeroflight imu` | CLI stream |

## Config

```yaml
msp:
  enabled: false
  device: /dev/ttyUSB0
  baud: 115200
  poll_hz: 10
```

## Acceptance

- [x] `MSP_RAW_IMU` request/parse
- [x] Background poller when `msp.enabled`
- [x] `status` and `/api/v1/status` expose `raw_imu`
- [x] `zeroflight imu` CLI command
