package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	middleware "easyPreparation_1.0/internal/middlerware"
)

var DataChan = make(chan map[string]interface{}, 100)

func StartServer() {
	mux := http.NewServeMux()

	mux.Handle("/submit", middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		fmt.Println("📩 Submit Received:", data)

		// 채널로 데이터 전달
		DataChan <- data

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isOk":         1,
			"isProcessing": true,
			"message":      "Data received successfully",
		})
	})))

	fmt.Println("🌐 Server running on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		panic(err)
	}
}
