package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CORS 설정을 위한 함수
func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // 모든 도메인에서 허용
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	enableCors(w) // CORS 적용

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	fmt.Println("Received Data:", requestData)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data received successfully"})
}

func main() {
	http.HandleFunc("/submit", handleSubmit)
	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
