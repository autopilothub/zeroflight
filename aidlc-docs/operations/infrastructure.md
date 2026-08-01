# Operations — ZeroFlight (Edge Deployment)

**Phase:** Operations  
**Status:** ✅ Implemented  
**Gate:** Approved

---

## 1. Deployment target

| Item | Value |
|------|-------|
| Platform | Raspberry Pi 4/5 (arm64) |
| OS | Raspberry Pi OS (64-bit) |
| Binary | `/usr/local/bin/zeroflight` |
| Config | `/etc/zeroflight/inav.yaml` |
| User | `pi` or dedicated `zeroflight` |
| Serial | `/dev/serial0`, group `dialout` |

---

## 2. Infrastructure (edge, not AWS cloud)

```text
LiPo → BEC 5V → RPi
LiPo → FC (Mamba F405 MK2)
RPi UART6 ↔ FC UART6
RC → FC UART1
GPS → FC
```

No AWS resources in MVP. Future optional:

- CloudWatch via IoT Core (backlog)
- S3 telemetry upload (backlog)

---

## 3. systemd service (draft)

```ini
[Unit]
Description=ZeroFlight INAV Companion
After=network.target

[Service]
Type=simple
User=pi
ExecStart=/usr/local/bin/zeroflight --config /etc/zeroflight/inav.yaml serve
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

---

## 4. CI/CD (planned)

| Step | Tool |
|------|------|
| Test | GitHub Actions `go test ./...` |
| Build | `GOOS=linux GOARCH=arm64 go build` |
| Release | GitHub Releases artifact |
| Deploy | Manual `scp` to RPi (MVP) |

---

## 5. Runbook

| Event | Action |
|-------|--------|
| No telemetry | Check UART6, `/dev/serial0`, INAV MAVLink port |
| Goto denied | Verify GCS NAV switch |
| RPi crash | FC failsafe RTH (INAV config) |
| Low battery | INAV RTH; RPi logs battery from SYS_STATUS |

---

## 6. Operations gates

- [x] Infrastructure design approved
- [x] systemd unit provided
- [x] CI pipeline configured
- [x] Deployment documentation

---

## Approval

- [x] **Approve and Continue**
