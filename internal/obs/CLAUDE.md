# internal/obs — OBS WebSocket 매니저

싱글턴 매니저. `config/obs.json` 없으면 전체 no-op.

## 동작

- goobs 라이브러리로 OBS WebSocket 연결
- 자동 재연결 (5초 간격)
- 씬 매핑: `config/obs.json`의 `scenes` 맵 → 항목 title → OBS 씬 전환
- 매핑 없는 항목은 씬 전환 스킵

## 주요 메서드

- `Init(configPath)` — 설정 로드 + 연결
- `Get()` — 싱글턴 인스턴스
- `resolveScene(title)` → `(sceneName, ok)`
- `StartStreaming()` / `StopStreaming()` / `GetStreamStatus()`
