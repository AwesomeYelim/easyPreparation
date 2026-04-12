package thumbnail

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"easyPreparation_1.0/internal/path"
)

// ThumbnailConfig — 썸네일 전체 설정
type ThumbnailConfig struct {
	Defaults map[string]DefaultTheme `json:"defaults"`
	Specials []SpecialDate           `json:"specials"`
}

// DefaultTheme — 예배 유형별 기본 테마
type DefaultTheme struct {
	Background  string `json:"background"`
	TitleFormat string `json:"titleFormat"`
}

// SpecialDate — 기념 주일 설정
type SpecialDate struct {
	Date          string `json:"date"`          // "2026-04-05"
	Label         string `json:"label"`         // "부활절 예배"
	Background    string `json:"background"`    // 커스텀 배경 경로
	TitleOverride string `json:"titleOverride"` // 타이틀 전체 오버라이드
}

func thumbnailConfigPath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "thumbnail_config.json")
}

// LoadConfig — 설정 파일 로드 (없으면 기본값 반환)
func LoadConfig() (*ThumbnailConfig, error) {
	data, err := os.ReadFile(thumbnailConfigPath())
	if err != nil {
		cfg := defaultConfig()
		_ = SaveConfig(cfg)
		return cfg, nil
	}
	var cfg ThumbnailConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("썸네일 설정 파싱 실패: %w", err)
	}
	return &cfg, nil
}

// SaveConfig — 설정 파일 저장
func SaveConfig(cfg *ThumbnailConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(thumbnailConfigPath()), 0755); err != nil {
		return err
	}
	return os.WriteFile(thumbnailConfigPath(), data, 0644)
}

// ResolveTheme — 예배 유형 + 날짜로 배경/타이틀 결정 (기념 주일은 주일예배만 적용)
func (c *ThumbnailConfig) ResolveTheme(worshipType string, date time.Time) (bgPath, title string) {
	dateStr := date.Format("2006-01-02")

	// 기념 주일 체크 — main_worship에만 적용
	if worshipType == "main_worship" {
		for _, s := range c.Specials {
			if s.Date == dateStr {
				bg := s.Background
				if bg == "" {
					if d, ok := c.Defaults[worshipType]; ok {
						bg = d.Background
					}
				}
				label := s.TitleOverride
				if label == "" {
					label = s.Label
				}
				// "N월 N째주 {기념예배}" 형식
				t := FormatTitle("{month}월 {weekOrd} "+label, date)
				return bg, t
			}
		}
	}

	// 기본 테마
	if d, ok := c.Defaults[worshipType]; ok {
		return d.Background, FormatTitle(d.TitleFormat, date)
	}

	return "", FormatTitle("{month}월 {weekOrd} 예배", date)
}

// FormatTitle — titleFormat 변수를 실제 값으로 치환
func FormatTitle(format string, date time.Time) string {
	r := strings.NewReplacer(
		"{year}", fmt.Sprintf("%d", date.Year()),
		"{month}", fmt.Sprintf("%d", date.Month()),
		"{day}", fmt.Sprintf("%d", date.Day()),
		"{weekOrd}", weekOrdinal(date),
	)
	return r.Replace(format)
}

// weekOrdinal — 해당 월의 N째주 반환 ("첫째주", "둘째주", ...)
func weekOrdinal(date time.Time) string {
	day := date.Day()
	week := (day-1)/7 + 1
	ordinals := []string{"첫째주", "둘째주", "셋째주", "넷째주", "다섯째주"}
	if week >= 1 && week <= 5 {
		return ordinals[week-1]
	}
	return fmt.Sprintf("%d째주", week)
}

func defaultConfig() *ThumbnailConfig {
	return &ThumbnailConfig{
		Defaults: map[string]DefaultTheme{
			"main_worship":  {Background: "data/templates/thumbnail/main_worship.png", TitleFormat: "{month}월 {weekOrd} 주일예배"},
			"after_worship": {Background: "data/templates/thumbnail/after_worship.png", TitleFormat: "{month}월 {weekOrd} 오후예배"},
			"wed_worship":   {Background: "data/templates/thumbnail/wed_worship.png", TitleFormat: "{month}월 {weekOrd} 수요예배"},
			"fri_worship":   {Background: "data/templates/thumbnail/fri_worship.png", TitleFormat: "{month}월 {weekOrd} 금요예배"},
		},
		Specials: []SpecialDate{},
	}
}
