package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// HymnListHandler — GET /api/hymns?page=&limit=&book=
func HymnListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	page := 1
	limit := 50
	hymnbook := ""

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if b := r.URL.Query().Get("book"); b != "" {
		hymnbook = b
	}

	offset := (page - 1) * limit

	var query string
	var args []interface{}

	if hymnbook != "" {
		query = `SELECT id, hymnbook, number, title, first_line, category, has_pdf
				 FROM hymns WHERE hymnbook = ? ORDER BY number LIMIT ? OFFSET ?`
		args = []interface{}{hymnbook, limit, offset}
	} else {
		query = `SELECT id, hymnbook, number, title, first_line, category, has_pdf
				 FROM hymns ORDER BY hymnbook, number LIMIT ? OFFSET ?`
		args = []interface{}{limit, offset}
	}

	rows, err := bibleDB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"query failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var hymns []map[string]interface{}
	for rows.Next() {
		var id, number int
		var hymnbookVal, title string
		var firstLine, category *string
		var hasPdf bool

		if err := rows.Scan(&id, &hymnbookVal, &number, &title, &firstLine, &category, &hasPdf); err != nil {
			continue
		}
		h := map[string]interface{}{
			"id":       id,
			"hymnbook": hymnbookVal,
			"number":   number,
			"title":    title,
			"has_pdf":  hasPdf,
		}
		if firstLine != nil {
			h["first_line"] = *firstLine
		}
		if category != nil {
			h["category"] = *category
		}
		hymns = append(hymns, h)
	}

	if hymns == nil {
		hymns = []map[string]interface{}{}
	}

	// 전체 수 조회
	var total int
	if hymnbook != "" {
		_ = bibleDB.QueryRow("SELECT COUNT(*) FROM hymns WHERE hymnbook = ?", hymnbook).Scan(&total)
	} else {
		_ = bibleDB.QueryRow("SELECT COUNT(*) FROM hymns").Scan(&total)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"hymns": hymns,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// HymnSearchHandler — GET /api/hymns/search?q=&type=
func HymnSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, `{"error":"q parameter required"}`, http.StatusBadRequest)
		return
	}

	searchType := r.URL.Query().Get("type")

	var query string
	var args []interface{}

	if searchType == "" {
		if _, err := strconv.Atoi(q); err == nil {
			searchType = "number"
		} else {
			searchType = "title"
		}
	}

	switch searchType {
	case "number":
		num, err := strconv.Atoi(q)
		if err != nil {
			http.Error(w, `{"error":"invalid number"}`, http.StatusBadRequest)
			return
		}
		query = `SELECT id, hymnbook, number, title, first_line, category, lyrics, has_pdf
				 FROM hymns WHERE number = ? ORDER BY hymnbook LIMIT 10`
		args = []interface{}{num}
	case "lyrics":
		query = `SELECT id, hymnbook, number, title, first_line, category, lyrics, has_pdf
				 FROM hymns WHERE lyrics LIKE '%' || ? || '%' ORDER BY number LIMIT 50`
		args = []interface{}{q}
	default: // title
		query = `SELECT id, hymnbook, number, title, first_line, category, lyrics, has_pdf
				 FROM hymns WHERE title LIKE '%' || ? || '%' OR first_line LIKE '%' || ? || '%'
				 ORDER BY number LIMIT 50`
		args = []interface{}{q, q}
	}

	rows, err := bibleDB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"search failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, number int
		var hymnbook, title string
		var firstLine, category, lyrics *string
		var hasPdf bool

		if err := rows.Scan(&id, &hymnbook, &number, &title, &firstLine, &category, &lyrics, &hasPdf); err != nil {
			continue
		}
		h := map[string]interface{}{
			"id":       id,
			"hymnbook": hymnbook,
			"number":   number,
			"title":    title,
			"has_pdf":  hasPdf,
		}
		if firstLine != nil {
			h["first_line"] = *firstLine
		}
		if category != nil {
			h["category"] = *category
		}
		if lyrics != nil {
			h["lyrics"] = *lyrics
		}
		results = append(results, h)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

// HymnDetailHandler — GET /api/hymns/detail?number=&book=
func HymnDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	numStr := r.URL.Query().Get("number")
	if numStr == "" {
		http.Error(w, `{"error":"number required"}`, http.StatusBadRequest)
		return
	}
	number, err := strconv.Atoi(numStr)
	if err != nil {
		http.Error(w, `{"error":"invalid number"}`, http.StatusBadRequest)
		return
	}

	hymnbook := r.URL.Query().Get("book")
	if hymnbook == "" {
		hymnbook = "new"
	}

	var id int
	var title string
	var firstLine, category, lyrics *string
	var hasPdf bool

	err = bibleDB.QueryRow(`
		SELECT id, title, first_line, category, lyrics, has_pdf
		FROM hymns WHERE hymnbook = ? AND number = ?
	`, hymnbook, number).Scan(&id, &title, &firstLine, &category, &lyrics, &hasPdf)

	if err != nil {
		http.Error(w, `{"error":"hymn not found"}`, http.StatusNotFound)
		return
	}

	h := map[string]interface{}{
		"id":       id,
		"hymnbook": hymnbook,
		"number":   number,
		"title":    title,
		"has_pdf":  hasPdf,
	}
	if firstLine != nil {
		h["first_line"] = *firstLine
	}
	if category != nil {
		h["category"] = *category
	}
	if lyrics != nil {
		h["lyrics"] = *lyrics
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h)
}
