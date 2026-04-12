# internal/license — 라이선스 + Feature Gating

## 파일

| 파일 | 역할 |
|------|------|
| `types.go` | Plan/Feature 상수, LicenseInfo 구조체, PlanFeatures 맵 |
| `manager.go` | 싱글턴 Manager (DB + 파일 캐시 이중 저장) |
| `offline.go` | `data/license.json` 캐시, 디바이스 ID (MAC SHA256) |
| `keygen.go` | HMAC-SHA256 서명 검증 |
| `config.go` | `config/license.json` 서버 설정 로드 |
| `verifier.go` | 24시간 백그라운드 서버 검증 |

## 플랜별 기능

| Feature | Free | Pro |
|---------|:----:|:---:|
| OBS 연동 | | O |
| 자동 스케줄러 | | O |
| YouTube | | O |
| 썸네일 | | O |
| 다중 예배 | | O |

## 미들웨어

```go
middleware.FeatureGate(license.FeatureAutoScheduler, handler)
// 403 {"error":"feature_locked"} 반환
```

## 오프라인

`config/license.json` 없으면 서버 검증 스킵. grace period 30일.
