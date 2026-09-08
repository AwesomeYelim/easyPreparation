package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"easyPreparation_1.0/internal/httpx"
	"easyPreparation_1.0/internal/quote"
)

var apiDB *sql.DB   // 앱 DB (SQLite — churches, licenses, settings)
var bibleDB *sql.DB // 성경 DB (PostgreSQL — verses, books, hymns)

// InitAPIDB — handlers 패키지에서 사용할 앱 DB 연결 설정
func InitAPIDB(db *sql.DB) {
	apiDB = db
}

// InitBibleDB — handlers 패키지에서 사용할 Bible DB 연결 설정
func InitBibleDB(db *sql.DB) {
	bibleDB = db
}

// BibleBooksHandler — GET /api/bible/books
// bible.db의 books 테이블에서 책 목록 + 장수를 반환 (bible_info.json 파일 불필요)
func BibleBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := quote.GetBooksWithChapters()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "bible books not available: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(books)
}

// UserHandler — GET /api/user?email=xxx  /  POST /api/user
func UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUserHandler(w, r)
	case http.MethodPost:
		upsertUserHandler(w, r)
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	// Desktop 앱: email 없으면 church_id=1 기본 조회
	var row *sql.Row
	if email == "" {
		row = apiDB.QueryRow(`
			SELECT id, name, english_name, email
			FROM churches WHERE id = 1 LIMIT 1
		`)
	} else {
		row = apiDB.QueryRow(`
			SELECT id, name, english_name, email
			FROM churches WHERE email = ? LIMIT 1
		`, email)
	}

	var id int
	var name, englishName, emailVal string
	if err := row.Scan(&id, &name, &englishName, &emailVal); err != nil {
		httpx.Error(w, http.StatusNotFound, "User church info not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           id,
		"name":         name,
		"english_name": englishName,
		"email":        emailVal,
	})
}

func upsertUserHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		EnglishName string `json:"english_name"`
		Email       string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		httpx.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	var exists int
	_ = apiDB.QueryRow("SELECT 1 FROM churches WHERE email=? LIMIT 1", body.Email).Scan(&exists)

	if exists == 1 {
		if _, err := apiDB.Exec("UPDATE churches SET name=?, english_name=? WHERE email=?",
			body.Name, body.EnglishName, body.Email); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "update failed")
			return
		}
	} else {
		if _, err := apiDB.Exec("INSERT INTO churches (name, english_name, email) VALUES (?,?,?)",
			body.Name, body.EnglishName, body.Email); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "insert failed")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// BibleVersionsHandler — GET /api/bible/versions
func BibleVersionsHandler(w http.ResponseWriter, r *http.Request) {
	versions, err := quote.GetBibleVersions()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "versions not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(versions)
}

// BibleSearchHandler — GET /api/bible/search?q=&version=
func BibleSearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httpx.Error(w, http.StatusBadRequest, "q parameter required")
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
		httpx.Error(w, http.StatusInternalServerError, "search failed")
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
		httpx.Error(w, http.StatusBadRequest, "book and chapter required")
		return
	}

	bookOrder, err := strconv.Atoi(bookStr)
	if err != nil || bookOrder <= 0 || bookOrder > 66 {
		httpx.Error(w, http.StatusBadRequest, "invalid book")
		return
	}
	chapter, err := strconv.Atoi(chapterStr)
	if err != nil || chapter < 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid chapter")
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
			httpx.Error(w, http.StatusInternalServerError, "chapter count failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"chapters": count})
		return
	}

	verses, err := quote.GetChapterVerses(versionID, bookOrder, chapter)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "verses not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(verses)
}
