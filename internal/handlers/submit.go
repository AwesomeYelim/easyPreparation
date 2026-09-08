package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"easyPreparation_1.0/internal/httpx"
	middleware "easyPreparation_1.0/internal/middleware"
	"easyPreparation_1.0/internal/types"
)

func SubmitHandler(dataChan chan types.DataEnvelope) http.Handler {
	return middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			httpx.Error(w, http.StatusMethodNotAllowed, "Invalid request method")
			return
		}

		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		fmt.Println("Submit Received:", data)

		// 채널로 데이터 전달
		dataChan <- types.DataEnvelope{
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
