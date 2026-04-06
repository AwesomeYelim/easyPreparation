package handlers

import (
	"encoding/json"
	"easyPreparation_1.0/internal/path"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// 허용할 예배 타입 (path traversal 방지)
var validWorshipTypes = map[string]bool{
	"main_worship":  true,
	"after_worship": true,
	"wed_worship":   true,
	"fri_worship":   true,
}

// WorshipOrderHandler — GET/PUT /api/worship-order
func WorshipOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getWorshipOrder(w, r)
	case http.MethodPut:
		putWorshipOrder(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /api/worship-order?type=main_worship
func getWorshipOrder(w http.ResponseWriter, r *http.Request) {
	worshipType := r.URL.Query().Get("type")
	if !validWorshipTypes[worshipType] {
		http.Error(w, "Invalid worship type", http.StatusBadRequest)
		return
	}

	execPath := path.ExecutePath("easyPreparation")
	filePath := filepath.Join(execPath, "config", worshipType+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		// 파일 없으면 빈 배열 반환
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// PUT /api/worship-order
func putWorshipOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type  string                   `json:"type"`
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !validWorshipTypes[body.Type] {
		http.Error(w, "Invalid worship type", http.StatusBadRequest)
		return
	}

	execPath := path.ExecutePath("easyPreparation")
	filePath := filepath.Join(execPath, "config", body.Type+".json")

	marshaled, err := json.MarshalIndent(body.Items, "", "  ")
	if err != nil {
		http.Error(w, "JSON marshal error", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(filePath, marshaled, 0644); err != nil {
		http.Error(w, "File write error", http.StatusInternalServerError)
		return
	}

	log.Printf("[worship-order] 저장 완료: %s (%d items)", body.Type, len(body.Items))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// WorshipOrderListHandler — GET /api/worship-order/list
// 사용 가능한 예배 타입 목록 + 각각 파일 존재 여부 반환
func WorshipOrderListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	execPath := path.ExecutePath("easyPreparation")
	result := make(map[string]bool)
	for t := range validWorshipTypes {
		filePath := filepath.Join(execPath, "config", t+".json")
		_, err := os.Stat(filePath)
		result[t] = err == nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// sanitizeWorshipType — 파일명에 안전한 문자만 허용
func sanitizeWorshipType(t string) bool {
	return regexp.MustCompile(`^[a-z_]+$`).MatchString(t)
}
