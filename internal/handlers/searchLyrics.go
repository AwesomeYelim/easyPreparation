package handlers

import (
	middleware "easyPreparation_1.0/internal/middlerware"
	"easyPreparation_1.0/internal/parser"
	"encoding/json"
	"fmt"
	"net/http"
)

type Song struct {
	Title  string `json:"title"`
	Lyrics string `json:"lyrics"`
}

func SearchLyrics() http.Handler {

	return middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		BroadcastProgress("Response SearchLyrics", 1, fmt.Sprintf("Response SearchLyrics: %+v", data))

		// targetInfo 파싱
		var songs []Song
		if rawTargetInfo, err := json.Marshal(data["songs"]); err == nil {
			_ = json.Unmarshal(rawTargetInfo, &songs)
		} else {
			BroadcastProgress("Songs parsing error", -1, fmt.Sprintf("Failed to parse Songs: %s", err))
			return
		}

		var responseLyrics []Song
		for _, song := range songs {
			newSong := &parser.SlideData{}
			if song.Lyrics == "" {
				newSong.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", song.Title, false)
			}
			responseLyrics = append(responseLyrics, Song{Title: song.Title, Lyrics: newSong.Lyrics})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":              "success",
			"received_song_count": len(data["songs"].([]interface{})),
			"searchedSongs":       responseLyrics,
		})
	}))

}
