package license

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Manager — 라이선스 싱글턴 매니저
type Manager struct {
	mu        sync.RWMutex
	license   *LicenseInfo
	deviceID  string
	cachePath string
	db        *sql.DB
}

var (
	instance *Manager
	once     sync.Once
)

// Init — DB 연결 수신, 디바이스 ID 생성, DB/파일 캐시에서 라이선스 로드
func Init(db *sql.DB) {
	once.Do(func() {
		m := &Manager{
			db:        db,
			cachePath: resolveCachePath(),
		}

		m.deviceID = generateDeviceID()

		// DB에서 라이선스 로드 시도
		if db != nil {
			if info, err := m.loadFromDB(); err == nil && info != nil {
				m.license = info
				log.Printf("[license] DB에서 라이선스 로드 완료 (plan=%s, expires=%s)", info.Plan, info.ExpiresAt.Format("2006-01-02"))
			} else {
				if err != nil && err != sql.ErrNoRows {
					log.Printf("[license] DB 라이선스 조회 오류: %v", err)
				}
				// DB에 없으면 파일 캐시 시도
				if cached, err := loadCache(); err == nil && cached != nil {
					m.license = cached
					log.Printf("[license] 파일 캐시에서 라이선스 로드 완료 (plan=%s)", cached.Plan)
				} else {
					log.Printf("[license] 라이선스 없음 — 무료 플랜으로 시작")
				}
			}
		} else {
			// DB 없음 — 파일 캐시만 시도
			if cached, err := loadCache(); err == nil && cached != nil {
				m.license = cached
				log.Printf("[license] 파일 캐시에서 라이선스 로드 완료 (plan=%s)", cached.Plan)
			}
		}

		// 개발모드: 기존 플랜에 관계없이 항상 Pro로 시작
		if os.Getenv("EASYPREP_DEV") == "true" && (m.license == nil || (m.license.Plan != PlanPro && m.license.Plan != PlanEnterprise)) {
			key, sig := GenerateTestKey(PlanPro, time.Time{})
			m.license = &LicenseInfo{
				LicenseKey:   key,
				Plan:         PlanPro,
				DeviceID:     m.deviceID,
				ChurchID:     1,
				IssuedAt:     time.Now(),
				ExpiresAt:    time.Time{},
				LastVerified: time.Now(),
				Signature:    sig,
			}
			_ = saveCache(m.license)
			log.Printf("[license] 개발모드 — Pro 자동 적용")
		}

		instance = m
	})
}

// Get — 싱글턴 반환 (nil-safe)
func Get() *Manager {
	return instance
}

// GetPlan — 현재 플랜 반환 (기본: PlanFree)
func (m *Manager) GetPlan() Plan {
	if m == nil {
		return PlanFree
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.license == nil {
		return PlanFree
	}
	return m.license.Plan
}

// HasFeature — 현재 플랜이 해당 기능을 포함하고 만료되지 않았는지 확인
func (m *Manager) HasFeature(f Feature) bool {
	if m == nil {
		return false
	}
	// 만료된 라이선스는 grace period 이내가 아니면 기능 차단
	if m.IsExpired() && !m.IsInGracePeriod() {
		return false
	}
	plan := m.GetPlan()
	features, ok := PlanFeatures[plan]
	if !ok {
		return false
	}
	for _, feat := range features {
		if feat == f {
			return true
		}
	}
	return false
}

// GetLicense — 현재 라이선스 정보 복사본 반환 (nil 가능)
func (m *Manager) GetLicense() *LicenseInfo {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.license == nil {
		return nil
	}
	copy := *m.license
	return &copy
}

// IsExpired — 라이선스 만료 여부 확인
func (m *Manager) IsExpired() bool {
	if m == nil {
		return true
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.license == nil {
		return true
	}
	if m.license.ExpiresAt.IsZero() {
		return false // 만료일 없으면 영구 라이선스
	}
	return time.Now().After(m.license.ExpiresAt)
}

// IsInGracePeriod — 만료됐지만 마지막 인증으로부터 GracePeriodDays 이내인지 확인
func (m *Manager) IsInGracePeriod() bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.license == nil {
		return false
	}
	// 만료되지 않았으면 grace period 불필요
	if !time.Now().After(m.license.ExpiresAt) {
		return false
	}
	if m.license.LastVerified.IsZero() {
		return false
	}
	graceCutoff := m.license.LastVerified.AddDate(0, 0, GracePeriodDays)
	return time.Now().Before(graceCutoff)
}

// DaysUntilExpiry — 만료까지 남은 일수 반환 (이미 만료시 음수, 만료일 없으면 -1)
func (m *Manager) DaysUntilExpiry() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.license == nil {
		return 0
	}
	if m.license.ExpiresAt.IsZero() {
		return -1 // 영구 라이선스
	}
	remaining := time.Until(m.license.ExpiresAt)
	return int(remaining.Hours() / 24)
}

