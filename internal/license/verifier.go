package license

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// StartBackgroundVerification — 24시간마다 라이선스 서버에 검증 요청
// context.Context를 받아 종료 시 중지
func StartBackgroundVerification(ctx context.Context) {
	go func() {
		// 시작 5분 후 첫 검증
		select {
		case <-time.After(5 * time.Minute):
			verifyWithServer()
		case <-ctx.Done():
			return
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				verifyWithServer()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// verifyWithServer — CF Worker /api/verify 엔드포인트에 라이선스 검증 요청
func verifyWithServer() {
	if os.Getenv("EASYPREP_DEV") == "true" {
		return // 개발모드: 서버 검증 스킵
	}

	cfg := GetServerConfig()
	if cfg == nil || cfg.ServerURL == "" {
		return // 오프라인 모드 — 검증 스킵
	}

	mgr := Get()
	if mgr == nil {
		return
	}

	info := mgr.GetLicense()
	if info == nil || info.Plan == PlanFree {
		return // Free 플랜은 검증 불필요
	}

	// POST /api/verify to CF Worker
	reqBody := map[string]string{
		"licenseKey": info.LicenseKey,
		"deviceId":   mgr.GetDeviceID(),
		"signature":  info.Signature,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(cfg.ServerURL+"/api/verify", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[license] 서버 검증 실패 (네트워크 오류, 오프라인 유예): %v", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Valid              bool   `json:"valid"`
		Plan               string `json:"plan"`
		ExpiresAt          string `json:"expiresAt"`
		SubscriptionStatus string `json:"subscriptionStatus"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[license] 서버 응답 파싱 실패: %v", err)
		return
	}

	if !result.Valid {
		// 라이선스 무효화 — Free로 전환
		log.Printf("[license] 서버 검증 실패 — 라이선스 비활성화")
		freeInfo := &LicenseInfo{
			Plan:     PlanFree,
			DeviceID: mgr.GetDeviceID(),
			ChurchID: 1,
			IssuedAt: time.Now(),
		}
		mgr.SetLicense(freeInfo)
		return
	}

	// last_verified 갱신 및 플랜/만료일 업데이트
	newPlan := info.Plan
	newExpires := info.ExpiresAt

	if result.Plan != "" {
		newPlan = Plan(result.Plan)
	}
	if result.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, result.ExpiresAt); err == nil {
			newExpires = t
		}
	}

	if err := mgr.UpdateVerification(newPlan, newExpires); err != nil {
		log.Printf("[license] 검증 갱신 저장 실패: %v", err)
		return
	}
	log.Printf("[license] 서버 검증 완료 (plan=%s, expires=%s)", newPlan, newExpires.Format("2006-01-02"))
}
