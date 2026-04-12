# internal/handlers — HTTP 핸들러

## 주요 파일

| 파일 | 역할 |
|------|------|
| `display.go` | `/display` 예배 슬라이드 HTML + `/display/overlay` 가사 오버레이 |
| `apiHandlers.go` | `/submit`, `/api/user`, `/api/worship-order` 등 REST API |
| `websocket.go` | `/ws` WebSocket — 진행상황, display 상태, 스케줄 카운트다운 |
| `scheduler.go` | 예배 시간 자동 감지 → 카운트다운 → 순서 로드 → OBS 스트리밍 |
| `mobileRemote.go` | `/mobile` PWA 리모컨 (인라인 HTML, Stitch 디자인) |
| `obsSourceHandlers.go` | OBS 소스 관리 API |
| `templateHandlers.go` | 예배 템플릿 관리 API |

## Display 시스템

- `/display` — 프로젝터용 (찬송 이미지, 성경 3절 페이징, 교독 이미지)
- `/display/overlay` — 방송 자막용 (반투명 배경, `lyricsMap` 가사 매핑)
- 상태 영속화: `data/display_state.json`

## 스케줄러

- 설정: `data/schedule.json`
- 기본: 주일 11:00 / 오후 14:00 / 수요 19:30 / 금요 20:30
- T-N분: WS `schedule_countdown` → T-0초: 순서 로드 + OBS 시작

## 모바일 리모컨

- `/mobile` — Stitch 디자인, Inter 폰트, WS 동기화
- `/mobile/qr.png` — 로컬 IP 자동 감지 QR 코드
- PWA 설치 지원 (manifest.json, sw.js)
