package handlers

import (
	"easyPreparation_1.0/internal/api/global"
	middleware "easyPreparation_1.0/internal/middlerware"
	"encoding/json"
	"fmt"
	"net/http"
)

func SubmitLyricsHandler(dataChan chan global.DataEnvelope) http.Handler {
	return middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		var response map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		BroadcastProgress("Response SearchLyrics", 1, fmt.Sprintf("Response SearchLyrics: %+v", response))

		// 채널로 데이터 전달
		dataChan <- global.DataEnvelope{
			Type:    "submitLyrics",
			Payload: response,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":              "success",
			"received_song_count": 1,
			"searchedSongs":       "adasdasd",
		})
	}))

}
