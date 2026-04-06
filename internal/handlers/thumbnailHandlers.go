package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/thumbnail"
	"easyPreparation_1.0/internal/youtube"
)

// ThumbnailGenerateHandler — POST /api/thumbnail/generate
// {worshipType: "main_worship", date?: "2026-04-05"}
func ThumbnailGenerateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		WorshipType string `json:"worshipType"`
		Date        string `json:"date"`
		Upload      bool   `json:"upload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	date := time.Now()
	if body.Date != "" {
		if d, err := time.Parse("2006-01-02", body.Date); err == nil {
			date = d
		}
	}

	outPath, err := generateThumbnail(body.WorshipType, date)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	// YouTube 업로드 (요청 시)
	if body.Upload {
		go func() {
			// 방송 제목 변경
			cfg, _ := thumbnail.LoadConfig()
			if cfg != nil {
				_, title := cfg.ResolveTheme(body.WorshipType, date)
				if err := youtube.UpdateBroadcastTitle(title); err != nil {
					log.Printf("[thumbnail] YouTube 제목 변경 실패: %v", err)
				}
			}
			// 썸네일 업로드
			if err := youtube.UploadThumbnail(outPath); err != nil {
				log.Printf("[thumbnail] YouTube 업로드 실패: %v", err)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":   true,
		"path": outPath,
	})
}

// ThumbnailPreviewHandler — GET /api/thumbnail/preview?worshipType=main_worship&date=2026-04-05
func ThumbnailPreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	worshipType := r.URL.Query().Get("worshipType")
	dateStr := r.URL.Query().Get("date")

	date := time.Now()
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = d
		}
	}
	if worshipType == "" {
		worshipType = "main_worship"
	}

	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "data", "templates", "thumbnail", "generated",
		fmt.Sprintf("%s_%s.png", date.Format("2006-01-02"), worshipType))

	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		// 자동 생성
		var genErr error
		imgPath, genErr = generateThumbnail(worshipType, date)
		if genErr != nil {
			http.Error(w, genErr.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, imgPath)
}

// ThumbnailConfigHandler — GET/POST /api/thumbnail/config
func ThumbnailConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		cfg, err := thumbnail.LoadConfig()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)

	case http.MethodPost:
		var cfg thumbnail.ThumbnailConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := thumbnail.SaveConfig(&cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// generateThumbnail — 내부 공통 생성 함수
func generateThumbnail(worshipType string, date time.Time) (string, error) {
	cfg, err := thumbnail.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("설정 로드 실패: %w", err)
	}

	bgPath, title := cfg.ResolveTheme(worshipType, date)

	execPath := path.ExecutePath("easyPreparation")
	outPath := filepath.Join(execPath, "data", "templates", "thumbnail", "generated",
		fmt.Sprintf("%s_%s.png", date.Format("2006-01-02"), worshipType))

	// 배경 경로가 상대 경로이면 절대 경로로 변환
	if bgPath != "" && !filepath.IsAbs(bgPath) {
		bgPath = filepath.Join(execPath, bgPath)
	}

	return thumbnail.Generate(thumbnail.GenerateConfig{
		BackgroundPath: bgPath,
		Title:          title,
		OutputPath:     outPath,
		Width:          1280,
		Height:         720,
	})
}

// ThumbnailUploadHandler — POST /api/thumbnail/upload (multipart)
// 배경 이미지를 data/templates/thumbnail/special/ 에 저장하고 상대 경로 반환
func ThumbnailUploadHandler(w http.ResponseWriter, r *http.Request) {
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

	// 저장 경로 결정
	execPath := path.ExecutePath("easyPreparation")
	specialDir := filepath.Join(execPath, "data", "templates", "thumbnail", "special")
	os.MkdirAll(specialDir, 0755)

	// 파일명 정리 (확장자 유지)
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".png"
	}
	// target 파라미터가 있으면 사용 (예: "default_main_worship")
	target := r.FormValue("target")
	var saveName string
	if target != "" {
		saveName = target + ext
	} else {
		saveName = header.Filename
	}

	// default_ 접두사면 data/templates/thumbnail/ 에 저장 (기본 배경 교체)
	var savePath, relPath string
	if len(target) > 8 && target[:8] == "default_" {
		typeName := target[8:] // "main_worship" 등
		savePath = filepath.Join(execPath, "data", "templates", "thumbnail", typeName+ext)
		relPath = "data/templates/thumbnail/" + typeName + ext
	} else {
		savePath = filepath.Join(specialDir, saveName)
		relPath = "data/templates/thumbnail/special/" + saveName
	}

	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "파일 저장 실패", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	buf := make([]byte, 1024*64)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
		}
		if readErr != nil {
			break
		}
	}

	log.Printf("[thumbnail] 배경 업로드: %s", savePath)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":   true,
		"path": relPath,
	})
}

// ThumbnailImageHandler — GET /api/thumbnail/image?path=data/templates/thumbnail/special/easter.png
// 배경 이미지 서빙 (미리보기용)
func ThumbnailImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		http.Error(w, "path 파라미터 없음", http.StatusBadRequest)
		return
	}

	execPath := path.ExecutePath("easyPreparation")
	absPath := filepath.Join(execPath, relPath)

	// 보안: data/templates/thumbnail/ 하위만 허용
	thumbDir := filepath.Join(execPath, "data", "templates", "thumbnail")
	if !isSubPath(thumbDir, absPath) {
		http.Error(w, "허용되지 않는 경로", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		http.Error(w, "파일 없음", http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, absPath)
}

func isSubPath(base, target string) bool {
	absBase, _ := filepath.Abs(base)
	absTarget, _ := filepath.Abs(target)
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false
	}
	return len(rel) >= 1 && rel[0] != '.'
}

// GenerateAndUploadThumbnail — 스케줄러에서 호출하는 공개 함수
// 썸네일 생성 + YouTube 업로드 + 방송 제목 변경
func GenerateAndUploadThumbnail(worshipType string) {
	date := time.Now()

	// 설정에서 제목 가져오기
	cfg, err := thumbnail.LoadConfig()
	if err != nil {
		log.Printf("[thumbnail] 설정 로드 실패: %v", err)
		return
	}
	_, title := cfg.ResolveTheme(worshipType, date)

	outPath, err := generateThumbnail(worshipType, date)
	if err != nil {
		log.Printf("[thumbnail] 생성 실패: %v", err)
		return
	}
	log.Printf("[thumbnail] 생성 완료: %s", outPath)

	// YouTube 방송 제목 변경
	if err := youtube.UpdateBroadcastTitle(title); err != nil {
		log.Printf("[thumbnail] YouTube 제목 변경 실패: %v", err)
	}

	// YouTube 썸네일 업로드
	if err := youtube.UploadThumbnail(outPath); err != nil {
		log.Printf("[thumbnail] YouTube 업로드 실패: %v", err)
	}
}

// GenerateAndUploadThumbnailTo — 특정 broadcastID에 썸네일 생성 + 업로드
// setup-obs에서 방송 생성 직후(upcoming 상태) 호출 → 확실히 반영됨
func GenerateAndUploadThumbnailTo(worshipType, broadcastID string) {
	date := time.Now()

	cfg, err := thumbnail.LoadConfig()
	if err != nil {
		log.Printf("[thumbnail] 설정 로드 실패: %v", err)
		return
	}
	_, title := cfg.ResolveTheme(worshipType, date)

	outPath, err := generateThumbnail(worshipType, date)
	if err != nil {
		log.Printf("[thumbnail] 생성 실패: %v", err)
		return
	}
	log.Printf("[thumbnail] 생성 완료: %s", outPath)

	// 방송 제목 변경
	if err := youtube.UpdateBroadcastTitle(title); err != nil {
		log.Printf("[thumbnail] YouTube 제목 변경 실패: %v", err)
	}

	// 특정 broadcastID에 썸네일 업로드 (upcoming 상태에서 호출 → 확실히 반영)
	if err := youtube.UploadThumbnailToBroadcast(broadcastID, outPath); err != nil {
		log.Printf("[thumbnail] YouTube 썸네일 업로드 실패: %v", err)
	}
}
