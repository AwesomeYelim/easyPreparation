package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CustomSong — 커스텀 찬양 곡 구조체
type CustomSong struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Artist    string   `json:"artist"`
	Lyrics    string   `json:"lyrics"`
	Tags      []string `json:"tags"`
	UsedCount int      `json:"used_count"`
	LastUsed  string   `json:"last_used"`
	CreatedAt string   `json:"created_at"`
}

// CustomSongSearchHandler — GET /api/songs/search?q=검색어
// title, artist, lyrics 전문 검색, 결과 최대 20개
func CustomSongSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "허용되지 않는 메서드입니다", http.StatusMethodNotAllowed)
		return
	}
	if apiDB == nil {
		http.Error(w, "데이터베이스에 연결할 수 없습니다", http.StatusServiceUnavailable)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "q 파라미터가 필요합니다", http.StatusBadRequest)
		return
	}

	rows, err := apiDB.Query(`
		SELECT id, title, artist, lyrics, tags, used_count, last_used, created_at
		FROM custom_songs
		WHERE title LIKE '%' || ? || '%'
		   OR artist LIKE '%' || ? || '%'
		   OR lyrics LIKE '%' || ? || '%'
		ORDER BY used_count DESC, title ASC
		LIMIT 20
	`, q, q, q)
	if err != nil {
		http.Error(w, "검색 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	songs := make([]CustomSong, 0)
	for rows.Next() {
		var s CustomSong
		var tagsJSON string
		if err := rows.Scan(&s.ID, &s.Title, &s.Artist, &s.Lyrics, &tagsJSON, &s.UsedCount, &s.LastUsed, &s.CreatedAt); err != nil {
			continue
		}
		_ = json.Unmarshal([]byte(tagsJSON), &s.Tags)
		if s.Tags == nil {
			s.Tags = []string{}
		}
		songs = append(songs, s)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(songs)
}

// CustomSongImportHandler — POST /api/songs/import
// 텍스트 파일 일괄 임포트
func CustomSongImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "허용되지 않는 메서드입니다", http.StatusMethodNotAllowed)
		return
	}
	if apiDB == nil {
		http.Error(w, "데이터베이스에 연결할 수 없습니다", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "요청 형식이 올바르지 않습니다: "+err.Error(), http.StatusBadRequest)
		return
	}

	songs := parseImportText(body.Text)
	if len(songs) == 0 {
		http.Error(w, "임포트할 곡이 없습니다. 형식을 확인해주세요", http.StatusBadRequest)
		return
	}

	imported := 0
	for _, s := range songs {
		tagsJSON, _ := json.Marshal(s.Tags)
		_, err := apiDB.Exec(`
			INSERT INTO custom_songs (title, artist, lyrics, tags)
			VALUES (?, ?, ?, ?)
		`, s.Title, s.Artist, s.Lyrics, string(tagsJSON))
		if err == nil {
			imported++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"imported": imported,
		"total":    len(songs),
	})
}

// CustomSongCRUDHandler — /api/songs/ 경로 분기 처리
func CustomSongCRUDHandler(w http.ResponseWriter, r *http.Request) {
	// /api/songs/{id} 형태에서 id 추출
	path := strings.TrimPrefix(r.URL.Path, "/api/songs/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		// /api/songs/ → POST(생성) 또는 GET(전체 목록)
		switch r.Method {
		case http.MethodGet:
			customSongListHandler(w, r)
		case http.MethodPost:
			customSongCreateHandler(w, r)
		default:
			http.Error(w, "허용되지 않는 메서드입니다", http.StatusMethodNotAllowed)
		}
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "유효하지 않은 곡 ID입니다", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		customSongGetHandler(w, r, id)
	case http.MethodPut:
		customSongUpdateHandler(w, r, id)
	case http.MethodDelete:
		customSongDeleteHandler(w, r, id)
	default:
		http.Error(w, "허용되지 않는 메서드입니다", http.StatusMethodNotAllowed)
	}
}

