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

// UpdateStatusHandler — GET /api/update/status
// 현재 업데이트 상태(다운로드 진행률 등)를 반환합니다.
func UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	status := selfupdate.GetUpdater().GetStatus()
	json.NewEncoder(w).Encode(status)
}

// UpdateDownloadHandler — POST /api/update/download
// GitHub Releases에서 현재 플랫폼 바이너리를 비동기 다운로드합니다.
func UpdateDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	rel, err := selfupdate.CheckLatest()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "최신 릴리즈 확인 실패: " + err.Error(),
		})
		return
	}

	current := version.Get()
	if !selfupdate.IsNewer(current.Version, rel.TagName) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      false,
			"error":   "이미 최신 버전입니다",
			"version": current.Version,
		})
		return
	}

	if err := selfupdate.GetUpdater().Download(rel); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"version": rel.TagName,
		"message": "다운로드를 시작했습니다",
	})
}

// UpdateApplyHandler — POST /api/update/apply
// 다운로드된 바이너리로 현재 실행 파일을 교체합니다.
// 성공 시 재시작이 필요합니다 (restartRequired: true).
func UpdateApplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if err := selfupdate.GetUpdater().Apply(); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":              true,
		"restartRequired": true,
		"message":         "업데이트 적용 완료. 서버를 재시작해 주세요.",
	})
}

// UpdateCancelHandler — POST /api/update/cancel
// 진행 중인 다운로드를 취소합니다.
func UpdateCancelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	selfupdate.GetUpdater().CancelDownload()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"message": "다운로드 취소 요청 완료",
	})
}
