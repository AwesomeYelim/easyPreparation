package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"easyPreparation_1.0/internal/httpx"
	"easyPreparation_1.0/internal/path"
)

// AssetServeHandler — data/cache/pdf/ 디렉토리에서 PDF 파일을 서빙합니다.
// 경로: /api/assets/{category}/{filename}
// 예: /api/assets/hymn/032.pdf → data/cache/pdf/hymn/032.pdf
func AssetServeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// /api/assets/hymn/032.pdf → ["hymn", "032.pdf"]
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/assets/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		httpx.Error(w, http.StatusNotFound, "Not Found")
		return
	}

	category := parts[0]
	filename := parts[1]

	// 허용 카테고리 제한
	if category != "hymn" && category != "responsive_reading" {
		httpx.Error(w, http.StatusNotFound, "Not Found")
		return
	}

	// filepath.Base로 디렉토리 순회 제거 (../etc 방지)
	category = filepath.Base(category)
	filename = filepath.Base(filename)

	execPath := path.ExecutePath("easyPreparation")
	baseDir := filepath.Join(execPath, "data", "cache", "pdf", category)
	filePath := filepath.Clean(filepath.Join(baseDir, filename))

	// 해석된 경로가 허용 디렉토리 내에 있는지 검증
	if !strings.HasPrefix(filePath, baseDir+string(filepath.Separator)) && filePath != baseDir {
		httpx.Error(w, http.StatusBadRequest, "Bad Request")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		httpx.Error(w, http.StatusNotFound, "Not Found")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, filePath)
}
