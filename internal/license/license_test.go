package license

import (
	"testing"
	"time"
)

// ─── 1. PlanFeatures 매핑 테스트 ───────────────────────────────────────────

func TestPlanFeatures_FreeHasNoFeatures(t *testing.T) {
	features := PlanFeatures[PlanFree]
	if len(features) != 0 {
		t.Errorf("PlanFree는 기능이 없어야 함, 실제: %v", features)
	}
}

func TestPlanFeatures_ProHasExpectedFeatures(t *testing.T) {
	expected := []Feature{
		FeatureOBSControl,
		FeatureAutoScheduler,
		FeatureYouTube,
		FeatureThumbnail,
		FeatureMultiWorship,
	}
	proFeatures := PlanFeatures[PlanPro]
	featureSet := make(map[Feature]bool, len(proFeatures))
	for _, f := range proFeatures {
		featureSet[f] = true
	}
	for _, want := range expected {
		if !featureSet[want] {
			t.Errorf("PlanPro에 %q 기능이 없음", want)
		}
	}
}

func TestPlanFeatures_ProDoesNotHaveCloudBackup(t *testing.T) {
	for _, f := range PlanFeatures[PlanPro] {
		if f == FeatureCloudBackup {
			t.Error("PlanPro에 FeatureCloudBackup이 포함되어 있음 — Enterprise 전용이어야 함")
		}
	}
}

func TestPlanFeatures_EnterpriseIncludesAllProFeatures(t *testing.T) {
	proSet := make(map[Feature]bool)
	for _, f := range PlanFeatures[PlanPro] {
		proSet[f] = true
	}
	enterpriseSet := make(map[Feature]bool)
	for _, f := range PlanFeatures[PlanEnterprise] {
		enterpriseSet[f] = true
	}
	for f := range proSet {
		if !enterpriseSet[f] {
			t.Errorf("PlanEnterprise에 Pro 기능 %q 없음", f)
		}
	}
}

func TestPlanFeatures_EnterpriseHasCloudBackup(t *testing.T) {
	for _, f := range PlanFeatures[PlanEnterprise] {
		if f == FeatureCloudBackup {
			return
		}
	}
	t.Error("PlanEnterprise에 FeatureCloudBackup이 없음")
}

// ─── 2. Plan 포함 관계 — Manager.HasFeature ────────────────────────────────
//
// Manager는 싱글턴이므로 테스트용 헬퍼 함수로 Manager를 직접 생성해 테스트한다.

func newManagerWithLicense(info *LicenseInfo) *Manager {
	m := &Manager{}
	m.license = info
	return m
}

func TestHasFeature_FreePlanBlocksOBSControl(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanFree,
		ExpiresAt: time.Time{}, // 영구
	})
	if m.HasFeature(FeatureOBSControl) {
		t.Error("Free 플랜에서 FeatureOBSControl이 허용됨 — 차단되어야 함")
	}
}

func TestHasFeature_ProPlanAllowsOBSControl(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanPro,
		ExpiresAt: time.Time{}, // 영구
	})
	if !m.HasFeature(FeatureOBSControl) {
		t.Error("Pro 플랜에서 FeatureOBSControl이 차단됨 — 허용되어야 함")
	}
}

func TestHasFeature_ProPlanBlocksCloudBackup(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanPro,
		ExpiresAt: time.Time{},
	})
	if m.HasFeature(FeatureCloudBackup) {
		t.Error("Pro 플랜에서 FeatureCloudBackup이 허용됨 — Enterprise 전용이어야 함")
	}
}

func TestHasFeature_EnterprisePlanAllowsCloudBackup(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanEnterprise,
		ExpiresAt: time.Time{},
	})
	if !m.HasFeature(FeatureCloudBackup) {
		t.Error("Enterprise 플랜에서 FeatureCloudBackup이 차단됨 — 허용되어야 함")
	}
}

func TestHasFeature_NilManagerReturnsFalse(t *testing.T) {
	var m *Manager
	if m.HasFeature(FeatureOBSControl) {
		t.Error("nil Manager에서 HasFeature가 true를 반환함")
	}
}

// ─── 3. 만료 계산 테스트 ────────────────────────────────────────────────────

func TestIsExpired_FutureExpiryNotExpired(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanPro,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 내일 만료
	})
	if m.IsExpired() {
		t.Error("만료일이 미래인데 IsExpired()가 true를 반환함")
	}
}

func TestIsExpired_PastExpiryIsExpired(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanPro,
		ExpiresAt: time.Now().Add(-24 * time.Hour), // 어제 만료
	})
	if !m.IsExpired() {
		t.Error("만료일이 과거인데 IsExpired()가 false를 반환함")
	}
}

