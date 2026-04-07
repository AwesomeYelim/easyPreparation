package handlers

import (
	"easyPreparation_1.0/internal/path"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// AssetServeHandler — data/pdf/ 디렉토리에서 PDF 파일을 서빙합니다.
// 경로: /api/assets/{category}/{filename}
// 예: /api/assets/hymn/032.pdf → data/pdf/hymn/032.pdf
func AssetServeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// /api/assets/hymn/032.pdf → ["hymn", "032.pdf"]
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/assets/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	category := parts[0]
	filename := parts[1]

	// 경로 순회 공격 방지
	if strings.Contains(category, "..") || strings.Contains(filename, "..") ||
		strings.Contains(category, "/") || strings.Contains(filename, "/") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 허용 카테고리 제한
	if category != "hymn" && category != "responsive_reading" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	execPath := path.ExecutePath("easyPreparation")
	filePath := filepath.Join(execPath, "data", "pdf", category, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, filePath)
}
