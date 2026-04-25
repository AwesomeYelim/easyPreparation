package handlers

import (
	"easyPreparation_1.0/internal/path"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// DisplayConfig — Display 전역 설정
type DisplayConfig struct {
	Font string `json:"font"` // "default" | "noto-sans-kr" | "gowun-dodum" | "nanum-myeongjo" | "black-han-sans"

	// 오버레이 커스터마이징
	OverlayBgOpacity float64 `json:"overlayBgOpacity"` // 0.0 ~ 1.0, 기본 0.75
	OverlayTextColor string  `json:"overlayTextColor"` // "#ffffff"
	OverlayPosition  string  `json:"overlayPosition"`  // "flex-end" | "center" | "flex-start"
	OverlayFontScale float64 `json:"overlayFontScale"` // 0.5 ~ 2.0, 기본 1.0

	// 비디오 배경
	GlobalVideoBg string `json:"globalVideoBg,omitempty"` // 파일명 (data/video-bg/ 하위)

	// 프로젝터 로고 위치/크기
	LogoPosition    string  `json:"logoPosition,omitempty"`    // "bottom-right" | "bottom-left" | "top-right" | "top-left"
	LogoSizePercent float64 `json:"logoSizePercent,omitempty"` // 5 ~ 30 (vw%), 기본 18
}

func displayConfigPath() string {
	execPath := path.ExecutePath("easyPreparation")
	return filepath.Join(execPath, "data", "display_config.json")
}

func loadDisplayConfig() DisplayConfig {
	data, err := os.ReadFile(displayConfigPath())
	if err != nil {
		return defaultDisplayConfig()
	}
	var cfg DisplayConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultDisplayConfig()
	}
	applyDisplayConfigDefaults(&cfg)
	return cfg
}

func defaultDisplayConfig() DisplayConfig {
	return DisplayConfig{
		Font:             "default",
		OverlayBgOpacity: 0.75,
		OverlayTextColor: "#ffffff",
		OverlayPosition:  "flex-end",
		OverlayFontScale: 1.0,
		LogoPosition:     "bottom-right",
		LogoSizePercent:  18,
	}
}

func applyDisplayConfigDefaults(cfg *DisplayConfig) {
	if cfg.Font == "" {
		cfg.Font = "default"
	}
	if cfg.OverlayBgOpacity == 0 {
		cfg.OverlayBgOpacity = 0.75
	}
	if cfg.OverlayTextColor == "" {
		cfg.OverlayTextColor = "#ffffff"
	}
	if cfg.OverlayPosition == "" {
		cfg.OverlayPosition = "flex-end"
	}
	if cfg.OverlayFontScale == 0 {
		cfg.OverlayFontScale = 1.0
	}
	if cfg.LogoPosition == "" {
		cfg.LogoPosition = "bottom-right"
	}
	if cfg.LogoSizePercent == 0 {
		cfg.LogoSizePercent = 18
	}
}

// PDFLogoConfig — 예배 PDF에서 사용할 로고 설정 (DisplayConfig에서 추출)
type PDFLogoConfig struct {
	LogoPath        string  // 로고 파일 절대 경로 (없으면 "")
	LogoPosition    string  // "bottom-right" | "bottom-left" | "top-right" | "top-left"
	LogoSizePercent float64 // 5 ~ 30 (%, 기본 18)
}

// GetPDFLogoConfig — PDF 생성 시 로고 설정 조회 (exported)
func GetPDFLogoConfig() PDFLogoConfig {
	cfg := loadDisplayConfig()
	return PDFLogoConfig{
		LogoPath:        findLogoPath(),
		LogoPosition:    cfg.LogoPosition,
		LogoSizePercent: cfg.LogoSizePercent,
	}
}

// HandleDisplayConfigGet — GET /api/display-config
func HandleDisplayConfigGet(w http.ResponseWriter, r *http.Request) {
	cfg := loadDisplayConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// HandleDisplayConfigSet — PUT /api/display-config
func HandleDisplayConfigSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var cfg DisplayConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "JSON 파싱 실패", http.StatusBadRequest)
		return
	}
	allowedFont := map[string]bool{
		"default":        true,
		"noto-sans-kr":   true,
		"gowun-dodum":    true,
		"nanum-myeongjo": true,
		"black-han-sans": true,
	}
	if !allowedFont[cfg.Font] {
		cfg.Font = "default"
	}
	// 오버레이 범위 검증
	if cfg.OverlayBgOpacity < 0 || cfg.OverlayBgOpacity > 1 {
		cfg.OverlayBgOpacity = 0.75
	}
	if cfg.OverlayFontScale < 0.5 || cfg.OverlayFontScale > 2.0 {
		cfg.OverlayFontScale = 1.0
	}
	allowedPos := map[string]bool{"flex-end": true, "center": true, "flex-start": true}
	if !allowedPos[cfg.OverlayPosition] {
		cfg.OverlayPosition = "flex-end"
	}
	if cfg.OverlayTextColor == "" {
		cfg.OverlayTextColor = "#ffffff"
	}
	// 비디오 배경 — 경로 traversal 방지
	if cfg.GlobalVideoBg != "" {
		if filepath.Base(cfg.GlobalVideoBg) != cfg.GlobalVideoBg {
			cfg.GlobalVideoBg = ""
		}
	}
	// 로고 위치 검증
	allowedLogoPos := map[string]bool{"bottom-right": true, "bottom-left": true, "top-right": true, "top-left": true}
	if !allowedLogoPos[cfg.LogoPosition] {
		cfg.LogoPosition = "bottom-right"
	}
	if cfg.LogoSizePercent < 5 || cfg.LogoSizePercent > 30 {
		cfg.LogoSizePercent = 18
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		http.Error(w, "JSON 직렬화 실패", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(displayConfigPath(), data, 0644); err != nil {
		http.Error(w, "설정 저장 실패", http.StatusInternalServerError)
		return
	}

	// Display/Overlay 페이지에 실시간 반영
	BroadcastMessage("display_config", map[string]interface{}{
		"font":             cfg.Font,
		"overlayBgOpacity": cfg.OverlayBgOpacity,
		"overlayTextColor": cfg.OverlayTextColor,
		"overlayPosition":  cfg.OverlayPosition,
		"overlayFontScale": cfg.OverlayFontScale,
		"globalVideoBg":    cfg.GlobalVideoBg,
		"logoPosition":     cfg.LogoPosition,
		"logoSizePercent":  cfg.LogoSizePercent,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}
