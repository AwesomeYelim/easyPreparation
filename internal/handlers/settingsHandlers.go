package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// SettingsHandler — GET/PUT /api/settings
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getSettingsHandler(w, r)
	case http.MethodPut:
		putSettingsHandler(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func getSettingsHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, `{"error":"email required"}`, http.StatusBadRequest)
		return
	}

	var churchID int
	err := apiDB.QueryRow("SELECT id FROM churches WHERE email = $1", email).Scan(&churchID)
	if err != nil {
		http.Error(w, `{"error":"church not found"}`, http.StatusNotFound)
		return
	}

	var preferredBibleVersion, fontSize, defaultBpm int
	var theme, displayLayout string

	err = apiDB.QueryRow(`
		SELECT preferred_bible_version, theme, font_size, default_bpm, display_layout
		FROM user_settings WHERE church_id = $1
	`, churchID).Scan(&preferredBibleVersion, &theme, &fontSize, &defaultBpm, &displayLayout)

	if err != nil {
		// 설정이 없으면 기본값 반환
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"preferred_bible_version": 1,
			"theme":                  "light",
			"font_size":              16,
			"default_bpm":            100,
			"display_layout":         "default",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"preferred_bible_version": preferredBibleVersion,
		"theme":                  theme,
		"font_size":              fontSize,
		"default_bpm":            defaultBpm,
		"display_layout":         displayLayout,
	})
}

func putSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email                 string `json:"email"`
		PreferredBibleVersion int    `json:"preferred_bible_version"`
		Theme                 string `json:"theme"`
		FontSize              int    `json:"font_size"`
		DefaultBpm            int    `json:"default_bpm"`
		DisplayLayout         string `json:"display_layout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	var churchID int
	err := apiDB.QueryRow("SELECT id FROM churches WHERE email = $1", body.Email).Scan(&churchID)
	if err != nil {
		http.Error(w, `{"error":"church not found"}`, http.StatusNotFound)
		return
	}

	// 기본값 보정
	if body.PreferredBibleVersion <= 0 {
		body.PreferredBibleVersion = 1
	}
	if body.Theme == "" {
		body.Theme = "light"
	}
	if body.FontSize <= 0 {
		body.FontSize = 16
	}
	if body.DefaultBpm <= 0 {
		body.DefaultBpm = 100
	}
	if body.DisplayLayout == "" {
		body.DisplayLayout = "default"
	}

	_, err = apiDB.Exec(`
		INSERT INTO user_settings (church_id, preferred_bible_version, theme, font_size, default_bpm, display_layout, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (church_id) DO UPDATE SET
			preferred_bible_version = EXCLUDED.preferred_bible_version,
			theme = EXCLUDED.theme,
			font_size = EXCLUDED.font_size,
			default_bpm = EXCLUDED.default_bpm,
			display_layout = EXCLUDED.display_layout,
			updated_at = NOW()
	`, churchID, body.PreferredBibleVersion, body.Theme, body.FontSize, body.DefaultBpm, body.DisplayLayout)

	if err != nil {
		http.Error(w, `{"error":"save failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HistoryHandler — GET /api/history?email=&type=&page=
func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, `{"error":"email required"}`, http.StatusBadRequest)
		return
	}

	var churchID int
	err := apiDB.QueryRow("SELECT id FROM churches WHERE email = $1", email).Scan(&churchID)
	if err != nil {
		http.Error(w, `{"error":"church not found"}`, http.StatusNotFound)
		return
	}

	genType := r.URL.Query().Get("type")
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	offset := (page - 1) * limit

	var query string
	var args []interface{}

	if genType != "" {
		query = `SELECT id, type, filename, status, metadata, created_at
				 FROM generation_history WHERE church_id = $1 AND type = $2
				 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		args = []interface{}{churchID, genType, limit, offset}
	} else {
		query = `SELECT id, type, filename, status, metadata, created_at
				 FROM generation_history WHERE church_id = $1
				 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{churchID, limit, offset}
	}

	rows, err := apiDB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"query failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int
		var genTypeVal, status string
		var filename *string
		var metadata *string
		var createdAt string

		if err := rows.Scan(&id, &genTypeVal, &filename, &status, &metadata, &createdAt); err != nil {
			continue
		}
		item := map[string]interface{}{
			"id":         id,
			"type":       genTypeVal,
			"status":     status,
			"created_at": createdAt,
		}
		if filename != nil {
			item["filename"] = *filename
		}
		if metadata != nil {
			item["metadata"] = json.RawMessage(*metadata)
		}
		items = append(items, item)
	}

	if items == nil {
		items = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
		"page":  page,
	})
}

// LicenseHandler — PUT /api/settings/license
func LicenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Email      string `json:"email"`
		LicenseKey string `json:"license_key"`
		Token      string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	var churchID int
	err := apiDB.QueryRow("SELECT id FROM churches WHERE email = $1", body.Email).Scan(&churchID)
	if err != nil {
		http.Error(w, `{"error":"church not found"}`, http.StatusNotFound)
		return
	}

	_, err = apiDB.Exec(`
		INSERT INTO licenses (church_id, license_key, license_token, issued_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (church_id) DO UPDATE SET
			license_key = EXCLUDED.license_key,
			license_token = EXCLUDED.license_token,
			issued_at = NOW()
	`, churchID, body.LicenseKey, body.Token)

	if err != nil {
		http.Error(w, `{"error":"save failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// RecordGeneration — 생성 이력 기록 (내부 호출용)
func RecordGeneration(churchEmail, genType, filename, filePath, status string) {
	if apiDB == nil {
		return
	}
	var churchID int
	err := apiDB.QueryRow("SELECT id FROM churches WHERE email = $1", churchEmail).Scan(&churchID)
	if err != nil {
		return
	}
	_, _ = apiDB.Exec(`
		INSERT INTO generation_history (church_id, type, filename, status, file_path, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, churchID, genType, filename, status, filePath)
}