func TestIsExpired_ZeroExpiryNeverExpires(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanPro,
		ExpiresAt: time.Time{}, // 영구 라이선스
	})
	if m.IsExpired() {
		t.Error("만료일이 zero(영구)인데 IsExpired()가 true를 반환함")
	}
}

func TestIsInGracePeriod_ExpiredWithinGracePeriod(t *testing.T) {
	// 만료: 어제 / LastVerified: 오늘 → 30일 grace period 이내
	m := newManagerWithLicense(&LicenseInfo{
		Plan:         PlanPro,
		ExpiresAt:    time.Now().Add(-24 * time.Hour),
		LastVerified: time.Now(),
	})
	if !m.IsInGracePeriod() {
		t.Error("만료 후 LastVerified가 최근인데 IsInGracePeriod()가 false를 반환함")
	}
}

func TestIsInGracePeriod_ExpiredBeyondGracePeriod(t *testing.T) {
	// 만료: 40일 전 / LastVerified: 40일 전 → 30일 grace period 초과
	m := newManagerWithLicense(&LicenseInfo{
		Plan:         PlanPro,
		ExpiresAt:    time.Now().AddDate(0, 0, -40),
		LastVerified: time.Now().AddDate(0, 0, -40),
	})
	if m.IsInGracePeriod() {
		t.Error("grace period 초과됐는데 IsInGracePeriod()가 true를 반환함")
	}
}

func TestIsInGracePeriod_NotExpiredReturnsFalse(t *testing.T) {
	// 만료 전이면 grace period 불필요
	m := newManagerWithLicense(&LicenseInfo{
		Plan:         PlanPro,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastVerified: time.Now(),
	})
	if m.IsInGracePeriod() {
		t.Error("만료 전인데 IsInGracePeriod()가 true를 반환함")
	}
}

// ─── 4. HasFeature — 만료 + grace period 상호작용 ─────────────────────────

func TestHasFeature_ExpiredOutsideGracePeriodBlocksFeature(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:         PlanPro,
		ExpiresAt:    time.Now().AddDate(0, 0, -40),
		LastVerified: time.Now().AddDate(0, 0, -40),
	})
	if m.HasFeature(FeatureOBSControl) {
		t.Error("grace period 초과 만료 후에도 HasFeature가 true를 반환함")
	}
}

func TestHasFeature_ExpiredWithinGracePeriodAllowsFeature(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:         PlanPro,
		ExpiresAt:    time.Now().Add(-24 * time.Hour),
		LastVerified: time.Now(),
	})
	if !m.HasFeature(FeatureOBSControl) {
		t.Error("grace period 이내 만료 라이선스인데 HasFeature가 false를 반환함")
	}
}

// ─── 5. DaysUntilExpiry 테스트 ─────────────────────────────────────────────

func TestDaysUntilExpiry_FutureDateReturnsDays(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanPro,
		ExpiresAt: time.Now().Add(7*24*time.Hour + 30*time.Minute), // 약 7일 후
	})
	days := m.DaysUntilExpiry()
	// 시간 계산 오차를 고려해 6~7일 허용
	if days < 6 || days > 7 {
		t.Errorf("만료까지 7일 남았는데 DaysUntilExpiry() = %d", days)
	}
}

func TestDaysUntilExpiry_PastDateReturnsNegative(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanPro,
		ExpiresAt: time.Now().Add(-48 * time.Hour), // 2일 전 만료
	})
	days := m.DaysUntilExpiry()
	if days >= 0 {
		t.Errorf("이미 만료됐는데 DaysUntilExpiry() = %d (음수여야 함)", days)
	}
}

func TestDaysUntilExpiry_ZeroExpiryReturnsMinus1(t *testing.T) {
	m := newManagerWithLicense(&LicenseInfo{
		Plan:      PlanPro,
		ExpiresAt: time.Time{}, // 영구
	})
	if m.DaysUntilExpiry() != -1 {
		t.Errorf("영구 라이선스에서 DaysUntilExpiry() = %d (-1이어야 함)", m.DaysUntilExpiry())
	}
}

func TestDaysUntilExpiry_NilManagerReturns0(t *testing.T) {
	var m *Manager
	if m.DaysUntilExpiry() != 0 {
		t.Errorf("nil Manager에서 DaysUntilExpiry() = %d (0이어야 함)", m.DaysUntilExpiry())
	}
}

// ─── 6. GetPlan 테스트 ────────────────────────────────────────────────────