// GetDeviceID — 현재 디바이스 ID 반환
func (m *Manager) GetDeviceID() string {
	if m == nil {
		return ""
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.deviceID
}

// SetLicense — 라이선스 정보 업데이트 + DB 저장 + 파일 캐시 저장
func (m *Manager) SetLicense(info *LicenseInfo) error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	m.license = info
	m.mu.Unlock()

	// DB 저장
	if m.db != nil {
		if err := m.saveToDB(info); err != nil {
			log.Printf("[license] DB 저장 실패: %v", err)
		}
	}

	// 파일 캐시 저장
	if err := saveCache(info); err != nil {
		log.Printf("[license] 파일 캐시 저장 실패: %v", err)
	}

	log.Printf("[license] 라이선스 업데이트 완료 (plan=%s, device=%s)", info.Plan, info.DeviceID)
	return nil
}

// UpdateVerification — 서버 검증 응답으로 플랜/만료일/lastVerified 업데이트
// verifier.go에서 사용
func (m *Manager) UpdateVerification(plan Plan, expiresAt time.Time) error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	if m.license == nil {
		m.mu.Unlock()
		return nil
	}
	m.license.Plan = plan
	m.license.ExpiresAt = expiresAt
	m.license.LastVerified = time.Now()
	info := *m.license
	m.mu.Unlock()

	// DB + 캐시 저장
	if m.db != nil {
		if err := m.saveToDB(&info); err != nil {
			log.Printf("[license] DB 저장 실패: %v", err)
		}
	}
	if err := saveCache(&info); err != nil {
		log.Printf("[license] 파일 캐시 저장 실패: %v", err)
	}
	return nil
}

// loadFromDB — DB의 licenses 테이블에서 church_id=1 기준으로 라이선스 로드
func (m *Manager) loadFromDB() (*LicenseInfo, error) {
	var (
		licenseKey   string
		plan         string
		deviceID     string
		signature    string
		issuedAt     string
		expiresAt    sql.NullString
		lastVerified sql.NullString
	)

	err := m.db.QueryRow(`
		SELECT license_key, COALESCE(plan,'free'), COALESCE(device_id,''),
		       COALESCE(signature,''), issued_at,
		       expires_at, last_verified
		FROM licenses
		WHERE church_id = 1
		LIMIT 1
	`).Scan(&licenseKey, &plan, &deviceID, &signature, &issuedAt, &expiresAt, &lastVerified)
	if err != nil {
		return nil, err
	}
	if licenseKey == "" {
		return nil, sql.ErrNoRows
	}

	info := &LicenseInfo{
		LicenseKey: licenseKey,
		Plan:       Plan(plan),
		DeviceID:   deviceID,
		ChurchID:   1,
		Signature:  signature,
	}

	// issued_at 파싱
	if t, err := parseSQLiteTime(issuedAt); err == nil {
		info.IssuedAt = t
	}

	// expires_at 파싱 (nullable)
	if expiresAt.Valid && expiresAt.String != "" {
		if t, err := parseSQLiteTime(expiresAt.String); err == nil {
			info.ExpiresAt = t
		}
	}

	// last_verified 파싱 (nullable)
	if lastVerified.Valid && lastVerified.String != "" {
		if t, err := parseSQLiteTime(lastVerified.String); err == nil {
			info.LastVerified = t
		}
	}

	return info, nil
}

// saveToDB — licenses 테이블에 라이선스 정보 upsert
func (m *Manager) saveToDB(info *LicenseInfo) error {
	expiresAt := ""
	if !info.ExpiresAt.IsZero() {
		expiresAt = info.ExpiresAt.Format("2006-01-02 15:04:05")
	}
	lastVerified := info.LastVerified.Format("2006-01-02 15:04:05")

	_, err := m.db.Exec(`
		INSERT INTO licenses (church_id, license_key, plan, device_id, signature, expires_at, last_verified, issued_at)
		VALUES (1, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT (church_id) DO UPDATE SET
			license_key   = excluded.license_key,
			plan          = excluded.plan,
			device_id     = excluded.device_id,
			signature     = excluded.signature,
			expires_at    = excluded.expires_at,
			last_verified = excluded.last_verified
	`, info.LicenseKey, string(info.Plan), info.DeviceID, info.Signature, nullableStr(expiresAt), lastVerified)
	return err
}

// nullableStr — 빈 문자열이면 nil 반환 (SQL NULL 저장용)
func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// parseSQLiteTime — SQLite datetime 문자열 파싱 (여러 형식 대응)
func parseSQLiteTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("알 수 없는 시간 형식: %s", s)
}
