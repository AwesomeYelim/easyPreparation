# easyPreparation 사용자 가이드

## 목차

1. [개발 모드 실행](#1-개발-모드-실행)
2. [프로덕션 빌드](#2-프로덕션-빌드)
3. [Desktop 앱 사용법](#3-desktop-앱-사용법)
4. [모바일 리모컨](#4-모바일-리모컨)
5. [라이선스 관리](#5-라이선스-관리)
6. [버전 업데이트 확인](#6-버전-업데이트-확인)
7. [릴리즈 방법 (개발자)](#7-릴리즈-방법-개발자)
8. [Pro 기능 상세](#8-pro-기능-상세)

---

## 1. 개발 모드 실행

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

## 2. 프로덕션 빌드

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

## 3. Desktop 앱 사용법

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
│   ├── db.json          ← PostgreSQL 연결 정보
│   ├── obs.json         ← OBS WebSocket 설정 (선택)
│   └── main_worship.json ← 예배 순서 (선택)
└── data/                ← 자동 생성
```

`config/db.json` 예시:

```json
{
  "dsn": "host=localhost port=5432 user=postgres password=yourpassword dbname=easyprep sslmode=disable"
}
```

### OBS 연동 설정

`config/obs.json` 예시:

```json
{
  "host": "localhost",
  "port": 4455,
  "password": "your-obs-password",
  "scenes": {
    "전주": "전주 씬",
    "찬양": "찬양 씬",
    "말씀": "말씀 씬"
  }
}
```

OBS → 도구 → WebSocket 서버 설정에서 포트/비밀번호를 확인하세요.

---

## 4. 모바일 리모컨

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

모바일 브라우저에서 "홈 화면에 추가"를 선택하면 앱처럼 사용할 수 있습니다.

- iOS Safari: 공유 버튼 → 홈 화면에 추가
- Android Chrome: 메뉴(점 세 개) → 홈 화면에 추가

### 사용법

- **다음 / 이전** 버튼으로 슬라이드 이동
- 현재 항목 이름이 화면에 표시됩니다
- 실시간으로 프로젝터 화면과 동기화됩니다

---

## 5. 라이선스 관리

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

라이선스 패널에서 다음 정보를 확인할 수 있습니다:

- 현재 플랜 (무료 / Pro / Enterprise)
- 만료일 (영구 라이선스는 표시 없음)
- 디바이스 ID (지원 문의 시 필요)

### 라이선스 비활성화

라이선스 패널에서 **비활성화** 클릭 → Free 플랜으로 복귀합니다.
다른 컴퓨터로 이전할 때도 동일하게 먼저 비활성화한 뒤 새 기기에서 활성화하세요.

### 오프라인 사용

- 라이선스 정보는 로컬 DB + 파일 캐시(`data/license.json`)에 이중 저장
- 인터넷 없이도 30일간 Pro 기능 유지 (grace period)
- 30일 이후에는 인터넷 연결 후 재검증 필요

---

## 6. 버전 업데이트 확인

### 자동 알림

앱 실행 시 자동으로 최신 버전을 확인합니다.
새 버전이 있으면 화면 상단에 업데이트 알림 배너가 표시됩니다.

### 수동 확인

브라우저 주소창에 입력:

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

## 7. 릴리즈 방법 (개발자)

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
# 빌드 성공 여부만 확인 (바이너리 생성 없음)
go build -tags dev -o /dev/null ./cmd/server/
```

---

## 8. Pro 기능 상세

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
