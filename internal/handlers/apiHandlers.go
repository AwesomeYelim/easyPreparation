package handlers

import (
	"database/sql"
	"easyPreparation_1.0/internal/quote"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var apiDB *sql.DB     // 앱 DB (SQLite — churches, licenses, settings)
var bibleDB *sql.DB   // 성경 DB (PostgreSQL — verses, books, hymns)

// InitAPIDB — handlers 패키지에서 사용할 앱 DB 연결 설정
func InitAPIDB(db *sql.DB) {
	apiDB = db
}

// InitBibleDB — handlers 패키지에서 사용할 Bible DB 연결 설정
func InitBibleDB(db *sql.DB) {
	bibleDB = db
}

// BibleBooksHandler — GET /api/bible/books
// config/bible_info.json을 그대로 서빙 (프론트 bible_info.json 대체)
func BibleBooksHandler(w http.ResponseWriter, r *http.Request) {
	path, _ := filepath.Abs("./config/bible_info.json")
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "bible_info not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// UserHandler — GET /api/user?email=xxx  /  POST /api/user
func UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUserHandler(w, r)
	case http.MethodPost:
		upsertUserHandler(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	// Desktop 앱: email 없으면 church_id=1 기본 조회
	var row *sql.Row
	if email == "" {
		row = apiDB.QueryRow(`
			SELECT id, name, english_name, email,
			       COALESCE(figma_key, '') AS figma_key,
			       COALESCE(figma_token, '') AS figma_token
			FROM churches WHERE id = 1 LIMIT 1
		`)
	} else {
		row = apiDB.QueryRow(`
			SELECT id, name, english_name, email,
			       COALESCE(figma_key, '') AS figma_key,
			       COALESCE(figma_token, '') AS figma_token
			FROM churches WHERE email = ? LIMIT 1
		`, email)
	}

	var id int
	var name, englishName, emailVal, figmaKey, figmaToken string
	if err := row.Scan(&id, &name, &englishName, &emailVal, &figmaKey, &figmaToken); err != nil {
		http.Error(w, `{"error":"User church info not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           id,
		"name":         name,
		"english_name": englishName,
		"email":        emailVal,
		"figmaInfo": map[string]string{
			"key":   figmaKey,
			"token": figmaToken,
		},
	})
}

func upsertUserHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		EnglishName string `json:"english_name"`
		Email       string `json:"email"`
		FigmaInfo   *struct {
			Key   string `json:"key"`
			Token string `json:"token"`
		} `json:"figmaInfo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	var exists int
	_ = apiDB.QueryRow("SELECT 1 FROM churches WHERE email=? LIMIT 1", body.Email).Scan(&exists)

	if exists == 1 {
		if body.FigmaInfo != nil {
			_, _ = apiDB.Exec("UPDATE churches SET name=?, english_name=?, figma_key=?, figma_token=? WHERE email=?",
				body.Name, body.EnglishName, body.FigmaInfo.Key, body.FigmaInfo.Token, body.Email)
		} else {
			_, _ = apiDB.Exec("UPDATE churches SET name=?, english_name=? WHERE email=?",
				body.Name, body.EnglishName, body.Email)
		}
	} else {
		figmaKey, figmaToken := "", ""
		if body.FigmaInfo != nil {
			figmaKey = body.FigmaInfo.Key
			figmaToken = body.FigmaInfo.Token
		}
		_, _ = apiDB.Exec("INSERT INTO churches (name, english_name, email, figma_key, figma_token) VALUES (?,?,?,?,?)",
			body.Name, body.EnglishName, body.Email, figmaKey, figmaToken)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// BibleVersionsHandler — GET /api/bible/versions
func BibleVersionsHandler(w http.ResponseWriter, r *http.Request) {
	versions, err := quote.GetBibleVersions()
	if err != nil {
		http.Error(w, `{"error":"versions not found"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(versions)
}

// BibleSearchHandler — GET /api/bible/search?q=&version=
func BibleSearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, `{"error":"q parameter required"}`, http.StatusBadRequest)
		return
	}
	versionID := 1
	if v := r.URL.Query().Get("version"); v != "" {
		if vid, err := strconv.Atoi(v); err == nil {
			versionID = vid
		}
	}

	results, err := quote.SearchBibleVerses(q, versionID, 50)
	if err != nil {
		http.Error(w, `{"error":"search failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

// BibleVersesHandler — GET /api/bible/verses?book=&chapter=&version=
func BibleVersesHandler(w http.ResponseWriter, r *http.Request) {
	bookStr := r.URL.Query().Get("book")
	chapterStr := r.URL.Query().Get("chapter")
	if bookStr == "" || chapterStr == "" {
		http.Error(w, `{"error":"book and chapter required"}`, http.StatusBadRequest)
		return
	}

	bookOrder, err := strconv.Atoi(bookStr)
	if err != nil {
		http.Error(w, `{"error":"invalid book"}`, http.StatusBadRequest)
		return
	}
	chapter, err := strconv.Atoi(chapterStr)
	if err != nil {
		http.Error(w, `{"error":"invalid chapter"}`, http.StatusBadRequest)
		return
	}

	versionID := 1
	if v := r.URL.Query().Get("version"); v != "" {
		if vid, err := strconv.Atoi(v); err == nil {
			versionID = vid
		}
	}

	// chapter=0이면 장 수만 반환
	if chapter == 0 {
		count, err := quote.GetBookChapterCount(versionID, bookOrder)
		if err != nil {
			http.Error(w, `{"error":"chapter count failed"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"chapters": count})
		return
	}

	verses, err := quote.GetChapterVerses(versionID, bookOrder, chapter)
	if err != nil {
		http.Error(w, `{"error":"verses not found"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(verses)
}

