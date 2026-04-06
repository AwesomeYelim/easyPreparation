package license

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"easyPreparation_1.0/internal/path"
)

// GracePeriodDays — 만료 후 오프라인 유예 기간 (일)
const GracePeriodDays = 30

// resolveCachePath — data/license.json 절대 경로 반환
func resolveCachePath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "license.json")
}

// saveCache — LicenseInfo를 data/license.json에 JSON으로 저장
func saveCache(info *LicenseInfo) error {
	cachePath := resolveCachePath()

	// 디렉터리 생성 보장
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return fmt.Errorf("캐시 디렉터리 생성 실패: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("라이선스 직렬화 실패: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("라이선스 파일 저장 실패: %w", err)
	}

	log.Printf("[license] 파일 캐시 저장 완료: %s", cachePath)
	return nil
}

// loadCache — data/license.json에서 LicenseInfo 로드
func loadCache() (*LicenseInfo, error) {
	cachePath := resolveCachePath()

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 파일 없음 — 정상 (첫 실행)
		}
		return nil, fmt.Errorf("라이선스 캐시 읽기 실패: %w", err)
	}

	var info LicenseInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("라이선스 캐시 파싱 실패: %w", err)
	}

	return &info, nil
}

// generateDeviceID — MAC 주소 기반 디바이스 고유 ID 생성
// 동일 머신에서 항상 동일한 ID를 반환 (재설치해도 유지)
func generateDeviceID() string {
	id, err := collectHardwareFingerprint()
	if err != nil || id == "" {
		// fallback: 호스트명 + OS 정보 해시
		hostname, _ := os.Hostname()
		id = fmt.Sprintf("%s-%s-%s", hostname, runtime.GOOS, runtime.GOARCH)
	}

	// SHA256 해시로 정규화
	h := sha256.Sum256([]byte(id))
	hex := fmt.Sprintf("%x", h[:8]) // 16자 (8바이트)
	return fmt.Sprintf("EP-%s-%s-%s-%s",
		strings.ToUpper(hex[0:4]),
		strings.ToUpper(hex[4:8]),
		strings.ToUpper(hex[8:12]),
		strings.ToUpper(hex[12:16]),
	)
}

// collectHardwareFingerprint — 네트워크 인터페이스에서 하드웨어 식별 정보 수집
func collectHardwareFingerprint() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var macs []string
	for _, iface := range interfaces {
		// 루프백·가상 인터페이스 제외
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if strings.HasPrefix(iface.Name, "veth") ||
			strings.HasPrefix(iface.Name, "docker") ||
			strings.HasPrefix(iface.Name, "br-") ||
			strings.HasPrefix(iface.Name, "virbr") {
			continue
		}
		mac := iface.HardwareAddr.String()
		if mac != "" && mac != "00:00:00:00:00:00" {
			macs = append(macs, mac)
		}
	}

	if len(macs) == 0 {
		return "", fmt.Errorf("유효한 MAC 주소 없음")
	}

	// 정렬 없이 첫 번째 물리 인터페이스 MAC 사용 (재현성 보장)
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s|%s|%s|%s", macs[0], hostname, runtime.GOOS, runtime.GOARCH), nil
}
