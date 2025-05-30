package handlers

import (
	"easyPreparation_1.0/internal/api/global"
	middleware "easyPreparation_1.0/internal/middlerware"
	"encoding/json"
	"fmt"
	"net/http"
)

func SubmitHandler(dataChan chan global.DataEnvelope) http.Handler {
	return middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		fmt.Println("Submit Received:", data)

		// 채널로 데이터 전달
		dataChan <- global.DataEnvelope{
			Type:    "submit",
			Payload: data,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"isOk":         1,
			"isProcessing": true,
			"message":      "Data received successfully",
		})
	}))
}
