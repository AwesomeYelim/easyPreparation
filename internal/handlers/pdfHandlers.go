package handlers

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"easyPreparation_1.0/internal/httpx"
	"easyPreparation_1.0/internal/path"
)

//go:embed html/pdf-display.html
var pdfDisplayHTML string

// ── PDF 슬라이드 상태 (메모리) ──
var (
	pdfMu           sync.RWMutex
	pdfSlideCount   int
	pdfCurrentIndex int
	pdfUploadTs     int64 // 업로드 시각 (ms) — 브라우저 변경 감지용
)

func pdfDir() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "pdf-slides")
}

func pdfFilePath() string {
	return filepath.Join(pdfDir(), "current.pdf")
}

// PDFUploadHandler — POST /api/pdf/upload (multipart, 최대 50 MB)
// PDF 파일을 저장만 함. 변환 없음 — PDF.js가 브라우저에서 직접 렌더링.
func PDFUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	w.Header().Set("Content-Type", "application/json")

	const maxSize = 50 << 20
	if err := r.ParseMultipartForm(maxSize); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "파일 크기 초과 또는 파싱 실패"})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "파일 없음"})
		return
	}
	defer file.Close()

	if err := os.MkdirAll(pdfDir(), 0755); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "디렉터리 생성 실패"})
		return
	}

	dst, err := os.Create(pdfFilePath())
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "파일 저장 실패"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "파일 쓰기 실패"})
		return
	}

	pdfMu.Lock()
	pdfSlideCount = 0 // 브라우저가 PDF 로드 후 /api/pdf/count로 보고
	pdfCurrentIndex = 0
	pdfUploadTs = time.Now().UnixMilli()
	pdfMu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// PDFCountHandler — POST /api/pdf/count {count: N}
// PDF.js 브라우저가 PDF 로드 완료 후 페이지 수를 서버에 보고
func PDFCountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var body struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Count <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pdfMu.Lock()
	pdfSlideCount = body.Count
	pdfMu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// PDFSlidesHandler — GET /api/pdf/slides (상태 조회) | DELETE /api/pdf/slides (초기화)
func PDFSlidesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		pdfMu.RLock()
		count := pdfSlideCount
		idx := pdfCurrentIndex
		ts := pdfUploadTs
		pdfMu.RUnlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":        count,
			"currentIndex": idx,
			"uploadTs":     ts,
		})

	case http.MethodDelete:
		_ = os.Remove(pdfFilePath())
		pdfMu.Lock()
		pdfSlideCount = 0
		pdfCurrentIndex = 0
		pdfUploadTs = 0
		pdfMu.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})

	default:
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// PDFNavigateHandler — POST /api/pdf/navigate {action: "prev"|"next"|"goto", index: N}
func PDFNavigateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Action string `json:"action"`
		Index  int    `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "잘못된 요청")
		return
	}

	pdfMu.Lock()
	switch body.Action {
	case "next":
		if pdfCurrentIndex < pdfSlideCount-1 {
			pdfCurrentIndex++
		}
	case "prev":
		if pdfCurrentIndex > 0 {
			pdfCurrentIndex--
		}
	case "goto":
		if body.Index >= 0 && body.Index < pdfSlideCount {
			pdfCurrentIndex = body.Index
		}
	}
	idx := pdfCurrentIndex
	count := pdfSlideCount
	pdfMu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "currentIndex": idx, "count": count})
}

// PDFFileHandler — GET /display/pdf-file
// 업로드된 PDF 원본 파일 서빙 (PDF.js가 이 URL로 PDF 로드)
func PDFFileHandler(w http.ResponseWriter, r *http.Request) {
	fp := pdfFilePath()
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		httpx.Error(w, http.StatusNotFound, "Not Found")
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, fp)
}

// PDFDisplayHandler — GET /display/pdf (OBS Browser Source용 HTML)
// PDF.js로 PDF를 직접 렌더링 — 외부 툴 없음
func PDFDisplayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(pdfDisplayHTML))
}
