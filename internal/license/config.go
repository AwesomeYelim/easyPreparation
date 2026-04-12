package license

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// LicenseServerConfig — config/license.json에서 로드
type LicenseServerConfig struct {
	ServerURL  string `json:"server_url"`  // CF Worker URL (예: https://easyprep-license-api.workers.dev)
	HMACSecret string `json:"hmac_secret"` // HMAC 서명 시크릿
}

var serverConfig *LicenseServerConfig

// LoadServerConfig — config/license.json 파일에서 설정 로드
// 파일 없으면 nil (오프라인 모드)
func LoadServerConfig(configDir string) {
	cfgPath := filepath.Join(configDir, "license.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[license] config/license.json 없음 — 오프라인 모드로 동작")
		} else {
			log.Printf("[license] config/license.json 읽기 실패: %v", err)
		}
		return
	}

	var cfg LicenseServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("[license] config/license.json 파싱 실패: %v", err)
		return
	}

	serverConfig = &cfg
	log.Printf("[license] 라이선스 서버 설정 로드 완료: %s", cfg.ServerURL)
}

// GetServerConfig — 설정 반환 (nil 가능)
func GetServerConfig() *LicenseServerConfig {
	return serverConfig
}

// GetHMACSecret — 설정 파일의 시크릿 반환, 없으면 기본값 사용
func GetHMACSecret() string {
	if serverConfig != nil && serverConfig.HMACSecret != "" {
		return serverConfig.HMACSecret
	}
	return defaultHMACSecret
}
