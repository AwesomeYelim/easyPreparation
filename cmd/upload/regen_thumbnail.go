//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"easyPreparation_1.0/internal/thumbnail"
	"easyPreparation_1.0/internal/youtube"
)

func main() {
	cfg, err := thumbnail.LoadConfig()
	if err != nil {
		log.Fatalf("설정 로드 실패: %v", err)
	}

	youtube.Init(
		filepath.Join("config", "google_oauth.json"),
		filepath.Join("data", "youtube_token.json"),
	)

	targets := []struct {
		worshipType string
		date        time.Time
		videoID     string
	}{
		// main_worship 이미 업로드 완료 — 스킵
		// {"main_worship", time.Date(2026, 5, 17, 0, 0, 0, 0, time.Local), "E4wMNUUCqK8"},
		{"after_worship", time.Date(2026, 5, 17, 0, 0, 0, 0, time.Local), "A7S2d-hDVlY"},
	}

	for _, t := range targets {
		bgPath, _ := cfg.ResolveTheme(t.worshipType, t.date)
		if bgPath != "" && !filepath.IsAbs(bgPath) {
			cwd, _ := os.Getwd()
			bgPath = filepath.Join(cwd, bgPath)
		}

		outDir := filepath.Join("data", "templates", "thumbnail", "generated")
		os.MkdirAll(outDir, 0755)
		outPath := filepath.Join(outDir, fmt.Sprintf("%s_%s.png", t.date.Format("2006-01-02"), t.worshipType))

		labels := map[string]string{"main_worship": "주일예배", "after_worship": "오후예배"}
		dateLabel := t.date.Format("06.01.02") + " " + labels[t.worshipType]

		sermonTitle, scripture := loadSermon(t.worshipType)
		// after_worship 실제 본문 수동 지정 (config가 다른 날짜 데이터)
		if t.worshipType == "after_worship" {
			sermonTitle = "고린도후서 1:17-20"
			scripture = "고린도후서 1:17-20"
		}
		logoPath := findLogo()

		ts := &thumbnail.TextStyles{
			Header: thumbnail.TextStyle{FontName: "NanumGothicBold", Size: 50, Color: "#ffffff"},
			Main:   thumbnail.TextStyle{FontName: "NanumGothicBold", Size: 90, Color: "#ffffff"},
			Footer: thumbnail.TextStyle{FontName: "NanumGothic", Size: 45, Color: "#ffffff"},
		}
		absPath, err := thumbnail.Generate(thumbnail.GenerateConfig{
			BackgroundPath:  bgPath,
			DateLabel:       dateLabel,
			SermonTitle:     sermonTitle,
			Scripture:       scripture,
			LogoPath:        logoPath,
			LogoPosition:    cfg.LogoPosition,
			LogoSizePercent: cfg.LogoSizePercent,
			TextStyles:      ts,
			OutputPath:      outPath,
			Width:           1280,
			Height:          720,
		})
		if err != nil {
			log.Fatalf("[%s] 썸네일 생성 실패: %v", t.worshipType, err)
		}
		log.Printf("[%s] 생성 완료: %s (SermonTitle=%q, Scripture=%q)", t.worshipType, absPath, sermonTitle, scripture)

		for attempt := 1; attempt <= 5; attempt++ {
			err := youtube.UploadThumbnailToBroadcast(t.videoID, absPath)
			if err == nil {
				fmt.Printf("완료: https://youtu.be/%s\n", t.videoID)
				break
			}
			if attempt == 5 {
				log.Fatalf("[%s] 업로드 실패 (5회): %v", t.worshipType, err)
			}
			wait := time.Duration(attempt*30) * time.Second
			log.Printf("[%s] 429 재시도 %d/5 — %v 후 재시도: %v", t.worshipType, attempt, wait, err)
			time.Sleep(wait)
		}
	}
}

func cleanBibleRef(ref string) string {
	if i := strings.Index(ref, "_"); i != -1 {
		if j := strings.Index(ref[i:], "/"); j != -1 {
			return ref[:i] + " " + ref[i+j+1:]
		}
	}
	return ref
}

func loadSermon(worshipType string) (sermonTitle, scripture string) {
	data, err := os.ReadFile(filepath.Join("config", worshipType+".json"))
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
		obj = cleanBibleRef(obj)
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

func findLogo() string {
	for _, ext := range []string{"png", "jpg", "jpeg", "svg"} {
		p := filepath.Join("data", "logo."+ext)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
