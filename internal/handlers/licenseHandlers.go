package handlers

import (
	"bytes"
	"easyPreparation_1.0/internal/license"
	"encoding/json"
	"net/http"
	"time"
)

// licenseStatusResponse — /api/license 공통 응답 구조
func licenseStatusResponse(mgr *license.Manager) map[string]interface{} {
	plan := license.PlanFree
	var expiresAt *time.Time
	var daysRemaining int
	var deviceID string
	var gracePeriod bool
	isActive := false

	if mgr != nil {
		plan = mgr.GetPlan()
		deviceID = mgr.GetDeviceID()
		gracePeriod = mgr.IsInGracePeriod()

		info := mgr.GetLicense()
		if info != nil {
			if !info.ExpiresAt.IsZero() {
				expiresAt = &info.ExpiresAt
			}
			daysRemaining = mgr.DaysUntilExpiry()
			isActive = !mgr.IsExpired() || gracePeriod
		}
	}

	// 활성화된 기능 목록
	features := license.PlanFeatures[plan]
	featureStrs := make([]string, 0, len(features))
	for _, f := range features {
		featureStrs = append(featureStrs, string(f))
	}

	resp := map[string]interface{}{
		"plan":           string(plan),
		"features":       featureStrs,
		"device_id":      deviceID,
		"grace_period":   gracePeriod,
		"is_active":      isActive,
		"days_remaining": daysRemaining,
	}
	if expiresAt != nil {
		resp["expires_at"] = expiresAt.Format(time.RFC3339)
	} else {
		resp["expires_at"] = nil
	}
	return resp
}

// LicenseStatusHandler — GET /api/license
func LicenseStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(licenseStatusResponse(license.Get()))
}

