package handlers

import (
	"encoding/json"
	"easyPreparation_1.0/internal/selfupdate"
	"easyPreparation_1.0/internal/version"
	"net/http"
)

// UpdateCheckHandler — GET /api/update/check
// GitHub Releases API로 최신 버전을 확인하고 현재 버전과 비교합니다.
func UpdateCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	current := version.Get()

	rel, err := selfupdate.CheckLatest()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      false,
			"error":   err.Error(),
			"current": current.Version,
		})
		return
	}

	isNewer := selfupdate.IsNewer(current.Version, rel.TagName)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":        true,
		"current":   current.Version,
		"latest":    rel.TagName,
		"updateUrl": rel.HTMLURL,
		"notes":     rel.Body,
		"hasUpdate": isNewer,
	})
}
