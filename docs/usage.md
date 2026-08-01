# 사용법

ZeroFlight는 Raspberry Pi에서 INAV 비행 컨트롤러와 MAVLink로 통신하여 텔레메트리를 수신하고, 사용자 명령(goto)을 전송하는 companion 앱입니다.

## 대상 환경

| 항목 | 값 |
|------|-----|
| FC | Mamba F405 MK2 |
| 펌웨어 | INAV 8.0 이상 |
| Companion | Raspberry Pi (권장: 4/5) |
| 통신 | MAVLink v2 (UART) |

---

## 1. 설치

### 1.1 Raspberry Pi 준비

시리얼 포트를 FC와 공유하려면 Pi의 serial console을 비활성화합니다.

`/boot/firmware/config.txt` (또는 `/boot/config.txt`):

```ini
enable_uart=1
```

재부팅 후 시리얼 장치 확인:

```bash
ls -l /dev/serial0 /dev/ttyAMA0
```

일반적으로 `/dev/serial0` → `/dev/ttyAMA0` 로 연결됩니다.

### 1.2 Go 설치

Raspberry Pi OS에서:

```bash
sudo apt update
sudo apt install -y golang
go version
```

### 1.3 ZeroFlight 빌드

프로젝트 디렉터리에서:

```bash
cd zeroflight
go build -o zeroflight ./cmd/zeroflight
```

다른 PC에서 RPi용으로 크로스 컴파일:

```bash
GOOS=linux GOARCH=arm64 go build -o zeroflight ./cmd/zeroflight
```

빌드된 바이너리를 RPi로 복사한 뒤 실행 권한을 부여합니다.

```bash
chmod +x zeroflight
```

### 1.4 설정 파일

기본 설정은 `configs/inav.yaml` 입니다. 실행 시 `--config` 로 경로를 바꿀 수 있습니다.

```bash
./zeroflight --config /home/pi/zeroflight/configs/inav.yaml status
```

설정 항목 설명은 [configuration.md](configuration.md)를 참고하세요.

---

## 2. 하드웨어 연결

배선과 INAV 포트 설정은 [hardware.md](hardware.md)에 정리되어 있습니다.

요약:

```
RPi GPIO14 (TXD)  →  FC RX3
RPi GPIO15 (RXD)  →  FC TX3
GND               →  GND
```

INAV Configurator → **Ports** 탭에서 **UART3** 에 **MAVLink** 를 켜고 baud rate를 **115200** 으로 맞춥니다.

---

## 3. CLI 개요

```bash
zeroflight [전역 옵션] <명령> [명령 옵션]
```

### 전역 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `--config` | `configs/inav.yaml` | 설정 파일 경로 |
| `--connection` | (설정 파일 값) | 연결 문자열로 설정 덮어쓰기 |

**연결 문자열 형식:**

| 형식 | 예시 | 용도 |
|------|------|------|
| 시리얼 | `serial:/dev/ttyAMA0:115200` | RPi ↔ FC (실기) |
| UDP | `udp:127.0.0.1:14550` | 개발/시뮬레이터 |

```bash
# 연결 덮어쓰기 예시
./zeroflight --connection serial:/dev/ttyAMA0:115200 status
./zeroflight --connection udp:192.168.1.10:14550 status
```

### 명령 목록

| 명령 | 설명 |
|------|------|
| `status` | 텔레메트리 실시간 표시 |
| `goto` | 목표 좌표로 이동 (`MAV_CMD_DO_REPOSITION`) |

도움말:

```bash
./zeroflight --help
./zeroflight status --help
./zeroflight goto --help
```

---

## 4. status — 텔레메트리 확인

FC와 연결된 뒤 attitude, GPS, 배터리, 센서 상태를 표시합니다.

### 기본 사용

```bash
./zeroflight status
```

0.5초마다 화면을 갱신합니다. 종료는 `Ctrl+C`.

### 한 번만 출력

```bash
./zeroflight status --once
```

### 갱신 주기 변경

```bash
./zeroflight status --interval 1s
./zeroflight status --interval 200ms
```

### 출력 항목

| 섹션 | 내용 |
|------|------|
| Connected | MAVLink 링크 상태 |
| armed / disarmed | 시동 상태 |
| Mode | INAV 비행 모드 (POS_HOLD, GCS_NAV 등) |
| GCS NAV | goto 수신 가능 여부 (`on` = GUIDED) |
| Attitude | roll, pitch, yaw (rad) 및 각속도 |
| GPS | 위경도, 고도, fix, 위성 수, HDOP, 속도 |
| Battery | 전압, 전류, 잔량 |
| Sensors | gyro/accel/mag/baro/gps/rc 건강 상태 |

### 정상 연결 확인 체크리스트

1. `Connected: true`
2. `HEARTBEAT` 수신 → Mode가 `UNKNOWN`이 아님
3. 야외에서 GPS `fix=3` (3D fix), 위성 6개 이상
4. `Sensors` 에 gyro, accel, gps 가 `true`

