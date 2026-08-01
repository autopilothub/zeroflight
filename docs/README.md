# ZeroFlight 문서

Mamba F405 MK2 + INAV + Raspberry Pi 기반 자율 드론 companion 앱 문서입니다.

| 문서 | 내용 |
|------|------|
| [사용법](usage.md) | 설치, 설정, CLI 명령, 비행 절차 |
| [하드웨어 설정](hardware.md) | 배선, INAV 포트/CLI 설정 |
| [설정 파일](configuration.md) | `configs/inav.yaml` 항목 설명 |
| [AI-DLC 방법론](aidlc.md) | AWS AI-DLC 개발 워크플로우 |
| [배포](deployment.md) | RPi 설치, systemd, REST API |

## 빠른 시작

```bash
# 빌드
go build -o zeroflight ./cmd/zeroflight

# 텔레메트리 확인
./zeroflight status

# goto (GCS NAV 활성, GPS 3D fix, armed)
./zeroflight goto --lat 37.5665000 --lon 126.9780000 --alt 15 --wait
```

자세한 내용은 [사용법](usage.md)을 참고하세요.
