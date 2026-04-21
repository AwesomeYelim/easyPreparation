package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"easyPreparation_1.0/internal/path"
)

// ── PDF 슬라이드 상태 (메모리) ──
var (
	pdfMu           sync.RWMutex
	pdfSlideCount   int
	pdfCurrentIndex int
)

// pdfSlidesDir — 슬라이드 PNG 저장 디렉터리
func pdfSlidesDir() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "pdf-slides")
}

// findGhostscript — Windows: gswin64c / gswin32c / gs 순, 그외: gs
func findGhostscript() (string, error) {
	candidates := []string{"gs"}
	if runtime.GOOS == "windows" {
		candidates = []string{"gswin64c", "gswin32c", "gs"}
	}
	for _, name := range candidates {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("Ghostscript를 찾을 수 없습니다 (gswin64c/gswin32c/gs). PATH에 추가하거나 설치하세요")
}

// PDFUploadHandler — POST /api/pdf/upload (multipart, 최대 50 MB)
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

	const maxSize = 50 << 20 // 50 MB
	if err := r.ParseMultipartForm(maxSize); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "파일 크기가 너무 크거나 파싱 실패: " + err.Error()})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "파일 없음"})
		return
	}
	defer file.Close()

	// 임시 파일에 저장
	tmpFile, err := os.CreateTemp("", "ep-pdf-*.pdf")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "임시 파일 생성 실패"})
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "파일 저장 실패"})
		return
	}
	tmpFile.Close()

	// 슬라이드 디렉터리 준비 (기존 파일 삭제 후 재생성)
	slideDir := pdfSlidesDir()
	_ = os.RemoveAll(slideDir)
	if err := os.MkdirAll(slideDir, 0755); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "디렉터리 생성 실패"})
		return
	}

	// Ghostscript 경로 탐색
	gsPath, err := findGhostscript()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	// 출력 패턴: 001.png, 002.png …
	outPattern := filepath.Join(slideDir, "%03d.png")

	cmd := exec.Command(gsPath,
		"-dNOPAUSE", "-dBATCH", "-dSAFER",
		"-sDEVICE=png16m",
		"-r150",
		"-sOutputFile="+outPattern,
		tmpFile.Name(),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[pdf] gs 변환 실패: %v\n%s", err, string(out))
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "PDF 변환 실패: " + err.Error()})
		return
	}

	// 생성된 PNG 파일 개수
	entries, _ := os.ReadDir(slideDir)
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".png") {
			count++
		}
	}

	pdfMu.Lock()
	pdfSlideCount = count
	pdfCurrentIndex = 0
	pdfMu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": count})
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
		pdfMu.RUnlock()
		json.NewEncoder(w).Encode(map[string]interface{}{"count": count, "currentIndex": idx})

	case http.MethodDelete:
		slideDir := pdfSlidesDir()
		if err := os.RemoveAll(slideDir); err != nil {
			log.Printf("[pdf] 슬라이드 삭제 실패: %v", err)
		}
		_ = os.MkdirAll(slideDir, 0755)

		pdfMu.Lock()
		pdfSlideCount = 0
		pdfCurrentIndex = 0
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

// PDFDisplayHandler — GET /display/pdf (OBS Browser Source용 HTML)
const pdfDisplayHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<title>PDF Display</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; }
  html, body {
    width:1920px; height:1080px;
    background:#000;
    overflow:hidden;
    display:flex; align-items:center; justify-content:center;
  }
  #slide-img {
    max-width:1920px; max-height:1080px;
    width:100%; height:100%;
    object-fit:contain;
    display:none;
  }
  #no-slide {
    color:rgba(255,255,255,0.4);
    font-family:sans-serif;
    font-size:32px;
  }
</style>
</head>
<body>
  <img id="slide-img" src="" alt="slide">
  <div id="no-slide">PDF 없음</div>
<script>
  var cur = -1;
  function poll() {
    fetch('/api/pdf/slides')
      .then(function(r){ return r.json(); })
      .then(function(d){
        if (!d.count) {
          document.getElementById('slide-img').style.display='none';
          document.getElementById('no-slide').style.display='block';
          cur = -1;
          return;
        }
        if (d.currentIndex !== cur) {
          cur = d.currentIndex;
          var n = String(cur+1).padStart(3,'0');
          var img = document.getElementById('slide-img');
          img.src = '/display/pdf-slides/'+n+'.png?t='+Date.now();
          img.style.display='block';
          document.getElementById('no-slide').style.display='none';
        }
      })
      .catch(function(){})
      .finally(function(){ setTimeout(poll,1000); });
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

// PDFSlidesServeHandler — GET /display/pdf-slides/{NNN}.png
// data/pdf-slides/ 에서 파일 서빙 (경로 탈출 방지)
func PDFSlidesServeHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Path)
	base := strings.TrimSuffix(name, ".png")
	if _, err := strconv.Atoi(base); err != nil || !strings.HasSuffix(name, ".png") {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, filepath.Join(pdfSlidesDir(), name))
}