연결이 안 될 때는 [문제 해결](#7-문제-해결)을 참고하세요.

---

## 5. goto — 목표 지점 이동

INAV의 **GCS NAV** 모드에서 waypoint #255 로 목표 위치를 전송합니다.

### 사전 조건 (필수)

| # | 조건 | status에서 확인 |
|---|------|-----------------|
| 1 | MAVLink 연결 | `Connected: true` |
| 2 | Armed | `armed` |
| 3 | GPS 3D fix | `fix=3`, 위성 ≥ 6 |
| 4 | GCS NAV 활성 | `GCS NAV: on`, Mode: `GCS_NAV` |

**주의:** GCS NAV는 RC 송신기 스위치로 켜야 합니다. INAV Configurator에서 **Modes** 탭에 **GCS NAV** 박스를 스위치에 할당하세요.

또한 실제 비행 전에는 **POS HOLD** 모드에서 이륙·호버 상태여야 합니다.

### 기본 사용

```bash
./zeroflight goto --lat 37.5665000 --lon 126.9780000 --alt 15
```

| 옵션 | 필수 | 기본값 | 설명 |
|------|------|--------|------|
| `--lat` | ✅ | — | 목표 위도 (도) |
| `--lon` | ✅ | — | 목표 경도 (도) |
| `--alt` | | `10` | 목표 고도 (m, relative) |
| `--set-yaw` | | `false` | 명시적 yaw 설정 사용 |
| `--yaw` | | `0` | 목표 방향 (1–359°, `--set-yaw` 와 함께) |
| `--wait` | | `false` | 도착할 때까지 대기 |
| `--timeout` | | `3m` | `--wait` 타임아웃 |

### 도착 대기

```bash
./zeroflight goto \
  --lat 37.5665000 \
  --lon 126.9780000 \
  --alt 15 \
  --wait \
  --timeout 5m
```

도착 판정 (설정 파일 `safety` 섹션):

- 수평 거리 ≤ `arrival_radius_m` (기본 3m)
- 고도 오차 ≤ `arrival_altitude_m` (기본 2m)

대기 중 출력 예:

```text
  distance=12.3m alt_err=1.5m mode=GCS_NAV
  distance=3.1m alt_err=0.8m mode=GCS_NAV
arrived at target
```

### yaw 지정 (선택)

INAV는 기본적으로 heading을 유지합니다. 방향을 지정하려면:

```bash
./zeroflight goto --lat 37.5665 --lon 126.9780 --alt 15 --set-yaw --yaw 90
```

### 고도 제한

`configs/inav.yaml` 의 `safety.max_altitude_m` 을 초과하면 명령이 거부됩니다.

```text
altitude 150.0m exceeds max 120.0m
```

---

## 6. 비행 절차 예시

### 6.1 벤치 테스트 (프로펠러 제거)

1. FC에 INAV 플래시, UART3 MAVLink 설정
2. RPi 배선 후 전원 인가
3. `./zeroflight status --once` 로 연결 확인
4. 야외 또는 창가에서 GPS fix 확인

### 6.2 첫 goto 비행

1. **사전 점검:** 나침반·가속도계 캘리브레이션, failsafe(RTH) 설정
2. 이륙 후 **POS HOLD** 로 안정 호버
3. `status` 로 GPS 3D fix, armed 확인
4. RC에서 **GCS NAV** 스위치 ON → `GCS NAV: on` 확인
5. 가까운 목표(10~20m)로 goto:

```bash
./zeroflight goto --lat <목표위도> --lon <목표경도> --alt 10 --wait
```

6. 이상 시 RC로 GCS NAV OFF 또는 RTH

### 6.3 권장 비행 순서

```text
지상 점검 → 이륙(수동) → POS HOLD → GCS NAV ON → goto → 호버 → GCS NAV OFF → 착륙(수동)
```

---

## 7. 문제 해결

### `timed out waiting for INAV telemetry`

| 원인 | 조치 |
|------|------|
| 배선 TX/RX 반대 | TX↔RX 교차 연결 확인 |
| baud 불일치 | FC와 설정 모두 115200 |
| UART3 MAVLink 미설정 | INAV Ports 탭 확인 |
| Pi serial console 점유 | `enable_uart=1`, console 비활성 |
| 잘못된 장치 경로 | `ls /dev/serial*` 후 `--connection` 수정 |

### Mode가 UNKNOWN

- HEARTBEAT는 오지만 모드 매핑 실패 → INAV 버전 확인 (8.0+)
- `mavlink_version = 2` 설정 후 `save`

### `GCS NAV mode is not active`

- RC 스위치로 GCS NAV 켜기
- POS HOLD 상태인지 확인
- `status` 에서 `GCS NAV: on` 확인 후 goto 실행

### `GPS 3D fix required`

- 야외 개활지에서 대기
- GPS 모듈 배선·UART 설정 확인
- 위성 6개 이상까지 대기

### goto 후 움직이지 않음

- GCS NAV가 켜져 있는지 재확인
- 목표 고도가 현재 고도보다 충분히 다른지 확인
- INAV CLI에서 `mavlink_pos_rate` 를 10 이상으로 설정

---

## 8. systemd 서비스 (선택)

부팅 시 자동 실행 예시 `/etc/systemd/system/zeroflight.service`:

```ini
[Unit]
Description=ZeroFlight INAV companion
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/zeroflight
ExecStart=/home/pi/zeroflight/zeroflight status
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable zeroflight
sudo systemctl start zeroflight
```

텔레메트리만 상시 실행하는 예시입니다. goto는 별도 CLI/API로 호출하는 구성을 권장합니다.

---

## 9. 제한 사항

INAV MAVLink는 **부분 구현**입니다.

| 기능 | 지원 |
|------|------|
| Attitude, GPS 텔레메트리 | ✅ |
| goto (`DO_REPOSITION`) | ✅ (GCS NAV 필요) |
| Raw IMU (gyro/accel/mag) | ❌ MAVLink 미지원 |
| 파라미터 원격 튜닝 | ❌ Configurator/CLI만 |
| MAVLink takeoff/land | ❌ RC로 수행 |
| 완전한 미션 프로토콜 | ⚠️ 제한적 (향후 Phase) |

---

## 10. 관련 문서

- [하드웨어 설정](hardware.md)
- [설정 파일](configuration.md)
- [프로젝트 README](../README.md)
