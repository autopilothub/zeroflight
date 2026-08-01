# AI-DLC 개발 방법론 — ZeroFlight

[AWS AI-DLC](https://aws.amazon.com/blogs/devops/ai-driven-development-life-cycle/)(AI-Driven Development Life Cycle)를 ZeroFlight 프로젝트에 적용한 가이드입니다.

---

## AI-DLC란?

AI가 계획·설계·코드를 **제안**하고, 사람이 **게이트(gate)** 에서 검증·승인하는 개발 방법론입니다.

| 기존 Agile | AI-DLC |
|------------|--------|
| Epic | **Unit of Work (UoW)** |
| Sprint | **Bolt** (수 시간~수 일) |
| 인라인 질문 | **Question → Doc → Approval** |

### 3단계

```text
INCEPTION  →  CONSTRUCTION  →  OPERATIONS
(무엇을)       (어떻게 구현)     (배포·운영)
```

---

## 프로젝트 구조

```text
zeroflight/
├── .cursor/rules/ai-dlc-workflow.mdc   # Cursor AI-DLC 규칙
├── .aidlc-rule-details/                # AWS 공식 워크플로우 스테이지
├── aidlc-docs/                         # AI-DLC 산출물 (런타임 코드 아님)
│   ├── aidlc-state.md                  # ★ 현재 진행 상태
│   ├── inception/                      # 요구사항·설계
│   ├── construction/                 # Unit별 설계·구현 계획
│   └── operations/                     # 배포·운영
├── cmd/ internal/ pkg/                 # 실제 애플리케이션 코드
└── docs/                               # 사용자 문서
```

---

## 시작하기

### 1. 새 Bolt / Unit 시작

Cursor에서:

```text
AI-DLC 워크플로우를 따라 UoW-02 구현을 시작해줘.
aidlc-docs/aidlc-state.md 와 construction/uow-02-goto/ 를 읽고 진행해.
```

### 2. 세션 재개 (컨텍스트 리셋 후)

```text
Go to aidlc-docs/aidlc-state.md, find the first unchecked item,
then go to the corresponding plan file and resume from that point.
```

### 3. 탐색만 (문서 변경 없음)

```text
Do not update any documents. INAV goto가 GCS NAV 없이 동작할 수 있는지 설명해줘.
```

### 4. 질문 답변 후 진행

AI가 `aidlc-docs/**/questions.md` 를 생성하면 `[Answer]:` 를 채운 뒤:

```text
We have answered your clarification questions. Please re-read the file and proceed.
```

---

## 게이트 승인

각 스테이지 완료 시 AI는 다음 중 하나를 기다립니다.

- **Request Changes** — 수정 요청
- **Approve and Continue** — 다음 단계 진행

승인 전에 생성된 마크다운 산출물을 반드시 읽으세요.

---

## ZeroFlight Units of Work

| Unit | 내용 | 상태 |
|------|------|------|
| UoW-01 | 텔레메트리 & status CLI | ✅ 완료 |
| UoW-02 | goto 강화 & hover | ⏳ 다음 |
| UoW-03 | 미션 업로드 | 대기 |
| UoW-04 | 안전·geofence | 대기 |
| UoW-05 | REST API & 배포 | 대기 |

상세: [aidlc-docs/inception/units-of-work.md](../aidlc-docs/inception/units-of-work.md)

---

## Cursor 규칙

`.cursor/rules/ai-dlc-workflow.mdc` 가 `alwaysApply: true` 로 설정되어 있어, 소프트웨어 개발 요청 시 AI-DLC 워크플로우를 따릅니다.

AWS 공식 규칙 상세: `.aidlc-rule-details/`

---

## 참고 링크

| 자료 | URL |
|------|-----|
| AWS AI-DLC 블로그 | https://aws.amazon.com/blogs/devops/ai-driven-development-life-cycle/ |
| aidlc-workflows (GitHub) | https://github.com/awslabs/aidlc-workflows |
| Working with AIDLC | https://github.com/awslabs/aidlc-workflows/blob/main/docs/WORKING-WITH-AIDLC.md |
| Method Definition Paper | https://prod.d13rzhkk8cj2z0.amplifyapp.com/ |

---

## 관련 문서

- [사용법](usage.md) — CLI·하드웨어
- [aidlc-state.md](../aidlc-docs/aidlc-state.md) — 진행 상태
