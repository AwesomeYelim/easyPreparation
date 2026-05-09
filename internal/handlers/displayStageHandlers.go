package handlers

import (
	_ "embed"
	"net/http"
)

//go:embed html/display-stage.html
var displayStageHTML string

// DisplayStageHandler — /display/stage
// 찬양팀·설교자용 무대 모니터 화면.
// 현재 슬라이드(크게) + 다음 항목 미리보기 + 경과 타이머
// WebSocket 'order' / 'navigate' / 'position' 메시지 재사용 (추가 서버 변경 없음)
func DisplayStageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(displayStageHTML))
}
