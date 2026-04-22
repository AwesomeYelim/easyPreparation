package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"easyPreparation_1.0/internal/path"
)

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
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// PDFNavigateHandler — POST /api/pdf/navigate {action: "prev"|"next"|"goto", index: N}
func PDFNavigateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Action string `json:"action"`
		Index  int    `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
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
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, fp)
}

// PDFDisplayHandler — GET /display/pdf (OBS Browser Source용 HTML)
// PDF.js로 PDF를 직접 렌더링 — 외부 툴 없음
const pdfDisplayHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<title>PDF Display</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; }
  html, body { width:1920px; height:1080px; background:#000; overflow:hidden; position:relative; }
  #c { display:none; position:absolute; top:0; left:0; width:1920px; height:1080px; }
  #msg { color:rgba(255,255,255,0.4); font:32px sans-serif;
         position:absolute; top:50%; left:50%; transform:translate(-50%,-50%); white-space:nowrap; }
</style>
</head>
<body>
<canvas id="c" width="1920" height="1080"></canvas>
<div id="msg">PDF 없음</div>
<script src="https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js"></script>
<script>
pdfjsLib.GlobalWorkerOptions.workerSrc =
  'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js';

var pdfDoc = null, loadedTs = 0, curPage = -1, rendering = false;

function render(idx) {
  if (!pdfDoc || rendering) return;
  rendering = true;
  pdfDoc.getPage(idx + 1).then(function(page) {
    var vp0 = page.getViewport({scale: 1});
    var scale = Math.min(1920 / vp0.width, 1080 / vp0.height);
    var vp = page.getViewport({scale: scale});
    var canvas = document.getElementById('c');
    var ctx = canvas.getContext('2d');
    ctx.fillStyle = '#000';
    ctx.fillRect(0, 0, 1920, 1080);
    var tx = (1920 - vp.width) / 2, ty = (1080 - vp.height) / 2;
    page.render({canvasContext: ctx, viewport: vp, transform: [1, 0, 0, 1, tx, ty]}).promise
      .then(function() { rendering = false; })
      .catch(function() { rendering = false; });
  }).catch(function() { rendering = false; });
}

function load(ts, idx) {
  loadedTs = ts;
  curPage = idx;
  pdfDoc = null;
  pdfjsLib.getDocument('/display/pdf-file').promise.then(function(pdf) {
    pdfDoc = pdf;
    fetch('/api/pdf/count', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({count: pdf.numPages})
    }).catch(function() {});
    document.getElementById('c').style.display = 'block';
    document.getElementById('msg').style.display = 'none';
    render(idx);
  }).catch(function() {});
}

function poll() {
  fetch('/api/pdf/slides')
    .then(function(r) { return r.json(); })
    .then(function(d) {
      if (!d.uploadTs) {
        document.getElementById('c').style.display = 'none';
        document.getElementById('msg').style.display = 'block';
        pdfDoc = null; loadedTs = 0; curPage = -1;
        return;
      }
      if (d.uploadTs !== loadedTs) { load(d.uploadTs, d.currentIndex); return; }
      if (d.currentIndex !== curPage && pdfDoc) { curPage = d.currentIndex; render(curPage); }
    })
    .catch(function() {})
    .finally(function() { setTimeout(poll, 1000); });
}
poll();
</script>
</body>
</html>`

func PDFDisplayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(pdfDisplayHTML))
}
