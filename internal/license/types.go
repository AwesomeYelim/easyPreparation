package license

import "time"

type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanEnterprise Plan = "enterprise"
)

type Feature string

const (
	FeatureOBSControl    Feature = "obs_control"
	FeatureAutoScheduler Feature = "auto_scheduler"
	FeatureYouTube       Feature = "youtube_integration"
	FeatureThumbnail     Feature = "thumbnail"
	FeatureMultiWorship  Feature = "multi_worship"
	FeatureCloudBackup   Feature = "cloud_backup"
)

// PlanFeatures — each plan's available features
var PlanFeatures = map[Plan][]Feature{
	PlanFree: {},
	PlanPro: {
		FeatureOBSControl, FeatureAutoScheduler,
		FeatureYouTube, FeatureThumbnail, FeatureMultiWorship,
	},
	PlanEnterprise: {
		FeatureOBSControl, FeatureAutoScheduler,
		FeatureYouTube, FeatureThumbnail, FeatureMultiWorship,
		FeatureCloudBackup,
	},
}

type LicenseInfo struct {
	LicenseKey   string    `json:"license_key"`
	Plan         Plan      `json:"plan"`
	DeviceID     string    `json:"device_id"`
	ChurchID     int       `json:"church_id"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastVerified time.Time `json:"last_verified"`
	Signature    string    `json:"signature"`
}
