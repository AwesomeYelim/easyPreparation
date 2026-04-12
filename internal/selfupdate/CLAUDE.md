# internal/selfupdate — 자동 업데이트

GitHub Releases API 기반. 저장소: `AwesomeYelim/easyPreparation`.

## 파일

| 파일 | 역할 |
|------|------|
| `checker.go` | `CheckLatest()` — 최신 버전 확인 + Asset 매칭 |
| `updater.go` | 다운로드 + WS 진행률 브로드캐스트 |
| `updater_unix.go` | macOS/Linux 바이너리 교체 (rename) |
| `updater_windows.go` | Windows batch script 교체 |
| `signature.go` | SHA256 checksums.txt 검증 |

## API

| 경로 | 설명 |
|------|------|
| `GET /api/update/check` | 최신 버전 비교 |
| `POST /api/update/download` | 바이너리 다운로드 |
| `POST /api/update/apply` | 교체 + 재시작 |

dev 빌드는 `IsNewer()` 항상 true.
