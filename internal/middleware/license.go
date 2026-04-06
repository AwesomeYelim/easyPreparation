package middleware

import (
	"easyPreparation_1.0/internal/license"
	"encoding/json"
	"net/http"
)

// RequireFeature — blocks requests if the feature is not available in the current plan.
func RequireFeature(feature license.Feature) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			mgr := license.Get()
			if mgr != nil && mgr.HasFeature(feature) {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			plan := license.PlanFree
			if mgr != nil {
				plan = mgr.GetPlan()
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "feature_locked",
				"feature": string(feature),
				"plan":    string(plan),
				"message": "이 기능은 Pro 플랜에서 사용할 수 있습니다.",
			})
		})
	}
}

// FeatureGate — CORS + RequireFeature 조합 핸들러
func FeatureGate(feature license.Feature, handler http.HandlerFunc) http.Handler {
	return CORS(RequireFeature(feature)(http.HandlerFunc(handler)))
}
