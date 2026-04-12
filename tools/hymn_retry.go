package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:02031122@138.2.119.220:5432/bible_db?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("DB 연결 실패: %v", err)
	}

	// 가사 없는 찬송가 조회
	rows, err := db.Query("SELECT number, title FROM hymns WHERE hymnbook='new' AND (lyrics IS NULL OR lyrics = '') ORDER BY number")
	if err != nil {
		log.Fatal(err)
	}
	var missing []struct {
		Number int
		Title  string
	}
	for rows.Next() {
		var num int
		var title string
		rows.Scan(&num, &title)
		missing = append(missing, struct {
			Number int
			Title  string
		}{num, title})
	}
	rows.Close()
	log.Printf("가사 없는 찬송가: %d곡", len(missing))

	reDiv := regexp.MustCompile(`<div class=['"]hymnlyrics['"]>(.*?)</div>`)
	reBr := regexp.MustCompile(`<br\s*/?>`)
	reTag := regexp.MustCompile(`<[^>]+>`)

	stmt, err := db.Prepare("UPDATE hymns SET lyrics = $1 WHERE hymnbook = 'new' AND number = $2")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	success, fail := 0, 0

	for _, h := range missing {
		// 3번까지 재시도
		var lyrics string
		for attempt := 1; attempt <= 3; attempt++ {
			lyrics = fetchLyrics(h.Number, reDiv, reBr, reTag)
			if lyrics != "" {
				break
			}
			log.Printf("  [%d] 시도 %d 실패, 재시도...", h.Number, attempt)
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}

		if lyrics == "" {
			log.Printf("  [%d] %s — 가사 없음 (3회 실패)", h.Number, h.Title)
			fail++
			continue
		}

		if _, err := stmt.Exec(lyrics, h.Number); err != nil {
			log.Printf("  [%d] DB 저장 실패: %v", h.Number, err)
			fail++
		} else {
			preview := lyrics
			if idx := strings.Index(preview, "\n"); idx > 0 {
				preview = preview[:idx]
			}
			if len(preview) > 50 {
				preview = preview[:50]
			}
			log.Printf("  [%d] %s — 저장 완료: %s", h.Number, h.Title, preview)
			success++
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("=== 완료: 성공 %d, 실패 %d ===", success, fail)
}

func fetchLyrics(num int, reDiv, reBr, reTag *regexp.Regexp) string {
	url := fmt.Sprintf("https://www.prayertents.com/hymns?nh=%d", num)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return ""
	}

	html := string(body)
	matches := reDiv.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return ""
	}

	korLyrics := matches[0][1]
	korLyrics = reBr.ReplaceAllString(korLyrics, "\n")
	korLyrics = reTag.ReplaceAllString(korLyrics, "")
	korLyrics = strings.TrimSpace(korLyrics)

	rawLines := strings.Split(korLyrics, "\n")
	var cleaned []string
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "아멘" {
			continue
		}
		if line == "" {
			if len(cleaned) > 0 && cleaned[len(cleaned)-1] != "" {
				cleaned = append(cleaned, "")
			}
			continue
		}
		cleaned = append(cleaned, line)
	}
	for len(cleaned) > 0 && cleaned[len(cleaned)-1] == "" {
		cleaned = cleaned[:len(cleaned)-1]
	}
	return strings.Join(cleaned, "\n")
}
