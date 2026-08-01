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
ls -l /dev/serial0
```

`/dev/serial0` 은 Raspberry Pi의 기본 UART 심볼릭 링크입니다 (모델에 따라 `ttyAMA0` 또는 `ttyS0`를 가리킴).

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
RPi GPIO14 (TXD)  →  FC RX6
RPi GPIO15 (RXD)  →  FC TX6
GND               →  GND
```

INAV Configurator → **Ports** 탭에서 **UART6** 에 **MAVLink** 를 켜고 baud rate를 **115200** 으로 맞춥니다.

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
| 시리얼 | `serial:/dev/serial0:115200` | RPi ↔ FC (실기) |
| UDP | `udp:127.0.0.1:14550` | 개발/시뮬레이터 |

```bash
# 연결 덮어쓰기 예시
./zeroflight --connection serial:/dev/serial0:115200 status
./zeroflight --connection udp:192.168.1.10:14550 status
```

### 명령 목록

| 명령 | 설명 |
|------|------|
| `status` | 텔레메트리 실시간 표시 |
| `goto` | 목표 좌표로 이동 (`MAV_CMD_DO_REPOSITION`) |
| `hover` | 현재 GPS 위치 유지 (고도 변경 가능) |
| `mission` | 웨이포인트 미션 업로드/삭제 |
| `preflight` | 자율비행 사전 점검 체크리스트 |
| `log telemetry` | CSV 텔레메트리 기록 |
| `imu` | MSP RAW_IMU 실시간 표시 (MSP UART 필요) |
| `orbit` | 원형 경로 순차 goto |
| `serve` | HTTP REST API 서버 |

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
| Home | 홈 위치 (GPS_GLOBAL_ORIGIN) |
| Raw IMU (MSP) | 가속도/자이로/자기 (MSP 활성화 시) |

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
| `--force` | | `false` | HDOP 높을 때 강행 |
| `--log` | | `logs/goto.jsonl` | 감사 로그 경로 |

### HDOP 검사

GPS HDOP가 2.0을 초과하면 명령이 거부됩니다. 긴급 시:

```bash
./zeroflight goto --lat 37.5665 --lon 126.9780 --alt 15 --force
```

`--force` 사용 시 경고가 stderr에 출력되고 `logs/goto.jsonl`에 `forced: true`로 기록됩니다.

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

## 5.1 hover — 현재 위치 유지

현재 GPS 좌표로 `DO_REPOSITION`을 전송해 위치를 고정합니다. 고도만 변경할 수 있습니다.

### 기본 사용

```bash
# 현재 위치·고도 유지
./zeroflight hover

# 현재 위치에서 고도 15m로 변경
./zeroflight hover --alt 15

# 도착(고도 안정)까지 대기
./zeroflight hover --alt 15 --wait
```

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `--alt` | `0` (현재 고도 유지) | 목표 상대 고도 (m) |
| `--wait` | `false` | 도착 판정까지 대기 |
| `--timeout` | `2m` | `--wait` 타임아웃 |
| `--force` | `false` | HDOP 높을 때 강행 |
| `--log` | `logs/goto.jsonl` | 감사 로그 경로 |

사전 조건은 `goto`와 동일합니다 (armed, GPS 3D fix, GCS NAV on).

---

## 5.3 orbit — 원형 경로 비행

현재 GPS 위치(또는 지정 좌표)를 중심으로 원형 웨이포인트를 생성하고, 순차적으로 `goto`를 전송합니다.

```bash
# 현재 위치 중심, 반경 50m, 8개 웨이포인트, 고도 15m
./zeroflight orbit --radius 50 --points 8 --alt 15

# 특정 좌표를 중심으로
./zeroflight orbit --lat 37.5665 --lon 126.9780 --center-here=false --radius 30 --alt 10
```

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `--lat` / `--lon` | `0` | 궤도 중심 (기본: 현재 GPS) |
| `--center-here` | `true` | 현재 GPS를 중심으로 사용 |
| `--radius` | `50` | 반경 (m) |
| `--points` | `8` | 웨이포인트 개수 |
| `--alt` | `10` | 목표 상대 고도 (m) |
| `--wait` | `true` | 각 웨이포인트 도착 대기 |
| `--timeout` | `2m` | 웨이포인트별 대기 타임아웃 |
| `--force` | `false` | HDOP 높을 때 강행 |

---

## 5.4 imu — MSP 원시 IMU

INAV MAVLink는 raw IMU를 제공하지 않습니다. 별도 UART에 **MSP**를 활성화하면 `MSP_RAW_IMU`를 폴링합니다.

`configs/inav.yaml`:

```yaml
msp:
  enabled: true
  device: "/dev/ttyUSB0"   # UART3 등 여분 UART
  baud: 115200
  poll_hz: 10
```

INAV Configurator → **Ports** 에서 해당 UART에 **MSP** 를 켭니다 (UART6은 MAVLink용).

```bash
./zeroflight imu
./zeroflight imu --once
./zeroflight imu --interval 50ms
```

---

## 5.5 mission — 웨이포인트 미션

YAML 파일로 웨이포인트를 INAV에 업로드합니다. **disarmed** 상태에서만 가능합니다.

### 미션 파일 형식

```yaml
# configs/example-mission.yaml
waypoints:
  - lat: 37.5665000
    lon: 126.9780000
    alt: 15
  - lat: 37.5668000
    lon: 126.9783000
    alt: 15
```

