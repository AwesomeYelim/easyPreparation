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
	log.Println("PostgreSQL 연결 성공")

	reDiv := regexp.MustCompile(`<div class=['"]hymnlyrics['"]>(.*?)</div>`)
	reBr := regexp.MustCompile(`<br\s*/?>`)
	reTag := regexp.MustCompile(`<[^>]+>`)

	stmt, err := db.Prepare("UPDATE hymns SET lyrics = $1 WHERE hymnbook = 'new' AND number = $2")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	success, skip, fail := 0, 0, 0

	for num := 1; num <= 645; num++ {
		// 이미 가사가 있는지 확인
		var existing sql.NullString
		db.QueryRow("SELECT lyrics FROM hymns WHERE hymnbook='new' AND number=$1", num).Scan(&existing)
		if existing.Valid && strings.TrimSpace(existing.String) != "" {
			skip++
			continue
		}

		url := fmt.Sprintf("https://www.prayertents.com/hymns?nh=%d", num)

		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("  [%d] 요청 실패: %v", num, err)
			fail++
			time.Sleep(300 * time.Millisecond)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("  [%d] 응답 읽기 실패: %v", num, err)
			fail++
			time.Sleep(300 * time.Millisecond)
			continue
		}

		html := string(body)

		// hymnlyrics div에서 한국어 가사 추출 (첫 번째가 한국어)
		matches := reDiv.FindAllStringSubmatch(html, -1)
		if len(matches) == 0 {
			log.Printf("  [%d] 가사 div 없음", num)
			fail++
			time.Sleep(300 * time.Millisecond)
			continue
		}

		// 첫 번째 div가 한국어 가사
		korLyrics := matches[0][1]
		korLyrics = reBr.ReplaceAllString(korLyrics, "\n")
		korLyrics = reTag.ReplaceAllString(korLyrics, "")
		korLyrics = strings.TrimSpace(korLyrics)

		// "아멘" 제거 + 빈 줄 정리
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

		// 끝의 빈 줄 제거
		for len(cleaned) > 0 && cleaned[len(cleaned)-1] == "" {
			cleaned = cleaned[:len(cleaned)-1]
		}

		lyrics := strings.Join(cleaned, "\n")
		if lyrics == "" {
			log.Printf("  [%d] 빈 가사", num)
			fail++
			time.Sleep(300 * time.Millisecond)
			continue
		}

		if _, err := stmt.Exec(lyrics, num); err != nil {
			log.Printf("  [%d] DB 저장 실패: %v", num, err)
			fail++
		} else {
			preview := lyrics
			if idx := strings.Index(preview, "\n"); idx > 0 {
				preview = preview[:idx]
			}
			if len(preview) > 50 {
				preview = preview[:50]
			}
			log.Printf("  [%d] 저장 — %s", num, preview)
			success++
		}
		time.Sleep(250 * time.Millisecond)
	}

	log.Printf("=== 완료: 성공 %d, 스킵 %d, 실패 %d ===", success, skip, fail)
}
