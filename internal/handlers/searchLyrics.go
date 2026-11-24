package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"easyPreparation_1.0/internal/httpx"
	middleware "easyPreparation_1.0/internal/middleware"
	"easyPreparation_1.0/internal/parser"
)

type Song struct {
	Title  string `json:"title"`
	Lyrics string `json:"lyrics"`
	Source string `json:"source,omitempty"` // input(입력한 가사) | saved(저장본) | hymn(찬송가) | gasazip, 못 찾으면 빈 값
	Error  string `json:"error,omitempty"`  // 가사집 접속 실패 사유
}

func SearchLyrics() http.Handler {

	return middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			httpx.Error(w, http.StatusMethodNotAllowed, "Invalid method")
			return
		}

		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid JSON")
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

		// 응답은 요청과 같은 순서로 돌려준다 (화면은 순서로 맞춘다).
		var responseLyrics []Song
		for _, song := range songs {
			res := Song{Title: song.Title, Lyrics: song.Lyrics}
			if song.Lyrics != "" {
				res.Source = "input"
			} else if lyrics := searchCustomSongs(song.Title); lyrics != "" {
				res.Lyrics, res.Source = lyrics, "saved"
			} else if lyrics := searchHymnDB(song.Title); lyrics != "" {
				res.Lyrics, res.Source = lyrics, "hymn"
			} else {
				newSong := &parser.SlideData{}
				if err := newSong.SearchLyricsList(song.Title); err != nil {
					res.Error = err.Error()
					BroadcastProgress("SearchLyrics Error", -1, fmt.Sprintf("가사 검색 실패 (%s): %v", song.Title, err))
				} else if newSong.Lyrics != "" {
					res.Lyrics, res.Source = newSong.Lyrics, "gasazip"
				}
			}
			responseLyrics = append(responseLyrics, res)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":              "success",
			"received_song_count": len(data["songs"].([]interface{})),
			"searchedSongs":       responseLyrics,
		})
	}))

}

// searchCustomSongs — apiDB의 custom_songs 테이블에서 같은 곡을 찾는다. 띄어쓰기·대소문자는 무시한다.
// 가수를 적었으면("제목 - 가수") 제목·가수가 모두 같은 곡만, 안 적었으면 제목이 같은 곡을 쓴다.
// 저장된 가수는 제목의 " - 가수"(가사 화면에서 저장) 또는 artist 칸(곡 관리 화면에서 등록)에서 읽는다.
// 제목 일부만 맞는 곡은 쓰지 않는다 — 짧은 제목이 다른 곡에 걸려 틀린 가사가 굳는 일을 막기 위함.
// 히트 시 used_count 증가 + last_used 갱신 후 가사를 반환한다. 결과 없으면 빈 문자열.
func searchCustomSongs(query string) string {
	if apiDB == nil {
		return ""
	}
	title, artist := parser.SplitQuery(query)
	want := parser.Squash(title)
	if want == "" {
		return ""
	}
	wantArtist := parser.Squash(artist)

	rows, err := apiDB.Query(`
		SELECT id, title, COALESCE(artist, ''), lyrics FROM custom_songs
		WHERE REPLACE(LOWER(title), ' ', '') LIKE '%' || ? || '%'
		ORDER BY used_count DESC
	`, want)
	if err != nil {
		return ""
	}

	var id int
	var lyrics string
	for rows.Next() {
		var rowID int
		var rowTitle, rowArtistCol, rowLyrics string
		if rows.Scan(&rowID, &rowTitle, &rowArtistCol, &rowLyrics) != nil || strings.TrimSpace(rowLyrics) == "" {
			continue
		}
		rowSong, rowArtist := parser.SplitQuery(rowTitle)
		if rowArtist == "" {
			rowArtist = rowArtistCol
		}
		if parser.Squash(rowSong) == want && (wantArtist == "" || parser.Squash(rowArtist) == wantArtist) {
			id, lyrics = rowID, rowLyrics
			break
		}
	}
	rows.Close() // 아래 UPDATE 전에 읽기 연결을 놓는다
	if lyrics == "" {
		return ""
	}

	// used_count 증가 + last_used 갱신 (실패해도 검색 결과는 반환)
	_, _ = apiDB.Exec(`
		UPDATE custom_songs SET used_count = used_count + 1, last_used = ? WHERE id = ?
	`, time.Now().UTC().Format("2006-01-02T15:04:05Z"), id)

	return lyrics
}

// saveCustomSongLyrics — 사용자가 Display 로 보낸(=확정한) 가사를 custom_songs 에 저장한다.
// 같은 제목이 있으면 가사만 갱신하고, 없으면 새로 등록한다. 다음 검색부터 이 가사가 먼저 나온다.
func saveCustomSongLyrics(title, lyrics string) {
	title = strings.TrimSpace(title)
	if apiDB == nil || title == "" || strings.TrimSpace(lyrics) == "" {
		return
	}

	res, err := apiDB.Exec(`UPDATE custom_songs SET lyrics = ? WHERE title = ?`, lyrics, title)
	if err != nil {
		log.Printf("[lyrics] failed to update custom song %q: %v", title, err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return
	}
	if _, err := apiDB.Exec(`INSERT INTO custom_songs (title, lyrics) VALUES (?, ?)`, title, lyrics); err != nil {
		log.Printf("[lyrics] failed to insert custom song %q: %v", title, err)
	}
}

// hymnNumberPattern — "305", "305장", "찬송가 305장", "새찬송가 305장 나 같은 죄인" 의 장 번호
var hymnNumberPattern = regexp.MustCompile(`^(?:새\s*)?(?:찬송가?\s*)?(\d{1,3})\s*장?(?:\s|$)`)

// searchHymnDB — bibleDB의 hymns 테이블에서 찬송가를 찾는다.
// 장 번호를 적었으면 번호로, 아니면 제목·첫 줄이 같은 곡(띄어쓰기 무시)만 쓴다.
// 제목 일부만 맞는 곡은 쓰지 않는다 — CCM 제목이 찬송가 제목 일부에 걸리는 일을 막기 위함.
// 결과 없으면 빈 문자열 반환.
func searchHymnDB(query string) string {
	if bibleDB == nil {
		return ""
	}
	query = strings.TrimSpace(query)
	if m := hymnNumberPattern.FindStringSubmatch(query); m != nil {
		number, _ := strconv.Atoi(m[1])
		return fetchHymnLyrics(number)
	}
	want := parser.Squash(query)
	if want == "" {
		return ""
	}

	var lyrics *string
	err := bibleDB.QueryRow(`
		SELECT lyrics FROM hymns
		WHERE REPLACE(title, ' ', '') = ? OR REPLACE(first_line, ' ', '') = ?
		ORDER BY number
		LIMIT 1
	`, want, want).Scan(&lyrics)
	if err != nil || lyrics == nil || *lyrics == "" {
		return ""
	}

	return *lyrics
}
