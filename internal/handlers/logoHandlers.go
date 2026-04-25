package handlers

import (
	"easyPreparation_1.0/internal/path"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetLogoPath — 저장된 로고 파일 경로 반환 (없으면 "") — PDF 생성 등 외부에서 사용
func GetLogoPath() string {
	return findLogoPath()
}

// findLogoPath — 저장된 로고 파일 경로 반환 (없으면 "")
func findLogoPath() string {
	execPath := path.ExecutePath("easyPreparation")
	dataDir := filepath.Join(execPath, "data")
	for _, ext := range []string{".png", ".jpg", ".jpeg"} {
		p := filepath.Join(dataDir, "logo"+ext)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// HandleLogoGet — GET /api/logo
// 로고 이미지 서빙 (없으면 404)
func HandleLogoGet(w http.ResponseWriter, r *http.Request) {
	logoPath := findLogoPath()
	if logoPath == "" {
		http.NotFound(w, r)
		return
	}
	ext := strings.ToLower(filepath.Ext(logoPath))
	switch ext {
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	default:
		w.Header().Set("Content-Type", "image/png")
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, logoPath)
}

// HandleLogoUpload — POST /api/logo (multipart form: "logo" 파일)
// 기존 로고 삭제 후 새 로고 저장 (PNG/JPG/JPEG 허용, 최대 5MB)
func HandleLogoUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "파일이 너무 큽니다 (최대 5MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, "파일 읽기 실패", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		http.Error(w, "PNG/JPG 파일만 업로드 가능합니다", http.StatusBadRequest)
		return
	}

	execPath := path.ExecutePath("easyPreparation")
	dataDir := filepath.Join(execPath, "data")

	// 기존 로고 파일 삭제
	for _, oldExt := range []string{".png", ".jpg", ".jpeg"} {
		os.Remove(filepath.Join(dataDir, "logo"+oldExt))
	}

	destPath := filepath.Join(dataDir, "logo"+ext)
	out, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "파일 저장 실패", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		http.Error(w, "파일 쓰기 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

// HandleLogoDelete — DELETE /api/logo
// 로고 파일 삭제
func HandleLogoDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	execPath := path.ExecutePath("easyPreparation")
	dataDir := filepath.Join(execPath, "data")
	for _, ext := range []string{".png", ".jpg", ".jpeg"} {
		os.Remove(filepath.Join(dataDir, "logo"+ext))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
