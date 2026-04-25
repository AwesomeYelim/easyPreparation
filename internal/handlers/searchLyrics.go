package handlers

import (
	middleware "easyPreparation_1.0/internal/middleware"
	"easyPreparation_1.0/internal/parser"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
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
			lyrics := ""

			if song.Lyrics != "" {
				// 이미 가사가 있으면 그대로 사용
				lyrics = song.Lyrics
			} else {
				// 1단계: custom_songs 테이블에서 먼저 검색
				lyrics = searchCustomSongs(song.Title)

				// 2단계: custom_songs 히트 없으면 찬송가 DB 검색
				if lyrics == "" {
					lyrics = searchHymnDB(song.Title)
				}

				// 3단계: 찬송가 DB에도 없으면 bugs.co.kr 크롤링
				if lyrics == "" {
					newSong := &parser.SlideData{}
					if err := newSong.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", song.Title, false); err != nil {
						BroadcastProgress("SearchLyrics Error", -1, fmt.Sprintf("가사 검색 실패 (%s): %v", song.Title, err))
					}
					lyrics = newSong.Lyrics
				}
			}

			responseLyrics = append(responseLyrics, Song{Title: song.Title, Lyrics: lyrics})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":              "success",
			"received_song_count": len(data["songs"].([]interface{})),
			"searchedSongs":       responseLyrics,
		})
	}))

}

// searchCustomSongs — apiDB의 custom_songs 테이블에서 제목으로 검색.
// 히트 시 used_count 증가 + last_used 갱신 후 가사를 반환한다.
// 결과 없으면 빈 문자열 반환.
func searchCustomSongs(title string) string {
	if apiDB == nil {
		return ""
	}

	titleTrim := strings.TrimSpace(title)
	if titleTrim == "" {
		return ""
	}

	var id int
	var lyrics string
	err := apiDB.QueryRow(`
		SELECT id, lyrics FROM custom_songs
		WHERE title LIKE '%' || ? || '%'
		ORDER BY used_count DESC
		LIMIT 1
	`, titleTrim).Scan(&id, &lyrics)
	if err != nil {
		// 결과 없음 또는 오류 — 다음 단계로 넘어감
		return ""
	}
	if lyrics == "" {
		return ""
	}

	// used_count 증가 + last_used 갱신 (실패해도 검색 결과는 반환)
	_, _ = apiDB.Exec(`
		UPDATE custom_songs SET used_count = used_count + 1, last_used = ? WHERE id = ?
	`, time.Now().UTC().Format("2006-01-02T15:04:05Z"), id)

	return lyrics
}

// searchHymnDB — bibleDB의 hymns 테이블에서 제목/첫줄로 검색.
// 결과 없으면 빈 문자열 반환.
func searchHymnDB(title string) string {
	if bibleDB == nil {
		return ""
	}

	titleTrim := strings.TrimSpace(title)
	if titleTrim == "" {
		return ""
	}

	var lyrics *string
	err := bibleDB.QueryRow(`
		SELECT lyrics FROM hymns
		WHERE title LIKE '%' || ? || '%' OR first_line LIKE '%' || ? || '%'
		ORDER BY number
		LIMIT 1
	`, titleTrim, titleTrim).Scan(&lyrics)
	if err != nil || lyrics == nil || *lyrics == "" {
		return ""
	}

	return *lyrics
}
