package handlers

import (
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
func LicenseActivateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		LicenseKey string `json:"license_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
		return
	}

	if !license.ValidateKeyFormat(body.LicenseKey) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
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

	// MVP: 유효한 형식의 키는 Pro 1년 라이선스로 활성화
	// 서명 검증은 선택적으로 수행하고, 실패해도 MVP 테스트용으로 허용
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

	// 서명 검증 시도 (실패해도 MVP에서는 허용 — warning만 로깅)
	signatureValid := license.ValidateSignature(info, "")
	_ = signatureValid // MVP에서는 서명 실패도 허용

	if mgr != nil {
		if err := mgr.SetLicense(info); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "라이선스 저장에 실패했습니다."})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := licenseStatusResponse(mgr)
	resp["activated"] = true
	if !signatureValid {
		resp["warning"] = "서명 검증 없이 활성화되었습니다 (MVP 테스트 모드)"
	}
	_ = json.NewEncoder(w).Encode(resp)
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
			// MVP: last_verified 타임스탬프만 갱신
			info.LastVerified = time.Now()
			if err := mgr.SetLicense(info); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "인증 갱신에 실패했습니다."})
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := licenseStatusResponse(mgr)
	resp["verified"] = true
	_ = json.NewEncoder(w).Encode(resp)
}