func TestGetPlan_ReturnsCorrectPlan(t *testing.T) {
	cases := []Plan{PlanFree, PlanPro, PlanEnterprise}
	for _, plan := range cases {
		m := newManagerWithLicense(&LicenseInfo{Plan: plan})
		if got := m.GetPlan(); got != plan {
			t.Errorf("GetPlan() = %q, want %q", got, plan)
		}
	}
}

func TestGetPlan_NilLicenseReturnsFree(t *testing.T) {
	m := &Manager{} // license == nil
	if got := m.GetPlan(); got != PlanFree {
		t.Errorf("라이선스 없을 때 GetPlan() = %q, want %q", got, PlanFree)
	}
}

func TestGetPlan_NilManagerReturnsFree(t *testing.T) {
	var m *Manager
	if got := m.GetPlan(); got != PlanFree {
		t.Errorf("nil Manager에서 GetPlan() = %q, want %q", got, PlanFree)
	}
}

// ─── 7. ValidateKeyFormat 테스트 ──────────────────────────────────────────

func TestValidateKeyFormat_ValidKey(t *testing.T) {
	valid := []string{
		"EP-ABCD-1234-EF56-7890",
		"EP-0000-0000-0000-0000",
		"EP-ZZZZ-ZZZZ-ZZZZ-ZZZZ",
	}
	for _, key := range valid {
		if !ValidateKeyFormat(key) {
			t.Errorf("ValidateKeyFormat(%q) = false, 유효한 키여야 함", key)
		}
	}
}

func TestValidateKeyFormat_InvalidKey(t *testing.T) {
	invalid := []string{
		"",
		"EP-ABCD-1234-EF56",          // 그룹 3개
		"EP-abcd-1234-ef56-7890",      // 소문자
		"XX-ABCD-1234-EF56-7890",      // 잘못된 접두사
		"EP-ABCDE-1234-EF56-7890",     // 그룹 5자
		"EP-ABCD-1234-EF56-7890-ZZZZ", // 그룹 5개
	}
	for _, key := range invalid {
		if ValidateKeyFormat(key) {
			t.Errorf("ValidateKeyFormat(%q) = true, 유효하지 않은 키여야 함", key)
		}
	}
}

// ─── 8. GenerateTestKey + ValidateSignature 테스트 ────────────────────────

func TestGenerateTestKey_ProducesValidFormat(t *testing.T) {
	key, _ := GenerateTestKey(PlanPro, time.Time{})
	if !ValidateKeyFormat(key) {
		t.Errorf("GenerateTestKey가 잘못된 형식의 키를 반환함: %q", key)
	}
}

func TestValidateSignature_ValidSignature(t *testing.T) {
	// GenerateTestKey가 반환한 key+sig 쌍을 그대로 signLicense로 재서명해 비교
	// (IssuedAt을 time.Now()로 설정하므로 동일한 secret으로 재서명해야 일치)
	key, _ := GenerateTestKey(PlanPro, time.Time{})
	issuedAt := time.Now()
	info := &LicenseInfo{
		LicenseKey: key,
		Plan:       PlanPro,
		DeviceID:   "",
		ChurchID:   1,
		IssuedAt:   issuedAt,
		ExpiresAt:  time.Time{},
	}
	info.Signature = signLicense(info, defaultHMACSecret)
	if !ValidateSignature(info, defaultHMACSecret) {
		t.Error("ValidateSignature가 유효한 서명을 거부함")
	}
}

func TestValidateSignature_TamperedPlanFails(t *testing.T) {
	key, sig := GenerateTestKey(PlanFree, time.Time{})
	info := &LicenseInfo{
		LicenseKey: key,
		Plan:       PlanPro, // 플랜 변조
		Signature:  sig,
	}
	if ValidateSignature(info, defaultHMACSecret) {
		t.Error("변조된 플랜인데 ValidateSignature가 true를 반환함")
	}
}

func TestValidateSignature_EmptySignatureFails(t *testing.T) {
	info := &LicenseInfo{
		LicenseKey: "EP-ABCD-1234-EF56-7890",
		Plan:       PlanPro,
		Signature:  "",
	}
	if ValidateSignature(info, defaultHMACSecret) {
		t.Error("서명이 비어 있는데 ValidateSignature가 true를 반환함")
	}
}

func TestValidateSignature_NilInfoFails(t *testing.T) {
	if ValidateSignature(nil, defaultHMACSecret) {
		t.Error("nil LicenseInfo인데 ValidateSignature가 true를 반환함")
	}
}