// customSongListHandler — GET /api/songs/
// 전체 목록 (최근 사용순)
func customSongListHandler(w http.ResponseWriter, r *http.Request) {
	if apiDB == nil {
		http.Error(w, "데이터베이스에 연결할 수 없습니다", http.StatusServiceUnavailable)
		return
	}

	rows, err := apiDB.Query(`
		SELECT id, title, artist, lyrics, tags, used_count, last_used, created_at
		FROM custom_songs
		ORDER BY used_count DESC, created_at DESC
		LIMIT 100
	`)
	if err != nil {
		http.Error(w, "목록 조회 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	songs := make([]CustomSong, 0)
	for rows.Next() {
		var s CustomSong
		var tagsJSON string
		if err := rows.Scan(&s.ID, &s.Title, &s.Artist, &s.Lyrics, &tagsJSON, &s.UsedCount, &s.LastUsed, &s.CreatedAt); err != nil {
			continue
		}
		_ = json.Unmarshal([]byte(tagsJSON), &s.Tags)
		if s.Tags == nil {
			s.Tags = []string{}
		}
		songs = append(songs, s)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(songs)
}

// customSongGetHandler — GET /api/songs/:id
// used_count + last_used 업데이트
func customSongGetHandler(w http.ResponseWriter, r *http.Request, id int) {
	if apiDB == nil {
		http.Error(w, "데이터베이스에 연결할 수 없습니다", http.StatusServiceUnavailable)
		return
	}

	var s CustomSong
	var tagsJSON string
	err := apiDB.QueryRow(`
		SELECT id, title, artist, lyrics, tags, used_count, last_used, created_at
		FROM custom_songs WHERE id = ?
	`, id).Scan(&s.ID, &s.Title, &s.Artist, &s.Lyrics, &tagsJSON, &s.UsedCount, &s.LastUsed, &s.CreatedAt)
	if err != nil {
		http.Error(w, "곡을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	_ = json.Unmarshal([]byte(tagsJSON), &s.Tags)
	if s.Tags == nil {
		s.Tags = []string{}
	}

	// used_count + last_used 업데이트
	today := time.Now().Format("2006-01-02")
	_, _ = apiDB.Exec(`
		UPDATE custom_songs SET used_count = used_count + 1, last_used = ? WHERE id = ?
	`, today, id)
	s.UsedCount++
	s.LastUsed = today

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}

// customSongCreateHandler — POST /api/songs
func customSongCreateHandler(w http.ResponseWriter, r *http.Request) {
	if apiDB == nil {
		http.Error(w, "데이터베이스에 연결할 수 없습니다", http.StatusServiceUnavailable)
		return
	}

	var input struct {
		Title  string   `json:"title"`
		Artist string   `json:"artist"`
		Lyrics string   `json:"lyrics"`
		Tags   []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "요청 형식이 올바르지 않습니다: "+err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Title) == "" {
		http.Error(w, "제목은 필수입니다", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Lyrics) == "" {
		http.Error(w, "가사는 필수입니다", http.StatusBadRequest)
		return
	}
	if input.Tags == nil {
		input.Tags = []string{}
	}

	tagsJSON, _ := json.Marshal(input.Tags)
	result, err := apiDB.Exec(`
		INSERT INTO custom_songs (title, artist, lyrics, tags)
		VALUES (?, ?, ?, ?)
	`, input.Title, input.Artist, input.Lyrics, string(tagsJSON))
	if err != nil {
		http.Error(w, "곡 저장 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}

	newID, _ := result.LastInsertId()

	var s CustomSong
	var tagsJSONOut string
	_ = apiDB.QueryRow(`
		SELECT id, title, artist, lyrics, tags, used_count, last_used, created_at
		FROM custom_songs WHERE id = ?
	`, newID).Scan(&s.ID, &s.Title, &s.Artist, &s.Lyrics, &tagsJSONOut, &s.UsedCount, &s.LastUsed, &s.CreatedAt)
	_ = json.Unmarshal([]byte(tagsJSONOut), &s.Tags)
	if s.Tags == nil {
		s.Tags = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(s)
}

// customSongUpdateHandler — PUT /api/songs/:id
func customSongUpdateHandler(w http.ResponseWriter, r *http.Request, id int) {
	if apiDB == nil {
		http.Error(w, "데이터베이스에 연결할 수 없습니다", http.StatusServiceUnavailable)
		return
	}

	var input struct {
		Title  string   `json:"title"`
		Artist string   `json:"artist"`
		Lyrics string   `json:"lyrics"`
		Tags   []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "요청 형식이 올바르지 않습니다: "+err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Title) == "" {
		http.Error(w, "제목은 필수입니다", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Lyrics) == "" {
		http.Error(w, "가사는 필수입니다", http.StatusBadRequest)
		return
	}
	if input.Tags == nil {
		input.Tags = []string{}
	}

	tagsJSON, _ := json.Marshal(input.Tags)
	result, err := apiDB.Exec(`
		UPDATE custom_songs SET title = ?, artist = ?, lyrics = ?, tags = ?
		WHERE id = ?
	`, input.Title, input.Artist, input.Lyrics, string(tagsJSON), id)
	if err != nil {
		http.Error(w, "곡 수정 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		http.Error(w, "곡을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	var s CustomSong
	var tagsJSONOut string
	_ = apiDB.QueryRow(`
		SELECT id, title, artist, lyrics, tags, used_count, last_used, created_at
		FROM custom_songs WHERE id = ?
	`, id).Scan(&s.ID, &s.Title, &s.Artist, &s.Lyrics, &tagsJSONOut, &s.UsedCount, &s.LastUsed, &s.CreatedAt)
	_ = json.Unmarshal([]byte(tagsJSONOut), &s.Tags)
	if s.Tags == nil {
		s.Tags = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}

// customSongDeleteHandler — DELETE /api/songs/:id
func customSongDeleteHandler(w http.ResponseWriter, r *http.Request, id int) {
	if apiDB == nil {
		http.Error(w, "데이터베이스에 연결할 수 없습니다", http.StatusServiceUnavailable)
		return
	}

	result, err := apiDB.Exec(`DELETE FROM custom_songs WHERE id = ?`, id)
	if err != nil {
		http.Error(w, "곡 삭제 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		http.Error(w, "곡을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"message": "곡이 삭제되었습니다",
	})
}

// parseImportText — 텍스트 형식 파싱 (여러 곡은 === 구분)
// 형식:
//
//	제목: 어메이징 그레이스
//	아티스트: 전통
//	---
//	가사 내용
func parseImportText(text string) []CustomSong {
	var songs []CustomSong

	// === 로 곡 분리
	blocks := strings.Split(text, "===")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		// --- 로 헤더와 가사 분리
		parts := strings.SplitN(block, "---", 2)
		if len(parts) < 2 {
			continue
		}

		header := strings.TrimSpace(parts[0])
		lyrics := strings.TrimSpace(parts[1])
		if lyrics == "" {
			continue
		}

		var title, artist string
		for _, line := range strings.Split(header, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "제목:") {
				title = strings.TrimSpace(strings.TrimPrefix(line, "제목:"))
			} else if strings.HasPrefix(line, "아티스트:") {
				artist = strings.TrimSpace(strings.TrimPrefix(line, "아티스트:"))
			}
		}

		if title == "" {
			continue
		}

		songs = append(songs, CustomSong{
			Title:  title,
			Artist: artist,
			Lyrics: lyrics,
			Tags:   []string{},
		})
	}

	return songs
}
