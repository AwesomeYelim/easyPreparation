package license

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// hmacSecret — 개발/테스트용 서명 시크릿 (프로덕션에서는 환경변수로 오버라이드)
const hmacSecret = "easyPrep-license-secret-2024"

// keyPattern — 라이선스 키 형식: EP-XXXX-XXXX-XXXX-XXXX (X: 대문자 알파벳 or 숫자)
var keyPattern = regexp.MustCompile(`^EP-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`)

// ValidateKeyFormat — 라이선스 키 형식 검증
func ValidateKeyFormat(key string) bool {
	return keyPattern.MatchString(key)
}

// GenerateTestKey — 개발 테스트용 라이선스 키 + 서명 생성
// 반환: key (EP-XXXX-XXXX-XXXX-XXXX 형식), signature (HMAC-SHA256 hex)
func GenerateTestKey(plan Plan, expiresAt time.Time) (key string, signature string) {
	// 랜덤 16바이트 생성 → 32자 hex → 그룹화
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		// fallback: 시간 기반
		now := time.Now().UnixNano()
		raw = []byte(fmt.Sprintf("%016x", now))[:8]
	}

	hexStr := strings.ToUpper(hex.EncodeToString(raw)) // 16자
	key = fmt.Sprintf("EP-%s-%s-%s-%s",
		hexStr[0:4],
		hexStr[4:8],
		hexStr[8:12],
		hexStr[12:16],
	)

	// 서명 생성용 임시 LicenseInfo
	info := &LicenseInfo{
		LicenseKey: key,
		Plan:       plan,
		DeviceID:   "",
		ChurchID:   1,
		IssuedAt:   time.Now(),
		ExpiresAt:  expiresAt,
	}

	signature = signLicense(info, hmacSecret)
	return key, signature
}

// ValidateSignature — LicenseInfo의 서명 검증
func ValidateSignature(info *LicenseInfo, secret string) bool {
	if info == nil || info.Signature == "" {
		return false
	}
	if secret == "" {
		secret = hmacSecret
	}
	expected := signLicense(info, secret)
	// hmac.Equal은 타이밍 공격 방지를 위한 상수 시간 비교
	expectedBytes, _ := hex.DecodeString(expected)
	actualBytes, err := hex.DecodeString(info.Signature)
	if err != nil {
		return false
	}
	return hmac.Equal(expectedBytes, actualBytes)
}

// signLicense — LicenseInfo 필드들을 조합해 HMAC-SHA256 서명 생성
// 서명 대상: licenseKey|plan|deviceID|churchID|issuedAt|expiresAt
func signLicense(info *LicenseInfo, secret string) string {
	expiresStr := ""
	if !info.ExpiresAt.IsZero() {
		expiresStr = info.ExpiresAt.UTC().Format(time.RFC3339)
	}

	payload := fmt.Sprintf("%s|%s|%s|%d|%s|%s",
		info.LicenseKey,
		string(info.Plan),
		info.DeviceID,
		info.ChurchID,
		info.IssuedAt.UTC().Format(time.RFC3339),
		expiresStr,
	)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