### 업로드

```bash
./zeroflight mission upload --file configs/example-mission.yaml
```

### 삭제

```bash
./zeroflight mission clear
```

### 비행 절차

1. 지상에서 `mission upload` (disarmed)
2. 이륙 후 **POS HOLD**
3. RC에서 **MISSION** 모드 전환
4. INAV가 웨이포인트 순서대로 비행

**주의:** INAV MAVLink 미션은 Configurator 대비 제한적입니다 (heading 등). 문제 시 Configurator Mission Control 사용.

---

## 5.3 preflight — 사전 점검

goto/mission 전에 연결·GPS·GCS NAV·geofence를 한 번에 확인합니다.

```bash
./zeroflight preflight
./zeroflight preflight --require-pass   # 실패 시 exit code 1
```

출력 예:

```text
[ OK ] MAVLink link — ok
[ OK ] GPS 3D fix — fix type 3
[FAIL] GCS NAV active — in POS HOLD — enable GCS NAV switch
```

---

## 5.4 log telemetry — CSV 기록

```bash
# 1초 간격으로 기록 (Ctrl+C 종료)
./zeroflight log telemetry -o logs/telemetry.csv

# 5분간 기록
./zeroflight log telemetry -o logs/flight.csv --interval 500ms --duration 5m
```

`goto`/`hover` 실행 시 `max_radius_m` geofence와 `max_altitude_m` 제한이 자동 적용됩니다.

---

## 6. serve — HTTP API & Web GCS

MAVLink 연결을 유지한 채 REST API와 **웹 GCS 대시보드**를 제공합니다.

```bash
./zeroflight serve
./zeroflight serve --listen 0.0.0.0:8080
```

브라우저에서 `http://127.0.0.1:8080/` (또는 `--listen` 주소)로 접속합니다.

### Web GCS 대시보드 (`/`)

| 기능 | 설명 |
|------|------|
| 텔레메트리 | 1초마다 status 폴링 (배터리, GPS, attitude) |
| 지도 | 홈 기준 상대 위치 SVG |
| Preflight | 체크리스트 실행 |
| Goto / Hover | 폼에서 명령 전송 |

외부 CDN 없이 동작하므로 필드에서도 RPi 로컬 네트워크로 사용할 수 있습니다.

### REST API

| Method | Path | 설명 |
|--------|------|------|
| GET | `/health` | 헬스체크 |
| GET | `/api/v1/status` | 텔레메트리 JSON |
| GET | `/api/v1/preflight` | 사전 점검 |
| POST | `/api/v1/goto` | `{"lat","lon","alt","force"}` |
| POST | `/api/v1/hover` | `{"alt","force"}` |
| POST | `/api/v1/mission/upload` | `{"waypoints":[...]}` |
| POST | `/api/v1/mission/clear` | 미션 삭제 |

RPi 배포는 [deployment.md](deployment.md) 참고.

---

## 7. 비행 절차 예시

### 7.1 벤치 테스트 (프로펠러 제거)

1. FC에 INAV 플래시, UART6 MAVLink 설정
2. RPi 배선 후 전원 인가
3. `./zeroflight status --once` 로 연결 확인
4. 야외 또는 창가에서 GPS fix 확인

### 7.2 첫 goto 비행

1. **사전 점검:** 나침반·가속도계 캘리브레이션, failsafe(RTH) 설정
2. 이륙 후 **POS HOLD** 로 안정 호버
3. `status` 로 GPS 3D fix, armed 확인
4. RC에서 **GCS NAV** 스위치 ON → `GCS NAV: on` 확인
5. 가까운 목표(10~20m)로 goto:

```bash
./zeroflight goto --lat <목표위도> --lon <목표경도> --alt 10 --wait
```

6. 이상 시 RC로 GCS NAV OFF 또는 RTH

### 7.3 권장 비행 순서

```text
지상 점검 → 이륙(수동) → POS HOLD → GCS NAV ON → goto → 호버 → GCS NAV OFF → 착륙(수동)
```

---

## 8. 문제 해결

### `timed out waiting for INAV telemetry`

| 원인 | 조치 |
|------|------|
| 배선 TX/RX 반대 | TX↔RX 교차 연결 확인 |
| baud 불일치 | FC와 설정 모두 115200 |
| UART6 MAVLink 미설정 | INAV Ports 탭 확인 |
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

## 9. systemd 서비스 (선택)

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

## 10. 제한 사항

INAV MAVLink는 **부분 구현**입니다.

| 기능 | 지원 |
|------|------|
| Attitude, GPS 텔레메트리 | ✅ |
| goto (`DO_REPOSITION`) | ✅ (GCS NAV 필요) |
| Raw IMU (gyro/accel/mag) | ✅ MSP 보조 UART (`zeroflight imu`) |
| 원형 경로 (`orbit`) | ✅ 순차 goto |
| 파라미터 원격 튜닝 | ❌ Configurator/CLI만 |
| MAVLink takeoff/land | ❌ RC로 수행 |
| 완전한 미션 프로토콜 | ⚠️ 제한적 (향후 Phase) |

---

## 11. 관련 문서

- [하드웨어 설정](hardware.md)
- [설정 파일](configuration.md)
- [프로젝트 README](../README.md)
