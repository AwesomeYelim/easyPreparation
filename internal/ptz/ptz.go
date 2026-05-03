package ptz

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"easyPreparation_1.0/internal/path"
)

// Preset — 카메라에 저장된 프리셋 정보
type Preset struct {
	Token string `json:"token"` // 프리셋 번호 문자열 ("1", "2", ...)
	Name  string `json:"name"`  // 이름 (e.g. "P1", "강대상")
}

// Config — PTZ 카메라 연결 설정
type Config struct {
	IP           string            `json:"ip"`
	Port         int               `json:"port"`         // 기본 80
	Username     string            `json:"username"`     // 기본 "admin"
	Password     string            `json:"password"`     // 기본 "admin"
	ProfileToken string            `json:"profileToken"` // ONVIF 프로파일 토큰 (미사용, 호환성 유지)
	StreamPath   string            `json:"streamPath"`   // MJPEG 스트림 경로, 기본 "/stream"
	PresetCount  int               `json:"presetCount"`  // 사용 가능한 프리셋 수 (기본 5, 최대 20)
	PresetNames  map[string]string `json:"presetNames"`  // 프리셋 커스텀 이름 {"1":"강대상","2":"피아노"}
	Enabled      bool              `json:"enabled"`
}

func configPath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "ptz_config.json")
}

// LoadConfig — 설정 로드 (없으면 기본값 반환)
func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return defaultConfig(), nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	applyDefaults(&cfg)
	return &cfg, nil
}

// SaveConfig — 설정 저장
func SaveConfig(cfg *Config) error {
	applyDefaults(cfg)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configPath()), 0755); err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

// GotoPreset — 카메라 CGI API로 프리셋 이동
// preset: 1~N (1-indexed), 0이면 no-op
func GotoPreset(preset int) error {
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("PTZ 설정 로드 실패: %w", err)
	}
	if !cfg.Enabled || cfg.IP == "" || preset <= 0 {
		return nil
	}

	port := cfg.Port
	if port == 0 {
		port = 80
	}

	// 이 카메라는 CGI API 사용: /cgi-bin/ptzctrl.cgi?ptzcmd&poscall&{preset}
	url := fmt.Sprintf("http://%s:%d/cgi-bin/ptzctrl.cgi?ptzcmd&poscall&%d", cfg.IP, port, preset)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("PTZ 요청 실패: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	log.Printf("[ptz] GotoPreset(%d) → %s", preset, resp.Status)
	return nil
}

// GetPresets — 카메라에서 사용 가능한 프리셋 목록 반환
// 이 카메라는 프리셋 목록 조회 API가 없으므로 PresetCount 기반으로 번호 생성
func GetPresets() ([]Preset, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("PTZ 설정 로드 실패: %w", err)
	}
	if cfg.IP == "" {
		return nil, fmt.Errorf("IP 주소가 설정되지 않았습니다")
	}

	// 카메라 연결 확인 (ping)
	if _, err := Ping(); err != nil {
		return nil, fmt.Errorf("카메라 연결 실패: %w", err)
	}

	count := cfg.PresetCount
	if count <= 0 {
		count = 5
	}
	if count > 20 {
		count = 20
	}

	presets := make([]Preset, count)
	for i := 0; i < count; i++ {
		n := i + 1 // 1-indexed
		token := fmt.Sprintf("%d", n)
		name := fmt.Sprintf("P%d", n)
		if cfg.PresetNames != nil {
			if custom, ok := cfg.PresetNames[token]; ok && custom != "" {
				name = custom
			}
		}
		presets[i] = Preset{Token: token, Name: name}
	}
	log.Printf("[ptz] GetPresets → %d개 (presetCount=%d)", len(presets), count)
	return presets, nil
}

// Ping — 카메라 HTTP 응답 여부 확인. 응답 시간(ms) 반환.
// 401 Unauthorized도 "연결됨"으로 처리 (Digest 인증 카메라)
func Ping() (latencyMs int64, err error) {
	cfg, err := LoadConfig()
	if err != nil {
		return 0, fmt.Errorf("PTZ 설정 로드 실패: %w", err)
	}
	if cfg.IP == "" {
		return 0, fmt.Errorf("IP 주소가 설정되지 않았습니다")
	}

	port := cfg.Port
	if port == 0 {
		port = 80
	}

	url := fmt.Sprintf("http://%s:%d/", cfg.IP, port)
	start := time.Now()

	client := &http.Client{
		Timeout: 3 * time.Second,
		// 리다이렉트 비활성화
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("카메라 연결 실패: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	ms := time.Since(start).Milliseconds()
	// 401 Unauthorized = 연결됨 (Digest 인증 카메라)
	if resp.StatusCode >= 500 {
		return ms, fmt.Errorf("카메라 응답 오류: %s", resp.Status)
	}
	return ms, nil
}

func defaultConfig() *Config {
	return &Config{
		Port:         80,
		Username:     "admin",
		Password:     "admin",
		ProfileToken: "Profile_1",
		StreamPath:   "/stream",
		PresetCount:  5,
		Enabled:      false,
	}
}

func applyDefaults(cfg *Config) {
	if cfg.Port == 0 {
		cfg.Port = 80
	}
	if cfg.Username == "" {
		cfg.Username = "admin"
	}
	if cfg.Password == "" {
		cfg.Password = "admin"
	}
	if cfg.ProfileToken == "" {
		cfg.ProfileToken = "Profile_1"
	}
	if cfg.StreamPath == "" {
		cfg.StreamPath = "/stream"
	}
	if cfg.PresetCount <= 0 {
		cfg.PresetCount = 5
	}
}
