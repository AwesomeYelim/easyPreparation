package handlers

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

//go:embed html/display-preview.html
var displayPreviewHTML string

// DisplayPreviewHandler — GET /display/preview?index=N
// 씬 패널 미리보기: WebSocket 없이 단일 항목을 Display와 동일한 스타일로 렌더링
func DisplayPreviewHandler(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	index, _ := strconv.Atoi(indexStr)

	orderMu.RLock()
	order := deepCopyOrder(currentOrder)
	orderMu.RUnlock()

	if order == nil {
		order = []map[string]interface{}{}
	}
	if index < 0 || index >= len(order) {
		index = 0
	}

	allJSON, _ := json.Marshal(order)

	subPageStr := r.URL.Query().Get("subPage")
	subPage, _ := strconv.Atoi(subPageStr)

	html := strings.Replace(displayPreviewHTML, "/*__SLIDES__*/[]", "/*__SLIDES__*/"+string(allJSON), 1)
	html = strings.Replace(html, "/*__IDX__*/0", "/*__IDX__*/"+strconv.Itoa(index), 1)
	html = strings.Replace(html, "/*__SUBPAGE__*/0", "/*__SUBPAGE__*/"+strconv.Itoa(subPage), 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(html))
}
