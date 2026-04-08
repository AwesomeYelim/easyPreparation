package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"easyPreparation_1.0/internal/path"
)

// 카테고리 → 디렉토리 매핑
func templateDir(category string) string {
	execPath := path.ExecutePath("easyPreparation")
	switch category {
	case "display":
		return filepath.Join(execPath, "data", "templates", "display")
	case "display-default", "lyrics":
		return filepath.Join(execPath, "data", "templates", "lyrics")
	default:
		return ""
	}
}

// 허용 확장자
func isImageExt(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

// TemplateListHandler — GET /api/templates?category=display
func TemplateListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	category := r.URL.Query().Get("category")
	dir := templateDir(category)
	if dir == "" {
		http.Error(w, "잘못된 카테고리", http.StatusBadRequest)
		return
	}

	type FileInfo struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Size int64  `json:"size"`
	}

	var files []FileInfo

	switch category {
	case "display":
		entries, err := os.ReadDir(dir)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"files": []FileInfo{}})
			return
		}
		for _, e := range entries {
			if e.IsDir() || !isImageExt(filepath.Ext(e.Name())) {
				continue
			}
			info, _ := e.Info()
			var size int64
			if info != nil {
				size = info.Size()
			}
			files = append(files, FileInfo{
				Name: e.Name(),
				URL:  "/api/templates/display/" + e.Name(),
				Size: size,
			})
		}

	case "display-default":
		fixedName := "Frame 2.png"
		fPath := filepath.Join(dir, fixedName)
		if info, err := os.Stat(fPath); err == nil {
			files = append(files, FileInfo{
				Name: fixedName,
				URL:  "/api/templates/display-default/" + fixedName,
				Size: info.Size(),
			})
		}

	case "lyrics":
		fixedName := "Frame 1.png"
		fPath := filepath.Join(dir, fixedName)
		if info, err := os.Stat(fPath); err == nil {
			files = append(files, FileInfo{
				Name: fixedName,
				URL:  "/api/templates/lyrics/" + fixedName,
				Size: info.Size(),
			})
		}
	}

	if files == nil {
		files = []FileInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"files": files})
}

// TemplateUploadHandler — POST /api/templates/upload (multipart: image, category, name)
func TemplateUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 10MB 제한
	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "이미지 파일 없음", http.StatusBadRequest)
		return
	}
	defer file.Close()

	category := r.FormValue("category")
	name := r.FormValue("name") // display 카테고리에서만 사용 (항목명)

	dir := templateDir(category)
	if dir == "" {
		http.Error(w, "잘못된 카테고리", http.StatusBadRequest)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isImageExt(ext) {
		http.Error(w, "PNG/JPG만 허용", http.StatusBadRequest)
		return
	}

	os.MkdirAll(dir, 0755)

	// 저장 파일명 결정
	var saveName string
	switch category {
	case "display":
		if name == "" {
			// name 없으면 원본 파일명 사용
			saveName = filepath.Base(header.Filename)
		} else {
			saveName = filepath.Base(name) + ext
		}
	case "display-default":
		saveName = "Frame 2.png"
		// display-default는 항상 PNG로 저장 (확장자 강제)
		ext = ".png"
	case "lyrics":
		saveName = "Frame 1.png"
		ext = ".png"
	}

	// 경로 순회 방지
	saveName = filepath.Base(saveName)
	if strings.Contains(saveName, "..") || saveName == "." || saveName == "" {
		http.Error(w, "잘못된 파일명", http.StatusBadRequest)
		return
	}

	savePath := filepath.Clean(filepath.Join(dir, saveName))
	cleanDir := filepath.Clean(dir)
	if !strings.HasPrefix(savePath, cleanDir+string(filepath.Separator)) {
		http.Error(w, "허용되지 않는 경로", http.StatusForbidden)
		return
	}

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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":   true,
		"name": saveName,
		"url":  "/api/templates/" + category + "/" + saveName,
	})
}

// TemplateDeleteHandler — DELETE /api/templates/{category}/{filename}
func TemplateDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL: /api/templates/{category}/{filename}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, "잘못된 경로", http.StatusBadRequest)
		return
	}

	category := parts[0]
	filename := filepath.Base(parts[1])

	dir := templateDir(category)
	if dir == "" {
		http.Error(w, "잘못된 카테고리", http.StatusBadRequest)
		return
	}

	// display-default/lyrics 고정 파일은 삭제 불가
	if category == "display-default" || category == "lyrics" {
		http.Error(w, "고정 파일은 삭제할 수 없습니다 (교체만 가능)", http.StatusForbidden)
		return
	}

	filePath := filepath.Clean(filepath.Join(dir, filename))
	cleanDir := filepath.Clean(dir)
	if !strings.HasPrefix(filePath, cleanDir+string(filepath.Separator)) {
		http.Error(w, "허용되지 않는 경로", http.StatusForbidden)
		return
	}

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "파일 없음", http.StatusNotFound)
		} else {
			http.Error(w, "삭제 실패", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// TemplateServeHandler — GET /api/templates/{category}/{filename}
func TemplateServeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, "잘못된 경로", http.StatusBadRequest)
		return
	}

	category := parts[0]
	filename := filepath.Base(parts[1])

	dir := templateDir(category)
	if dir == "" {
		http.Error(w, "잘못된 카테고리", http.StatusBadRequest)
		return
	}

	filePath := filepath.Clean(filepath.Join(dir, filename))
	cleanDir := filepath.Clean(dir)
	if !strings.HasPrefix(filePath, cleanDir+string(filepath.Separator)) {
		http.Error(w, "허용되지 않는 경로", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "파일 없음", http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, filePath)
}

// TemplateHandler — /api/templates 라우트 분기 핸들러
// GET /api/templates?category=... → List
// POST /api/templates/upload → Upload
// GET/DELETE /api/templates/{category}/{filename} → Serve/Delete
func TemplateHandler(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/api/templates")

	// GET /api/templates (or /api/templates?category=...)
	if p == "" || p == "/" {
		TemplateListHandler(w, r)
		return
	}

	// POST /api/templates/upload
	if p == "/upload" {
		TemplateUploadHandler(w, r)
		return
	}

	// GET/DELETE /api/templates/{category}/{filename}
	switch r.Method {
	case http.MethodGet, http.MethodOptions:
		TemplateServeHandler(w, r)
	case http.MethodDelete:
		TemplateDeleteHandler(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
