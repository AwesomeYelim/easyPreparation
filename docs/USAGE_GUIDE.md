# easyPreparation 사용자 가이드

> 서버 시작: `make dev` → UI: `http://localhost:3000` / API: `http://localhost:8080`

---

## 목차

1. [시작하기](#1-시작하기)
2. [주보 생성](#2-주보-생성)
3. [가사 관리](#3-가사-관리)
4. [성경 조회](#4-성경-조회)
5. [Display 제어판](#5-display-제어판)
6. [자동 스케줄러](#6-자동-스케줄러)
7. [설정](#7-설정)
8. [Display 화면 & OBS 연동](#8-display-화면--obs-연동)
9. [생성 이력](#9-생성-이력)
10. [모바일 리모컨](#10-모바일-리모컨)
11. [라이선스 관리](#11-라이선스-관리)
12. [버전 업데이트](#12-버전-업데이트)
13. [개발 모드 실행](#13-개발-모드-실행)
14. [프로덕션 빌드](#14-프로덕션-빌드)
15. [Desktop 앱](#15-desktop-앱)
16. [릴리즈 방법 (개발자)](#16-릴리즈-방법-개발자)
17. [Pro 기능 상세](#17-pro-기능-상세)

---

## 1. 시작하기

### 로그인
1. 우측 상단 로그인 아이콘 클릭 → Google 계정으로 로그인
2. 로그인 후 설정/테마/교회 정보가 자동 로드됩니다

### 교회 정보 등록
1. 우측 상단 프로필 → 사이드바 열기
2. "교회 정보" 영역 클릭 → 한글명/영문명 입력 → 저장

---

## 2. 주보 생성

### 예배 순서 편집
1. **Bulletin** 탭 → 상단 드롭다운에서 예배 유형 선택
   - 주일예배 / 오후예배 / 수요예배 / 금요예배
2. 좌측 **WorshipOrder**에서 항목 태그 클릭 → 순서에 추가
3. 중간 **SelectedOrder**에서 드래그로 순서 변경, X로 삭제
4. 항목 클릭 → 우측 **Detail**에서 상세 편집
   - 찬송: 곡번호 입력
   - 성경: 책/장/절 선택
   - 기도: 인도자명 입력
   - 교회소식: 계층형 공지 편집

> 드롭다운을 전환해도 각 예배의 편집 내용은 유지됩니다.

### 주보 PDF 생성
1. Figma key/token 설정 필요 (사이드바 → 설정)
2. "예배 자료 생성하기" 클릭
3. 진행 상황이 실시간으로 표시되고, 완료 시 자동 다운로드

### Display 전송
1. "Display 전송" 클릭 → Display 창이 열리고 제어판 활성화
2. 찬송/성경 등 이미지 전처리 후 Display에 예배 순서 표시

---

## 3. 가사 관리

### 자유 곡 탭
1. **Lyrics** 탭 → "자유 곡" 탭
2. 곡 제목 입력 → 추가 → 가사 입력
3. "중복 제거" 버튼으로 반복 구절 정리
4. "전체 가사를 검색"으로 Google 자동 검색
5. "가사 기반으로 PDF 생성"으로 ZIP 다운로드

### 찬송가 검색 탭
1. "찬송가 검색" 탭 → 번호 또는 제목 검색
2. 검색 결과 클릭 → 상세 보기
3. "Display 전송"으로 악보 이미지를 Display에 추가

---

## 4. 성경 조회

1. **Bible** 탭 → 구약/신약 선택 → 책 → 장
2. 구절 클릭으로 선택, Shift+클릭으로 범위 선택
3. "Display 전송"으로 선택 구절을 Display에 추가
4. "비교" 버튼으로 2개 버전 병렬 비교

---

## 5. Display 제어판

Display 전송 후 제어판이 열립니다.

### 기본 조작
- **항목 클릭**: 해당 슬라이드로 즉시 이동
- **◀ / ▶ 버튼** 또는 **키보드 ← →**: 이전/다음 슬라이드
- **드래그**: 항목 순서 변경 (Display에 즉시 반영)
- **휴지통**: 항목 삭제

### 서브 섹션
- 찬송/성경 항목을 펼치면 페이지별 미리보기
- 미리보기 클릭으로 특정 페이지에 바로 점프

### 타이머
- Enable 토글: 자동 넘김 카운트다운
- Repeat: 항목 끝나면 다음 항목으로 자동 이동
- 속도 조절: 0.5x ~ 2x

---

## 6. 자동 스케줄러

예배 시간에 맞춰 자동으로 예배 순서를 Display에 로드하고, OBS 스트리밍을 시작합니다.

### 스케줄 설정
1. 사이드바 → **설정** → **"스트리밍 스케줄"** 클릭
2. 스케줄 설정 모달에서:
   - 각 예배의 **활성/비활성** 체크
   - **시간** 변경
   - **사전 카운트다운**: 예배 시작 몇 분 전부터 카운트다운 표시할지 (기본 5분)
   - **OBS 자동 스트리밍**: ON이면 예배 시간에 자동 방송 시작
3. "저장" 클릭 → 서버 재시작해도 설정 유지

### 기본 스케줄
| 예배 | 요일 | 시간 |
|------|------|------|
| 주일예배 | 일요일 | 11:00 |
| 오후예배 | 일요일 | 14:00 |
| 수요예배 | 수요일 | 19:30 |
| 금요예배 | 금요일 | 20:30 |

### 자동 동작 흐름
1. 예배 시간 N분 전 → **카운트다운 시작**
   - Display: 검은 배경에 예배명 + MM:SS 큰 숫자
   - 제어판: 빨간 배너에 카운트다운
2. 예배 시간 도달 → **자동 실행**
   - `config/{예배타입}.json` 순서 자동 로드
   - Display에 예배 순서 표시
   - OBS 스트리밍 시작 (ON일 때)

### 테스트 방법
스케줄 설정 모달에서 각 예배 옆 버튼:

| 버튼 | 동작 |
|------|------|
| **⏱** (시계) | 10초 카운트다운 테스트 — Display/제어판에서 카운트다운 확인 |
| **▶** (재생) | 즉시 실행 테스트 — 해당 예배 순서를 Display에 바로 로드 |

> ▶(즉시 실행)은 현재 Display 순서를 덮어씁니다. 테스트 후 필요 시 다시 전송하세요.

### OBS 스트리밍 수동 제어
- 제어판 상단에 **LIVE** 뱃지 (스트리밍 중일 때)
- **"방송 시작"** / **"방송 종료"** 버튼으로 수동 제어
- OBS가 연결되어 있지 않으면 에러 없이 무시됨

---

## 7. 설정

사이드바 → **설정** 클릭

| 항목 | 설명 |
|------|------|
| 성경 버전 | 성경 탭 기본 버전 선택 |
| 테마 | 라이트/다크 모드 |
| 폰트 크기 | 전체 UI 글꼴 크기 (12~24px) |
| 기본 BPM | 가사 탭에서 새 곡 추가 시 기본 BPM |
| 스트리밍 스케줄 | 예배 스케줄 + 카운트다운 + 자동 스트리밍 설정 |

---

## 8. Display 화면 & OBS 연동

| URL | 용도 | OBS 설정 |
|-----|------|----------|
| `localhost:8080/display` | 프로젝터용 전체화면 슬라이드 | Browser Source (1920x1080) |
| `localhost:8080/display/overlay` | 방송용 가사/텍스트 오버레이 | Browser Source (투명 배경) |

- **배경**: Figma에서 생성한 배경 이미지 + 항목별 커스텀 배경 (전주, 찬양, 참회의 기도)
- **키보드**: Display 창에서 ← → 로 직접 이동 가능
- **서버 재시작**: 마지막 순서/위치가 자동 복원됨

### OBS 씬 자동 전환 설정

슬라이드가 바뀔 때 OBS 씬을 자동으로 전환하려면 `config/obs.json`을 설정합니다.

```json
{
  "host": "localhost:4455",
  "password": "OBS_웹소켓_비밀번호",
  "scenes": {
    "찬송": "camera",
    "찬양": "camera",
    "대표기도": "camera",
    "말씀": "camera",
    "헌금봉헌": "camera",
    "봉헌기도": "camera",
    "축도": "camera",
    "교회소식": "monitor",
    "전주": "monitor",
    "성시교독": "monitor",
    "신앙고백": "monitor",
    "주기도문": "monitor",
    "성경봉독": "monitor",
    "예배의 부름": "monitor",
    "참회의 기도": "monitor"
  },
  "cameraScene": "camera",
  "displayScene": "monitor",
  "fadeMs": 800,
  "fadeDelaySec": 3
}
```

| 필드 | 설명 |
|------|------|
| `host` | OBS WebSocket 주소 (`host:port` 형식, 기본: `localhost:4455`) |
| `password` | OBS → 도구 → WebSocket 서버 설정에서 확인 |
| `scenes` | 예배 항목 title → OBS 씬 이름 매핑 (슬라이드 전환 시 자동 씬 변경) |
| `cameraScene` | 카메라 씬 이름 (자동 페이드백용) |
| `displayScene` | 화면 씬 이름 (자동 페이드백용) |
| `fadeMs` | 페이드 트랜지션 길이 (ms, 기본 800) |
| `fadeDelaySec` | displayScene 표시 후 camera 복귀까지 대기 시간 (초, 기본 3) |

> `scenes` 맵의 키(key)는 예배 순서 항목의 **title**과 정확히 일치해야 합니다.
> 매핑에 없는 항목은 씬 전환을 건너뜁니다.
> 설정 변경 후 서버를 **재시작**해야 반영됩니다.

### OBS Browser Source 자동 설정

OBS 소스 패널(제어판 내)의 **Display 탭**에서 자동으로 브라우저 소스를 추가할 수 있습니다.

1. 제어판 → OBS 소스 버튼 클릭
2. **Display** 탭 선택
3. 씬 선택 (예: `monitor`)
4. URL 확인 (`http://localhost:8080/display`)
5. **"Display 소스 설정"** 클릭 → `EP_Display` 소스가 자동 생성됨

---

## 9. 생성 이력

사이드바에서 각 유형별 이력 조회:
- 주보 생성 내역
- PPT 생성 내역
- 가사 PPT 생성 내역
- Display 생성 내역

각 이력에 성공/실패 상태, 생성 날짜가 표시됩니다.

---

## 10. 모바일 리모컨

예배 진행 중 스마트폰으로 슬라이드를 원격 제어할 수 있습니다.

### 접속 방법

**방법 1: QR 코드 스캔**
1. 웹 제어판에서 QR 아이콘 클릭
2. 스마트폰 카메라로 QR 코드 스캔
3. 자동으로 모바일 리모컨 페이지가 열립니다

**방법 2: IP 직접 입력**
1. 서버 컴퓨터의 WiFi IP 주소 확인 (예: `192.168.1.100`)
2. 스마트폰 브라우저에서 `http://192.168.1.100:8080/mobile` 입력

조건: 서버와 스마트폰이 **같은 WiFi 네트워크**에 연결되어 있어야 합니다.

### PWA 설치 (홈 화면 추가)

- iOS Safari: 공유 버튼 → 홈 화면에 추가
- Android Chrome: 메뉴(점 세 개) → 홈 화면에 추가

### 사용법

- **다음 / 이전** 버튼으로 슬라이드 이동
- 현재 항목 이름이 화면에 표시됩니다
- 실시간으로 프로젝터 화면과 동기화됩니다

---

## 11. 라이선스 관리

### 플랜 비교

| 기능 | Free | Pro |
|------|------|-----|
| 주보/가사 PDF 생성 | O | O |
| 성경 조회 | O | O |
| Display 프로젝터 | O | O |
| 모바일 리모컨 | O | O |
| OBS 제어 + 스트리밍 | X | O |
| 자동 스케줄러 | X | O |
| YouTube 연동 | X | O |
| 썸네일 자동 생성 | X | O |

### Pro 활성화 방법

1. 사이드바 → **라이선스 정보** 클릭
2. 라이선스 키 입력 (`EP-XXXX-XXXX-XXXX-XXXX` 형식)
3. **활성화** 클릭
4. 서버 재시작 없이 즉시 적용

### 라이선스 상태 확인

- 현재 플랜 (무료 / Pro / Enterprise)
- 만료일 (영구 라이선스는 표시 없음)
- 디바이스 ID (지원 문의 시 필요)

### 라이선스 비활성화

라이선스 패널에서 **비활성화** 클릭 → Free 플랜으로 복귀합니다.
다른 컴퓨터로 이전할 때 먼저 비활성화한 뒤 새 기기에서 활성화하세요.

### 오프라인 사용

- 라이선스 정보는 로컬 DB + 파일 캐시(`data/license.json`)에 이중 저장
- 인터넷 없이도 30일간 Pro 기능 유지 (grace period)
- 30일 이후에는 인터넷 연결 후 재검증 필요

---

## 12. 버전 업데이트

### 자동 알림

앱 실행 시 자동으로 최신 버전을 확인합니다.
새 버전이 있으면 화면 상단에 업데이트 알림 배너가 표시됩니다.

### 수동 확인

```
http://localhost:8080/api/version       ← 현재 버전 확인
http://localhost:8080/api/update/check  ← 최신 버전과 비교
```

### 업데이트 방법

1. 업데이트 배너 클릭 → GitHub Releases 페이지로 이동
2. 플랫폼에 맞는 바이너리 다운로드
   - macOS (Apple Silicon): `easyPreparation_darwin_arm64`
   - macOS (Intel): `easyPreparation_darwin_amd64`
   - Linux: `easyPreparation_linux_amd64`
   - Windows: `easyPreparation_windows_amd64.exe`
3. 기존 바이너리 교체 후 재시작

**주의**: `config/`, `data/` 디렉토리는 유지하세요. 설정과 캐시가 보존됩니다.

---

## 13. 개발 모드 실행

### 기본 개발 서버 (Go + Next.js)

```bash
make dev
```

- Go 서버: `http://localhost:8080`
- Next.js 개발 서버: `http://localhost:3000` (프록시)
- 브라우저에서 `http://localhost:3000` 접속

포트가 이미 사용 중이면 자동으로 정리 후 재시작합니다.

### 캐시 초기화 후 재시작

```bash
make restart
```

`.next` 캐시를 삭제하고 개발 서버를 재시작합니다. 스타일/라우팅 오류 발생 시 사용합니다.

### Desktop 앱 개발 모드

```bash
make dev-desktop
```

Wails 핫리로드 모드로 실행됩니다. 코드 변경 시 앱이 자동으로 갱신됩니다.

---

## 14. 프로덕션 빌드

### 서버 바이너리 빌드

```bash
make build
```

1. Next.js static export (`ui/out/`)
2. `cmd/server/frontend/`로 복사
3. Go binary 빌드 (frontend embed 포함)
4. 결과물: `bin/server`

실행:

```bash
./bin/server
```

접속: `http://localhost:8080`

### Go 서버만 빌드 (frontend 이미 있을 때)

```bash
make build-go
```

### Next.js만 빌드

```bash
make build-ui
```

---

## 15. Desktop 앱

### 빌드

```bash
make build-desktop
```

사전 조건:
- Wails CLI 설치: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- macOS: Xcode Command Line Tools 설치

결과물: `build/bin/easyPreparation.app`

### 설치 및 실행

1. `build/bin/easyPreparation.app`을 `/Applications`로 복사
2. 더블클릭으로 실행
3. 앱이 자동으로 내장 서버를 시작하고 WebView를 엽니다

### 첫 실행 설정

Desktop 앱은 실행 파일 위치를 기준으로 설정 파일을 탐색합니다.

```
easyPreparation.app/Contents/MacOS/ (실행 파일 위치)
├── config/
│   ├── db.json           ← PostgreSQL 연결 정보
│   ├── obs.json          ← OBS WebSocket 설정 (선택)
│   └── main_worship.json ← 예배 순서 (선택)
└── data/                 ← 자동 생성
```

`config/db.json` 예시:

```json
{
  "dsn": "host=localhost port=5432 user=postgres password=yourpassword dbname=easyprep sslmode=disable"
}
```

---

## 16. 릴리즈 방법 (개발자)

### 새 버전 릴리즈

```bash
# 1. 변경사항을 커밋하고 push
git add -p
git commit -m "feat: 새 기능 추가"
git push

# 2. 버전 태그 생성 및 push
git tag v1.2.0
git push origin v1.2.0
```

태그를 push하면 GitHub Actions가 자동으로:

1. 4개 플랫폼용 바이너리 빌드 (darwin arm64/amd64, linux amd64, windows amd64)
2. GitHub Release 생성
3. 커밋 메시지 기반 릴리즈 노트 자동 생성
4. 빌드된 바이너리를 Release에 첨부

### 버전 태그 형식

Semantic Versioning 사용: `v{메이저}.{마이너}.{패치}`

```
v1.0.0   ← 정식 릴리즈
v1.0.1   ← 버그 수정
v1.1.0   ← 새 기능
v2.0.0   ← 하위 호환성 깨지는 변경
```

### 빌드 확인 (로컬)

```bash
go build -tags dev -o /dev/null ./cmd/server/
```

---

## 17. Pro 기능 상세

| 기능 | 설명 |
|------|------|
| **OBS 연동** | 예배 항목 전환 시 OBS 씬 자동 변경 |
| **자동 스케줄러** | 예배 시간에 자동으로 순서 로드 + OBS 스트리밍 시작 |
| **YouTube 연동** | YouTube Live 방송 자동 생성 + 스트림 키 OBS 자동 설정 |
| **썸네일 생성** | YouTube 영상 썸네일 자동 생성 + 업로드 |
| **다중 예배** | 주일/오후/수요/금요 예배 구분 관리 |
| **클라우드 백업** | 예배 순서 데이터 클라우드 저장 (Enterprise) |

### 무료 플랜에서 사용 가능한 기능

- 주보 PDF 생성
- 찬양 PDF 생성
- Google Drive 파일 다운로드 (찬송/교독)
- 예배 화면 Display (`/display`)
- 가사 오버레이 (`/display/overlay`)
- 성경 조회 (다중 버전)
- 찬송가 검색
- 모바일 리모컨 (`/mobile`)
- 생성 이력 조회
