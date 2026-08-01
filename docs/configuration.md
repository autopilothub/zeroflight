# 설정 파일

ZeroFlight는 YAML 설정 파일을 사용합니다. 기본 경로는 `configs/inav.yaml` 입니다.

## 파일 위치

```bash
./zeroflight --config /path/to/inav.yaml status
```

`--config` 를 생략하면 실행 디렉터리 기준 `configs/inav.yaml` 을 읽습니다.

---

## 전체 예시

```yaml
mavlink:
  connection: "serial:/dev/serial0:115200"
  device: "/dev/serial0"
  baud: 115200
  target_system_id: 1
  target_component_id: 1
  out_system_id: 255
  out_component_id: 190

safety:
  max_altitude_m: 120
  max_radius_m: 500
  arrival_radius_m: 3
  arrival_altitude_m: 2
```

---

## mavlink 섹션

### connection

MAVLink 연결 문자열. CLI `--connection` 으로 덮어쓸 수 있습니다.

| 형식 | 예시 |
|------|------|
| 시리얼 | `serial:/dev/serial0:115200` |
| UDP | `udp:127.0.0.1:14550` |

```yaml
connection: "serial:/dev/serial0:115200"
```

### device / baud

`connection` 이 비어 있을 때 사용하는 시리얼 fallback 값입니다. 일반적으로 `connection` 만 설정하면 충분합니다.

```yaml
device: "/dev/serial0"
baud: 115200
```

USB 어댑터 예:

```yaml
connection: "serial:/dev/ttyUSB0:115200"
```

### target_system_id / target_component_id

명령을 보낼 FC의 MAVLink ID. INAV 기본값:

```yaml
target_system_id: 1
target_component_id: 1
```

INAV CLI `mavlink_sysid` 와 일치해야 합니다.

### out_system_id / out_component_id

ZeroFlight(RPi)가 사용하는 송신 ID:

```yaml
out_system_id: 255
out_component_id: 190
```

FC와 겹치지 않는 값이면 됩니다.

---

## safety 섹션

자율 명령 전 검증 및 도착 판정에 사용됩니다.

### max_altitude_m

`goto` 명령의 최대 허용 고도 (m). 초과 시 명령 거부.

```yaml
max_altitude_m: 120
```

`0` 으로 설정하면 제한 없음 (권장하지 않음).

### max_radius_m

홈 포인트(`GPS_GLOBAL_ORIGIN`) 기준 최대 비행 반경 (m). `goto`, `hover`, `mission upload`에 적용됩니다.

```yaml
max_radius_m: 500
```

### arrival_radius_m

`goto --wait` / `hover --wait` 시 수평 도착 판정 반경 (m).

```yaml
arrival_radius_m: 3
```

### arrival_altitude_m

`goto --wait` / `hover --wait` 시 고도 도착 판정 허용 오차 (m).

```yaml
arrival_altitude_m: 2
```

### link_timeout_sec

MAVLink 텔레메트리 갱신 타임아웃 (초). 초과 시 자율 명령 거부 또는 `--wait` 중단.

```yaml
link_timeout_sec: 3
```

---

## api 섹션

### listen

HTTP API 바인드 주소. `zeroflight serve` 에서 사용.

```yaml
api:
  listen: "127.0.0.1:8080"
```

LAN에서 접근하려면 `0.0.0.0:8080` (방화벽 주의).

---

## msp 섹션

MAVLink(UART6)와 별도 UART에서 MSP `RAW_IMU`를 폴링합니다. `zeroflight imu` 및 `status`의 Raw IMU 섹션에 사용됩니다.

```yaml
msp:
  enabled: false
  device: "/dev/ttyUSB0"
  baud: 115200
  poll_hz: 10
```

| 항목 | 설명 |
|------|------|
| `enabled` | `true`이면 시작 시 MSP 시리얼 연결 및 백그라운드 폴링 |
| `device` | MSP UART 장치 경로 (UART6은 MAVLink 전용) |
| `baud` | 시리얼 속도 (INAV Ports 탭과 일치) |
| `poll_hz` | RAW_IMU 요청 주기 (Hz) |

INAV Configurator → **Ports** 에서 해당 UART에 **MSP** 를 활성화하세요.

---

## 환경별 설정 예시

### Raspberry Pi + UART6 (실기)

```yaml
mavlink:
  connection: "serial:/dev/serial0:115200"
```

### USB 시리얼 디버그

```yaml
mavlink:
  connection: "serial:/dev/ttyUSB0:57600"
```

### UDP (개발용)

```yaml
mavlink:
  connection: "udp:127.0.0.1:14550"
```

---

## CLI와 설정 우선순위

1. CLI `--connection` (최우선)
2. 설정 파일 `mavlink.connection`
3. 설정 파일 `device` + `baud` fallback

예:

```bash
# 설정 파일 무시하고 시리얼 지정
./zeroflight --connection serial:/dev/serial0:115200 status

# 설정 파일의 connection 사용
./zeroflight status
```

---

## 관련 문서

- [사용법](usage.md)
- [하드웨어 설정](hardware.md)
