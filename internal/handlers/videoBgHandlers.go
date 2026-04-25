package handlers

import (
	"easyPreparation_1.0/internal/path"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func videoBgDir() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "video-bg")
}

func ensureVideoBgDir() error {
	return os.MkdirAll(videoBgDir(), 0755)
}

// VideoBgUploadHandler — POST /api/video-bg/upload
func VideoBgUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(300 << 20); err != nil { // 300MB
		http.Error(w, "파일이 너무 큽니다 (최대 300MB)", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "파일 읽기 실패", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mp4" && ext != ".webm" && ext != ".mov" {
		http.Error(w, "mp4, webm, mov 파일만 허용됩니다", http.StatusBadRequest)
		return
	}

	if err := ensureVideoBgDir(); err != nil {
		http.Error(w, "디렉토리 생성 실패", http.StatusInternalServerError)
		return
	}

	// 안전한 파일명 (기본 이름만 사용)
	safeName := filepath.Base(header.Filename)
	savePath := filepath.Join(videoBgDir(), safeName)

	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "파일 저장 실패", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "파일 쓰기 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"ok":       "true",
		"filename": safeName,
		"url":      "/display/video-bg/" + safeName,
	})
}

// VideoBgListHandler — GET /api/video-bg/list
func VideoBgListHandler(w http.ResponseWriter, r *http.Request) {
	_ = ensureVideoBgDir()
	entries, err := os.ReadDir(videoBgDir())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}
	var files []map[string]string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".mp4" && ext != ".webm" && ext != ".mov" {
			continue
		}
		files = append(files, map[string]string{
			"filename": e.Name(),
			"url":      "/display/video-bg/" + e.Name(),
		})
	}
	if files == nil {
		files = []map[string]string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// VideoBgDeleteHandler — DELETE /api/video-bg/delete?filename=xxx
func VideoBgDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	name := filepath.Base(r.URL.Query().Get("filename"))
	if name == "" || name == "." {
		http.Error(w, "파일명 필요", http.StatusBadRequest)
		return
	}
	target := filepath.Join(videoBgDir(), name)
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		http.Error(w, "삭제 실패", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// VideoBgServeHandler — GET /display/video-bg/{filename}
func VideoBgServeHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(strings.TrimPrefix(r.URL.Path, "/display/video-bg/"))
	if name == "" || name == "." {
		http.NotFound(w, r)
		return
	}
	filePath := filepath.Join(videoBgDir(), name)
	http.ServeFile(w, r, filePath)
}
