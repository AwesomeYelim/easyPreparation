package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/thumbnail"
	"easyPreparation_1.0/internal/youtube"
)

// loadSermonDataFromOrder — 현재 display 메모리(currentOrder)에서 말씀 제목과 성경봉독 추출
// 성경봉독: title=="성경봉독" 우선, 없으면 b_edit 항목 fallback
func loadSermonDataFromOrder() (sermonTitle, scripture string) {
	orderMu.RLock()
	order := deepCopyOrder(currentOrder)
	orderMu.RUnlock()

	var bEditFallback string
	for _, item := range order {
		title, _ := item["title"].(string)
		obj, _ := item["obj"].(string)
		info, _ := item["info"].(string)
		if (title == "말씀" || title == "설교") && obj != "" && obj != "-" {
			sermonTitle = obj
		}
		if title == "성경봉독" && obj != "" && obj != "-" {
			scripture = obj
		}
		if strings.HasPrefix(info, "b_") && obj != "" && obj != "-" && bEditFallback == "" {
			bEditFallback = obj
		}
	}
	if scripture == "" {
		scripture = bEditFallback
	}
	return
}

// loadSermonDataFromConfig — config/{worshipType}.json에서 말씀 제목과 성경봉독 참조 추출
// 성경봉독: title=="성경봉독" 우선, 없으면 b_edit 항목 fallback
func loadSermonDataFromConfig(configPath string) (sermonTitle, scripture string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(data, &items); err != nil {
		return
	}
	var bEditFallback string
	for _, item := range items {
		title, _ := item["title"].(string)
		obj, _ := item["obj"].(string)
		info, _ := item["info"].(string)
		if (title == "말씀" || title == "설교") && obj != "" && obj != "-" {
			sermonTitle = obj
		}
		if title == "성경봉독" && obj != "" && obj != "-" {
			scripture = obj
		}
		if strings.HasPrefix(info, "b_") && obj != "" && obj != "-" && bEditFallback == "" {
			bEditFallback = obj
		}
	}
	if scripture == "" {
		scripture = bEditFallback
	}
	return
}

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

	// worshipType에서 경로 순회 방지
	worshipType = filepath.Base(worshipType)
	if strings.Contains(worshipType, "..") || worshipType == "." {
		http.Error(w, "잘못된 worshipType", http.StatusBadRequest)
		return
	}

	// 항상 재생성 — 예배 순서·로고·배경·폰트 변경 사항을 즉시 반영
	imgPath, err := generateThumbnail(worshipType, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	// worshipType 경로 순회 방지
	worshipType = filepath.Base(worshipType)
	if strings.Contains(worshipType, "..") || worshipType == "." || worshipType == "" {
		return "", fmt.Errorf("잘못된 worshipType: %s", worshipType)
	}

	cfg, err := thumbnail.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("설정 로드 실패: %w", err)
	}

	bgPath, _ := cfg.ResolveTheme(worshipType, date)

	execPath := path.ExecutePath("easyPreparation")
	outPath := filepath.Join(execPath, "data", "templates", "thumbnail", "generated",
		fmt.Sprintf("%s_%s.png", date.Format("2006-01-02"), worshipType))

	// 배경 경로가 상대 경로이면 절대 경로로 변환
	if bgPath != "" && !filepath.IsAbs(bgPath) {
		bgPath = filepath.Join(execPath, bgPath)
	}

	// dateLabel 빌드: "26.04.05 주일예배" — 기념 주일이면 해당 레이블로 오버라이드
	worshipTypeLabels := map[string]string{
		"main_worship":  "주일예배",
		"after_worship": "오후예배",
		"wed_worship":   "수요예배",
		"fri_worship":   "금요예배",
	}
	typeLabel := worshipTypeLabels[worshipType]
	if typeLabel == "" {
		typeLabel = "예배"
	}
	dateStr := date.Format("2006-01-02")
	for _, s := range cfg.Specials {
		if s.Date == dateStr {
			if s.TitleOverride != "" {
				typeLabel = s.TitleOverride
			} else if s.Label != "" {
				typeLabel = s.Label
			}
			break
		}
	}
	dateLabel := date.Format("06.01.02") + " " + typeLabel

	// 말씀 제목 + 성경봉독: display 메모리(currentOrder) 우선, 없으면 config fallback
	sermonTitle, scripture := loadSermonDataFromOrder()
	if sermonTitle == "" && scripture == "" {
		configPath := filepath.Join(execPath, "config", worshipType+".json")
		sermonTitle, scripture = loadSermonDataFromConfig(configPath)
	}

	// 로고 설정 — 썸네일 전용 설정 우선, 없으면 Display 설정 fallback
	displayCfg := loadDisplayConfig()
	logoPath := findLogoPath()
	logoPosition := displayCfg.LogoPosition
	logoSizePercent := displayCfg.LogoSizePercent
	if cfg.LogoPosition != "" {
		logoPosition = cfg.LogoPosition
	}
	if cfg.LogoSizePercent > 0 {
		logoSizePercent = cfg.LogoSizePercent
	}

	return thumbnail.Generate(thumbnail.GenerateConfig{
		BackgroundPath:  bgPath,
		DateLabel:       dateLabel,
		SermonTitle:     sermonTitle,
		Scripture:       scripture,
		LogoPath:        logoPath,
		LogoPosition:    logoPosition,
		LogoSizePercent: logoSizePercent,
		FontName:        cfg.FontName,
		OutputPath:      outPath,
		Width:           1280,
		Height:          720,
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

	// 파일명 정리 (확장자 유지, 경로 순회 방지)
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".png"
	}
	// target 파라미터가 있으면 사용 (예: "default_main_worship")
	target := r.FormValue("target")
	var saveName string
	if target != "" {
		// target에서 경로 구분자 및 .. 제거
		saveName = filepath.Base(target) + ext
	} else {
		// 업로드 파일명에서 디렉토리 부분 제거
		saveName = filepath.Base(header.Filename)
	}

	// 파일명에 경로 순회 문자가 없는지 최종 확인
	if strings.Contains(saveName, "..") || saveName == "." || saveName == "" {
		http.Error(w, "잘못된 파일명", http.StatusBadRequest)
		return
	}

	// default_ 접두사면 data/templates/thumbnail/ 에 저장 (기본 배경 교체)
	var savePath, relPath string
	if len(target) > 8 && target[:8] == "default_" {
		typeName := filepath.Base(target[8:]) // "main_worship" 등 — Base로 경로 순회 방지
		if strings.Contains(typeName, "..") || typeName == "." || typeName == "" {
			http.Error(w, "잘못된 target", http.StatusBadRequest)
			return
		}
		savePath = filepath.Join(execPath, "data", "templates", "thumbnail", typeName+ext)
		relPath = "data/templates/thumbnail/" + typeName + ext
	} else {
		savePath = filepath.Join(specialDir, saveName)
		relPath = "data/templates/thumbnail/special/" + saveName
	}

	// 최종 경로가 허용 디렉토리 내에 있는지 검증
	cleanSavePath := filepath.Clean(savePath)
	thumbDir := filepath.Clean(filepath.Join(execPath, "data", "templates", "thumbnail"))
	if !strings.HasPrefix(cleanSavePath, thumbDir+string(filepath.Separator)) {
		http.Error(w, "허용되지 않는 저장 경로", http.StatusForbidden)
		return
	}
	savePath = cleanSavePath

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
	absPath := filepath.Clean(filepath.Join(execPath, relPath))

	// 보안: data/templates/thumbnail/ 하위만 허용
	thumbDir := filepath.Clean(filepath.Join(execPath, "data", "templates", "thumbnail"))
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
	absBase, err := filepath.Abs(base)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	// Clean으로 정규화 후 접두사 검사
	absBase = filepath.Clean(absBase)
	absTarget = filepath.Clean(absTarget)
	// target이 base 디렉토리 내의 파일이어야 함 (base 자체는 불허)
	return strings.HasPrefix(absTarget, absBase+string(filepath.Separator))
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

// ThumbnailGeneratedListHandler — GET /api/thumbnail/generated
// 생성된 썸네일 목록 반환: [{filename, date, worshipType, label, url}]
func ThumbnailGeneratedListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	execPath := path.ExecutePath("easyPreparation")
	genDir := filepath.Join(execPath, "data", "templates", "thumbnail", "generated")

	entries, err := os.ReadDir(genDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type GeneratedItem struct {
		Filename    string `json:"filename"`
		Date        string `json:"date"`
		WorshipType string `json:"worshipType"`
		Label       string `json:"label"`
		URL         string `json:"url"`
	}

	worshipTypeLabels := map[string]string{
		"main_worship":  "주일예배",
		"after_worship": "오후예배",
		"wed_worship":   "수요예배",
		"fri_worship":   "금요예배",
	}

	var items []GeneratedItem
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".png") {
			continue
		}
		// 파일명 형식: YYYY-MM-DD_worship_type.png
		base := strings.TrimSuffix(name, ".png")
		// 첫 번째 _ 로 날짜와 예배유형 분리
		idx := strings.Index(base, "_")
		if idx < 0 {
			continue
		}
		dateStr := base[:idx]
		worshipType := base[idx+1:]
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			continue
		}
		// 레이블: "2026.04.06 주일예배"
		typeLabel := worshipTypeLabels[worshipType]
		if typeLabel == "" {
			typeLabel = "예배"
		}
		dateParts := strings.Split(dateStr, "-")
		var labelDate string
		if len(dateParts) == 3 {
			labelDate = dateParts[0] + "." + dateParts[1] + "." + dateParts[2]
		} else {
			labelDate = dateStr
		}
		label := labelDate + " " + typeLabel
		url := "/api/thumbnail/preview?worshipType=" + worshipType + "&date=" + dateStr

		items = append(items, GeneratedItem{
			Filename:    name,
			Date:        dateStr,
			WorshipType: worshipType,
			Label:       label,
			URL:         url,
		})
	}

	// 날짜 기준 내림차순 정렬 (최신 먼저)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Date < items[j].Date {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	if items == nil {
		items = []GeneratedItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// ThumbnailGeneratedDeleteHandler — DELETE /api/thumbnail/generated?filename=xxx
// 특정 생성 썸네일 삭제 (DELETE 또는 POST 메서드 허용)
func ThumbnailGeneratedDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "filename 파라미터 없음", http.StatusBadRequest)
		return
	}

	// 경로 순회 방지
	filename = filepath.Base(filename)
	if strings.Contains(filename, "..") || filename == "." || filename == "" {
		http.Error(w, "잘못된 파일명", http.StatusBadRequest)
		return
	}

	execPath := path.ExecutePath("easyPreparation")
	targetPath := filepath.Join(execPath, "data", "templates", "thumbnail", "generated", filename)

	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "파일 없음", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
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
