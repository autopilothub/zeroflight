# Raspberry Pi 배포

## 1. 크로스 컴파일

개발 PC에서:

```bash
make build-pi
scp zeroflight pi@raspberrypi.local:/tmp/
```

RPi에서:

```bash
sudo install -m 755 /tmp/zeroflight /usr/local/bin/zeroflight
```

## 2. 설정

```bash
sudo mkdir -p /etc/zeroflight
sudo cp configs/inav.yaml /etc/zeroflight/inav.yaml
```

`/etc/zeroflight/inav.yaml` 에서 `api.listen` 을 필요에 따라 변경:

```yaml
api:
  listen: "0.0.0.0:8080"   # LAN 접근 허용 시
```

## 3. serial 권한

```bash
sudo usermod -aG dialout pi
```

재로그인 후 `/dev/serial0` 접근 가능.

## 4. systemd 서비스

```bash
make install-service
sudo systemctl enable zeroflight
sudo systemctl start zeroflight
sudo systemctl status zeroflight
```

로그:

```bash
journalctl -u zeroflight -f
```

## 5. API 확인

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/api/v1/status
curl http://127.0.0.1:8080/api/v1/preflight
```

### goto 예시

```bash
curl -X POST http://127.0.0.1:8080/api/v1/goto \
  -H 'Content-Type: application/json' \
  -d '{"lat":37.5665,"lon":126.9780,"alt":15}'
```

### 미션 업로드 예시

```bash
curl -X POST http://127.0.0.1:8080/api/v1/mission/upload \
  -H 'Content-Type: application/json' \
  -d '{"waypoints":[{"lat":37.5665,"lon":126.9780,"alt":15}]}'
```

## 6. 수동 실행 (디버그)

```bash
./zeroflight --config configs/inav.yaml serve --listen 127.0.0.1:8080
```

## 관련 문서

- [사용법](usage.md)
- [하드웨어 설정](hardware.md)
