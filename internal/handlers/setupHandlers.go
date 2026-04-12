package handlers

import (
	"encoding/json"
	"easyPreparation_1.0/internal/quote"
	"net/http"
)

// SetupStatusHandler — GET /api/setup/status
// churches 테이블에서 id=1 조회, 없으면 needsSetup=true
func SetupStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// apiDB가 nil이면 quote 패키지의 공유 DB로 폴백 (Desktop 앱 SQLite)
	db := apiDB
	if db == nil {
		db = quote.GetDB()
	}
	if db == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"needsSetup": true,
			"church":     nil,
		})
		return
	}

	var id int
	var name, englishName, email string
	err := db.QueryRow(`
		SELECT id, name, english_name, email FROM churches WHERE id = 1
	`).Scan(&id, &name, &englishName, &email)

	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"needsSetup": true,
			"church":     nil,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"needsSetup": false,
		"church": map[string]interface{}{
			"id":          id,
			"name":        name,
			"englishName": englishName,
			"email":       email,
		},
	})
}

// SetupHandler — POST /api/setup
// name, englishName 받아서 churches(id=1)에 INSERT OR REPLACE
func SetupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// apiDB가 nil이면 quote 패키지의 공유 DB로 폴백 (Desktop 앱 SQLite)
	db := apiDB
	if db == nil {
		db = quote.GetDB()
	}
	if db == nil {
		http.Error(w, `{"error":"DB not initialized"}`, http.StatusInternalServerError)
		return
	}

	var body struct {
		Name        string `json:"name"`
		EnglishName string `json:"englishName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`
		INSERT OR REPLACE INTO churches (id, name, english_name, email)
		VALUES (1, ?, ?, 'local@localhost')
	`, body.Name, body.EnglishName)

	if err != nil {
		http.Error(w, `{"error":"save failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"church": map[string]interface{}{
			"id":          1,
			"name":        body.Name,
			"englishName": body.EnglishName,
			"email":       "local@localhost",
		},
	})
}
