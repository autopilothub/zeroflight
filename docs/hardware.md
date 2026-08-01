# 하드웨어 설정

Mamba F405 MK2 + INAV + Raspberry Pi 연결 가이드입니다.

## 구성 요소

| 부품 | 역할 |
|------|------|
| Mamba F405 MK2 | 비행 컨트롤러 (INAV) |
| Raspberry Pi | Companion computer (ZeroFlight 실행) |
| GPS 모듈 | FC에 연결 (RPi가 아닌 FC 경유) |
| RC 수신기 | 수동 조종, 모드 전환, failsafe |
| BEC 5V | RPi 전원 (FC 5V 핀 직결 비권장) |

---

## UART 포트 배치 (Mamba F405 MK2)

| UART | 패드 | 일반 용도 | ZeroFlight |
|------|------|-----------|------------|
| UART0 | USB | Configurator | 개발용 |
| UART1 | SBUS / TX1 | RC 수신기 | RC 전용 (MAVLink 사용 안 함) |
| UART3 | TX3 / RX3 | VTX 등 | **RPi MAVLink 연결** |
| UART6 | TX6 / RX6 | ESC 텔레메트리 | (미사용 시 대안) |

RC 수신기는 UART1에 두고, RPi는 **UART3** 에 연결하는 구성을 권장합니다.

---

## 배선

### RPi ↔ FC (UART3)

```
Raspberry Pi          Mamba F405 MK2
─────────────         ───────────────
GPIO14 (TXD)    →     RX3
GPIO15 (RXD)    ←     TX3
GND             ─     GND
```

- **3.3V TTL** — 레벨 시프터 불필요
- TX와 RX는 **교차** 연결
- GND는 반드시 공통

### 전원

- FC: LiPo 배터리
- RPi: 별도 BEC 5V (2A 이상 권장)
- FC의 5V pad로 RPi를 급전하면 전압 강하·노이즈 문제가 생길 수 있음

---

## Raspberry Pi 시리얼 설정

### 1. UART 활성화

`/boot/firmware/config.txt`:

```ini
enable_uart=1
```

### 2. Serial console 비활성화 (필요 시)

```bash
sudo raspi-config
# Interface Options → Serial Port
# Login shell over serial: No
# Serial port hardware enabled: Yes
```

### 3. 장치 확인

```bash
ls -l /dev/serial0
# /dev/serial0 -> ttyAMA0
```

ZeroFlight 기본 연결:

```yaml
connection: "serial:/dev/ttyAMA0:115200"
```

USB 시리얼 어댑터를 쓰는 경우 `/dev/ttyUSB0` 등으로 변경합니다.

---

## INAV Configurator 설정

### Ports 탭

| UART | Serial RX | Telemetry | MSP | MAVLink | Baud |
|------|-----------|-----------|-----|---------|------|
| UART1 | ON (RC) | (RX에 따라) | OFF | OFF | 115200 |
| UART3 | OFF | OFF | OFF | **ON** | **115200** |

**Save and Reboot** 후 적용합니다.

### Modes 탭

다음 모드를 RC 스위치에 할당합니다.

| 모드 | 용도 |
|------|------|
| POS HOLD | 호버 / goto 전제 |
| RTH | 복귀 (failsafe 포함) |
| **GCS NAV** | **goto 명령 수신** (필수) |
| MISSION | (향후) 웨이포인트 비행 |

### Configuration 탭

- **GPS** 활성화
- **Magnetometer** 활성화 (내장 또는 외장)
- **Barometer** 활성화

### 캘리브레이션

1. Accelerometer — 수평에서 6면 캘리브레이션
2. Magnetometer — 금속·전자기 간섭 없는 야외에서
3. GPS — 야외에서 3D fix 확인

---

## INAV CLI 권장 설정

Configurator **CLI** 탭 또는 시리얼 CLI:

```bash
# MAVLink v2
set mavlink_version = 2

# 텔레메트리 주기 상향 (기본 2~3Hz → 자율비행용)
set mavlink_extra1_rate = 20    # ATTITUDE
set mavlink_pos_rate = 10       # GPS, GLOBAL_POSITION_INT
set mavlink_ext_status_rate = 2 # SYS_STATUS

# system ID (기본값 유지 가능)
set mavlink_sysid = 1

save
```

### Failsafe (권장)

Configurator **Failsafe** 탭:

| 이벤트 | 동작 |
|--------|------|
| RC loss | RTH |
| GPS loss (비행 중) | Land 또는 RTH |
| Low battery | RTH |

---

## RC 수신기별 UART1 참고

| 수신기 | UART1 설정 |
|--------|------------|
| ExpressLRS (CRSF) | Serial RX ON, provider CRSF |
| FrSky FPort | TX1 pad, halfduplex 설정 |
| SBUS | SBUS pad (단방향) |

RC 설정은 MAVLink 포트(UART3)와 **분리**하는 것이 안정적입니다.

---

## 체크리스트

배선·설정 완료 후:

- [ ] UART3 MAVLink ON, 115200 baud
- [ ] RPi `/dev/ttyAMA0` 접근 가능
- [ ] GPS 3D fix (야외)
- [ ] Compass 캘리브레이션 완료
- [ ] GCS NAV 스위치 할당
- [ ] Failsafe → RTH 설정
- [ ] `./zeroflight status --once` 연결 확인

다음: [사용법](usage.md)
