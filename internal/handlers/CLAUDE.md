# internal/handlers — HTTP 핸들러

## 주요 파일

| 파일 | 역할 |
|------|------|
| `display.go` | `/display` 예배 슬라이드 HTML + `/display/overlay` 가사 오버레이 + 비디오 배경 `<video>` 요소 |
| `displayConfigHandlers.go` | `/api/display-config` GET/PUT — 폰트·오버레이·비디오 배경 설정 |
| `displayStageHandlers.go` | `/display/stage` 무대 모니터 HTML (현재 슬라이드 + 다음 항목 + 경과 타이머) |
| `videoBgHandlers.go` | 비디오 배경 업로드·목록·삭제·서빙 (`data/video-bg/`) |
| `apiHandlers.go` | `/submit`, `/api/user`, `/api/worship-order` 등 REST API |
| `websocket.go` | `/ws` WebSocket — 진행상황, display 상태, 스케줄 카운트다운 |
| `scheduler.go` | 예배 시간 자동 감지 → 카운트다운 → 순서 로드 → OBS 스트리밍 |
| `mobileRemote.go` | `/mobile` PWA 리모컨 (인라인 HTML, Stitch 디자인) |
| `obsSourceHandlers.go` | OBS 소스 관리 API (EP_PDF 소스명 분기 포함) |
| `templateHandlers.go` | 예배 템플릿 관리 API |
| `pdfHandlers.go` | 외부 PDF 업로드·Ghostscript 변환·슬라이드 제어·OBS Browser Source HTML |

## Display 시스템

- `/display` — 프로젝터용 (찬송 이미지, 성경 3절 페이징, 교독 이미지)
- `/display/overlay` — 방송 자막용 (반투명 배경, `lyricsMap` 가사 매핑, 오버레이 커스터마이징 CSS 변수 적용)
- `/display/stage` — 무대 모니터용 (현재 슬라이드 대형 표시 + 우측 다음 항목 패널 + 하단 경과 타이머)
- 상태 영속화: `data/display_state.json`

## DisplayConfig (`displayConfigHandlers.go`)

`data/display_config.json`에 저장. 모든 Display 시각 설정을 단일 구조체로 관리:

| 필드 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| `font` | string | `"default"` | Display 폰트 (Google Fonts 로드) |
| `overlayBgOpacity` | float64 | `0.75` | 오버레이 배경 투명도 (0.0~1.0) |
| `overlayTextColor` | string | `"#ffffff"` | 오버레이 텍스트 색상 |
| `overlayPosition` | string | `"flex-end"` | 오버레이 위치 (flex-end/center/flex-start) |
| `overlayFontScale` | float64 | `1.0` | 오버레이 폰트 크기 배율 (0.5~2.0) |
| `globalVideoBg` | string | `""` | 전역 비디오 배경 파일명 |

## 비디오 배경 (`videoBgHandlers.go`)

| 엔드포인트 | 메서드 | 설명 |
|-----------|--------|------|
| `/api/video-bg/upload` | POST | MP4/WebM/MOV 업로드 (최대 300MB) |
| `/api/video-bg/list` | GET | 업로드된 파일 목록 `[{filename, url}]` |
| `/api/video-bg/delete` | DELETE | `?filename=xxx` 파일 삭제 |
| `/display/video-bg/{name}` | GET | 파일 서빙 |

- 저장 경로: `data/video-bg/`
- Display HTML이 시작 시 `/api/display-config`에서 `globalVideoBg` 읽어 `<video>` 요소에 적용

## 스케줄러

- 설정: `data/schedule.json`
- 기본: 주일 11:00 / 오후 14:00 / 수요 19:30 / 금요 20:30
- T-N분: WS `schedule_countdown` → T-0초: 순서 로드 + OBS 시작

## 모바일 리모컨

- `/mobile` — Stitch 디자인, Inter 폰트, WS 동기화
- `/mobile/qr.png` — 로컬 IP 자동 감지 QR 코드
- PWA 설치 지원 (manifest.json, sw.js)

## 외부 PDF (pdfHandlers.go)

| 엔드포인트 | 메서드 | 설명 |
|-----------|--------|------|
| `/api/pdf/upload` | POST | PDF 업로드 + Ghostscript PNG 변환 (최대 50MB) |
| `/api/pdf/slides` | GET | 슬라이드 상태 조회 `{count, currentIndex}` |
| `/api/pdf/slides` | DELETE | 슬라이드 초기화 |
| `/api/pdf/navigate` | POST | 슬라이드 이동 `{action: prev|next|goto, index}` |
| `/display/pdf` | GET | OBS Browser Source HTML (1920×1080, 폴링 1초) |
| `/display/pdf-slides/{NNN}.png` | GET | 슬라이드 이미지 서빙 |

- **전제 조건**: Ghostscript 설치 필요 (Windows: `gswin64c` / `gswin32c` / `gs`)
- OBS 소스명: `EP_PDF` (EP_Display 충돌 방지를 위해 별도 분기)
- 슬라이드 이미지 저장 경로: `data/pdf-slides/`