// LicenseActivateHandler — POST /api/license/activate
// 서버 검증 우선, 네트워크 오류 시 오프라인 fallback
func LicenseActivateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		LicenseKey string `json:"license_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if !license.ValidateKeyFormat(body.LicenseKey) {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_key_format",
			"message": "라이선스 키 형식이 올바르지 않습니다. (EP-XXXX-XXXX-XXXX-XXXX)",
		})
		return
	}

	mgr := license.Get()
	deviceID := ""
	if mgr != nil {
		deviceID = mgr.GetDeviceID()
	}

	// 1. 서버 검증 시도
	cfg := license.GetServerConfig()
	if cfg != nil && cfg.ServerURL != "" {
		reqBody, _ := json.Marshal(map[string]string{
			"licenseKey": body.LicenseKey,
			"deviceId":   deviceID,
			"signature":  "",
		})

		client := &http.Client{Timeout: 15 * time.Second}
		serverResp, err := client.Post(cfg.ServerURL+"/api/verify", "application/json", bytes.NewReader(reqBody))
		if err == nil {
			defer serverResp.Body.Close()
			var result struct {
				Valid     bool   `json:"valid"`
				Plan      string `json:"plan"`
				ExpiresAt string `json:"expiresAt"`
			}
			if json.NewDecoder(serverResp.Body).Decode(&result) == nil {
				if result.Valid {
					// 서버 검증 성공 — 라이선스 활성화
					now := time.Now()
					expiresAt := now.AddDate(1, 0, 0)
					if result.ExpiresAt != "" {
						if t, err := time.Parse(time.RFC3339, result.ExpiresAt); err == nil {
							expiresAt = t
						}
					}
					plan := license.PlanPro
					if result.Plan != "" {
						plan = license.Plan(result.Plan)
					}
					info := &license.LicenseInfo{
						LicenseKey:   body.LicenseKey,
						Plan:         plan,
						DeviceID:     deviceID,
						ChurchID:     1,
						IssuedAt:     now,
						ExpiresAt:    expiresAt,
						LastVerified: now,
					}
					if mgr != nil {
						mgr.SetLicense(info)
					}
					respMap := licenseStatusResponse(mgr)
					respMap["activated"] = true
					respondJSON(w, http.StatusOK, respMap)
					return
				}
				// 서버에서 유효하지 않다고 응답 (valid=false)
				respondJSON(w, http.StatusBadRequest, map[string]string{
					"error":   "invalid_key",
					"message": "유효하지 않은 라이선스 키입니다.",
				})
				return
			}
		}
		// 네트워크 오류 — 오프라인 fallback으로 진행
		// (err != nil 인 경우 아래 fallback 실행)
	}

	// 2. 오프라인 fallback (서버 설정 없거나 네트워크 오류)
	// 유효한 형식의 키는 Pro 1년 라이선스로 활성화
	now := time.Now()
	expiresAt := now.AddDate(1, 0, 0)

	info := &license.LicenseInfo{
		LicenseKey:   body.LicenseKey,
		Plan:         license.PlanPro,
		DeviceID:     deviceID,
		ChurchID:     1,
		IssuedAt:     now,
		ExpiresAt:    expiresAt,
		LastVerified: now,
	}

	if mgr != nil {
		if err := mgr.SetLicense(info); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "라이선스 저장에 실패했습니다."})
			return
		}
	}

	respMap := licenseStatusResponse(mgr)
	respMap["activated"] = true
	respMap["warning"] = "오프라인 모드로 활성화되었습니다."
	respondJSON(w, http.StatusOK, respMap)
}

// LicenseDeactivateHandler — POST /api/license/deactivate
func LicenseDeactivateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	mgr := license.Get()
	if mgr != nil {
		// Free 플랜으로 되돌리기 — 빈 LicenseInfo 저장
		freeInfo := &license.LicenseInfo{
			LicenseKey: "",
			Plan:       license.PlanFree,
			DeviceID:   mgr.GetDeviceID(),
			ChurchID:   1,
			IssuedAt:   time.Now(),
		}
		if err := mgr.SetLicense(freeInfo); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "라이선스 비활성화에 실패했습니다."})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := licenseStatusResponse(mgr)
	resp["deactivated"] = true
	_ = json.NewEncoder(w).Encode(resp)
}

// LicenseVerifyHandler — POST /api/license/verify
func LicenseVerifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	mgr := license.Get()
	if mgr != nil {
		info := mgr.GetLicense()
		if info != nil {
			// last_verified 타임스탬프 갱신
			info.LastVerified = time.Now()
			if err := mgr.SetLicense(info); err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "인증 갱신에 실패했습니다."})
				return
			}
		}
	}

	resp := licenseStatusResponse(mgr)
	resp["verified"] = true
	respondJSON(w, http.StatusOK, resp)
}

// LicenseCheckoutHandler — POST /api/license/checkout
// 클라이언트에서 plan (pro_monthly | pro_annual) 수신 → CF Worker에 checkout 세션 요청 → URL 반환
func LicenseCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Plan string `json:"plan"` // "pro_monthly" or "pro_annual"
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if body.Plan != "pro_monthly" && body.Plan != "pro_annual" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid plan"})
		return
	}

	cfg := license.GetServerConfig()
	if cfg == nil || cfg.ServerURL == "" {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "라이선스 서버가 설정되지 않았습니다."})
		return
	}

	mgr := license.Get()
	deviceID := ""
	if mgr != nil {
		deviceID = mgr.GetDeviceID()
	}

	reqBody, _ := json.Marshal(map[string]string{
		"deviceId": deviceID,
		"plan":     body.Plan,
	})

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(cfg.ServerURL+"/api/checkout", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": "결제 서버 연결 실패"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": "서버 응답 파싱 실패"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_ = json.NewEncoder(w).Encode(result)
}

// LicenseCallbackHandler — POST /api/license/callback
// 결제 완료 후 폴링: sessionId로 CF Worker에 activate 요청
func LicenseCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	cfg := license.GetServerConfig()
	if cfg == nil || cfg.ServerURL == "" {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "라이선스 서버가 설정되지 않았습니다."})
		return
	}

	mgr := license.Get()
	deviceID := ""
	if mgr != nil {
		deviceID = mgr.GetDeviceID()
	}

	reqBody, _ := json.Marshal(map[string]string{
		"sessionId": body.SessionID,
		"deviceId":  deviceID,
	})

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(cfg.ServerURL+"/api/activate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": "결제 서버 연결 실패"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Status     string `json:"status"`     // "pending" | "completed"
		LicenseKey string `json:"licenseKey"`
		Plan       string `json:"plan"`
		ExpiresAt  string `json:"expiresAt"`
		Signature  string `json:"signature"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": "서버 응답 파싱 실패"})
		return
	}

	if result.Status == "completed" && result.LicenseKey != "" {
		now := time.Now()
		expiresAt := now.AddDate(1, 0, 0)
		if result.ExpiresAt != "" {
			if t, err := time.Parse(time.RFC3339, result.ExpiresAt); err == nil {
				expiresAt = t
			}
		}

		info := &license.LicenseInfo{
			LicenseKey:   result.LicenseKey,
			Plan:         license.PlanPro,
			DeviceID:     deviceID,
			ChurchID:     1,
			IssuedAt:     now,
			ExpiresAt:    expiresAt,
			LastVerified: now,
			Signature:    result.Signature,
		}
		if mgr != nil {
			mgr.SetLicense(info)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     result.Status,
		"plan":       result.Plan,
		"licenseKey": result.LicenseKey,
	})
}

// LicensePortalHandler — POST /api/license/portal
// 결제 정보 조회 + 구독 관리
func LicensePortalHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := license.GetServerConfig()
	if cfg == nil || cfg.ServerURL == "" {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "라이선스 서버가 설정되지 않았습니다."})
		return
	}

	mgr := license.Get()
	if mgr == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "라이선스 매니저 미초기화"})
		return
	}

	info := mgr.GetLicense()
	if info == nil || info.LicenseKey == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "활성화된 라이선스가 없습니다."})
		return
	}

	reqBody, _ := json.Marshal(map[string]string{
		"licenseKey": info.LicenseKey,
		"deviceId":   mgr.GetDeviceID(),
	})

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(cfg.ServerURL+"/api/portal", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": "결제 서버 연결 실패"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": "서버 응답 파싱 실패"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_ = json.NewEncoder(w).Encode(result)
}

// respondJSON — 간편 JSON 응답 헬퍼
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
