# Vision Document — ZeroFlight

**Status:** Approved (retroactive)  
**Gate:** Inception — Mob Elaboration

---

## 1. Problem statement

FPV/네비게이션 드론에 Raspberry Pi companion computer를 추가하여, INAV 비행 컨트롤러의 센서 데이터를 수신하고 사용자 명령에 따라 자율 비행(위치 이동, 미션)을 수행하고자 한다.

기존 INAV + RC 조종만으로는 프로그래밍 가능한 고수준 미션 실행이 어렵다. ZeroFlight는 **RPi에서 고수준 자율 로직**을, **FC에서 저수준 자세·모터 제어**를 담당하는 분리 아키텍처를 목표로 한다.

---

## 2. Target users

| Persona | Need |
|---------|------|
| 드론 개발자 | Go로 companion 앱을 확장·디버깅 |
| 운용자 | CLI/API로 goto·미션 실행 |
| 연구자 | 텔레메트리 로깅·자율 알고리즘 실험 |

---

## 3. MVP scope (in)

- [x] MAVLink 텔레메트리 수신 (attitude, GPS, battery, mode)
- [x] CLI `status` — 실시간 상태 표시
- [x] CLI `goto` — INAV GCS NAV + `MAV_CMD_DO_REPOSITION`
- [x] YAML 설정 (`/dev/serial0`, UART6)
- [x] 한국어 사용 문서 (`docs/`)
- [ ] goto 도착 판정·preflight 강화 (UoW-02)
- [ ] 웨이포인트 미션 업로드·실행 (UoW-03)
- [ ] Geofence·안전 검증 (UoW-04)
- [ ] REST API + RPi 배포 (UoW-05)

---

## 4. Out of scope (explicit no)

| Item | Rationale |
|------|-----------|
| RPi에서 직접 모터 PWM 제어 | 안전·인증; FC에 위임 |
| ArduPilot/PX4 완전 호환 | INAV partial MAVLink |
| Raw IMU over MAVLink only | INAV 미지원; MSP는 Phase 2+ |
| MAVLink 파라미터 튜닝 | INAV Configurator만 |
| 클라우드 GCS (QGC 대체) | companion 로컬 CLI 우선 |
| 비전·장애물 회피 | 향후 별도 unit |

---

## 5. Success criteria

1. RPi에서 `./zeroflight status` 로 attitude/GPS 10Hz급 수신
2. GCS NAV 활성 상태에서 50m goto 후 호버
3. 3개 이상 웨이포인트 순차 비행 (UoW-03)
4. RC 손실 시 INAV failsafe RTH 동작 유지

---

## 6. Hardware context (fixed)

| Component | Spec |
|-----------|------|
| FC | Mamba F405 MK2 |
| Firmware | INAV 8.0+ |
| Link | UART6 (TX6/RX6) ↔ RPi GPIO14/15 |
| Serial device | `/dev/serial0` @ 115200 |
| Protocol | MAVLink v2 |

---

## 7. Open questions (resolved)

| Question | Decision |
|----------|----------|
| UART port | UART6 (user confirmed) |
| Serial device | `/dev/serial0` (user confirmed) |
| Control protocol | MAVLink primary |
| Language | Go |

---

## Approval

- [x] **Approve and Continue** — Inception requirements baseline
