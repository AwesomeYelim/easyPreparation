// CLI: 예배 영상을 YouTube에 업로드하고 썸네일을 자동 생성·설정합니다.
//
// 사용법:
//
//	go run ./cmd/upload -video <파일경로> -type main_worship
//	go run ./cmd/upload -video <파일경로> -type after_worship -privacy unlisted
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"easyPreparation_1.0/internal/thumbnail"
	"easyPreparation_1.0/internal/youtube"
)

func main() {
	videoPath := flag.String("video", "", "업로드할 동영상 파일 경로 (필수)")
	worshipType := flag.String("type", "main_worship", "예배 유형 (main_worship|after_worship|wed_worship|fri_worship)")
	customTitle := flag.String("title", "", "YouTube 영상 제목 (기본: 썸네일 설정에서 자동 생성)")
	desc := flag.String("desc", "", "영상 설명")
	privacy := flag.String("privacy", "public", "공개 설정 (public|unlisted|private)")
	flag.Parse()

	if *videoPath == "" {
		fmt.Fprintln(os.Stderr, "사용법: go run ./cmd/upload -video <파일경로> [-type main_worship] [-title 제목] [-privacy public]")
		os.Exit(1)
	}

	// YouTube 초기화
	youtube.Init(
		filepath.Join("config", "google_oauth.json"),
		filepath.Join("data", "youtube_token.json"),
	)
	if !youtube.Get().IsEnabled() {
		log.Fatal("[upload] YouTube 미연결 — config/google_oauth.json, data/youtube_token.json 확인")
	}

	date := time.Now()

	// 썸네일 생성
	thumbPath, err := generateThumb(*worshipType, date)
	if err != nil {
		log.Printf("[upload] 썸네일 생성 실패 (업로드는 계속): %v", err)
		thumbPath = ""
	} else {
		log.Printf("[upload] 썸네일 생성 완료: %s", thumbPath)
	}

	// 제목 결정
	title := *customTitle
	if title == "" {
		cfg, err := thumbnail.LoadConfig()
		if err == nil {
			sermonTitle, scripture := loadSermonFromConfig(*worshipType)
			_, title = cfg.ResolveTheme(*worshipType, date, sermonTitle, scripture)
		}
	}
	if title == "" {
		title = date.Format("2006-01-02") + " 예배"
	}

	log.Printf("[upload] 영상 업로드 시작")
	log.Printf("[upload]   파일:  %s", *videoPath)
	log.Printf("[upload]   제목:  %s", title)
	log.Printf("[upload]   공개:  %s", *privacy)

	videoID, err := youtube.UploadVideo(*videoPath, title, *desc, *privacy, thumbPath)
	if err != nil {
		log.Fatalf("[upload] 업로드 실패: %v", err)
	}

	fmt.Printf("\n업로드 완료! https://youtu.be/%s\n", videoID)
}

func generateThumb(worshipType string, date time.Time) (string, error) {
	cfg, err := thumbnail.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("설정 로드 실패: %w", err)
	}

	// config 파일에서 말씀 제목 + 성경봉독 로드
	sermonTitle, scripture := loadSermonFromConfig(worshipType)

	bgPath, _ := cfg.ResolveTheme(worshipType, date, sermonTitle, scripture)

	// 상대 경로 → 절대 경로 변환
	if bgPath != "" && !filepath.IsAbs(bgPath) {
		cwd, _ := os.Getwd()
		bgPath = filepath.Join(cwd, bgPath)
	}

	outDir := filepath.Join("data", "templates", "thumbnail", "generated")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}
	outPath := filepath.Join(outDir, fmt.Sprintf("%s_%s.png", date.Format("2006-01-02"), worshipType))

	typeLabels := map[string]string{
		"main_worship":  "주일예배",
		"after_worship": "오후예배",
		"wed_worship":   "수요예배",
		"fri_worship":   "금요예배",
	}
	typeLabel := typeLabels[worshipType]
	if typeLabel == "" {
		typeLabel = "예배"
	}
	dateStr := date.Format("2006-01-02")
	for _, s := range cfg.Specials {
		if s.Date == dateStr {
			if s.Label != "" {
				typeLabel = s.Label
			}
			break
		}
	}
	dateLabel := date.Format("06.01.02") + " " + typeLabel

	logoPath := findLogoPath()

	absPath, err := thumbnail.Generate(thumbnail.GenerateConfig{
		BackgroundPath:  bgPath,
		DateLabel:       dateLabel,
		SermonTitle:     sermonTitle,
		Scripture:       scripture,
		LogoPath:        logoPath,
		LogoPosition:    cfg.LogoPosition,
		LogoSizePercent: cfg.LogoSizePercent,
		TextStyles:      cfg.EffectiveTextStyles(),
		OutputPath:      outPath,
		Width:           1280,
		Height:          720,
	})
	return absPath, err
}

// loadSermonFromConfig — config/{type}.json 에서 말씀 제목과 성경봉독 로드
func loadSermonFromConfig(worshipType string) (title, scripture string) {
	data, err := os.ReadFile(filepath.Join("config", worshipType+".json"))
	if err != nil {
		return "", ""
	}

	var order struct {
		Entries []struct {
			Title string `json:"title"`
			Info  string `json:"info"`
			BEdit string `json:"b_edit"`
			REdit string `json:"r_edit"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(data, &order); err != nil {
		return "", ""
	}

	for _, e := range order.Entries {
		if e.Title == "설교" || e.Title == "말씀" {
			title = e.Info
		}
		if e.Title == "성경봉독" {
			if e.BEdit != "" {
				scripture = e.BEdit
			} else if e.REdit != "" {
				scripture = e.REdit
			}
		}
	}
	return title, scripture
}

// findLogoPath — data/logo.{png|jpg|jpeg|svg} 탐색
func findLogoPath() string {
	for _, ext := range []string{"png", "jpg", "jpeg", "svg"} {
		p := filepath.Join("data", "logo."+ext)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
