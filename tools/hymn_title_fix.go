// +build ignore

// prayertents.com에서 새찬송가 1~645번 제목 크롤링 → SQLite + PostgreSQL 동시 업데이트
package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

const (
	pgDSN     = "postgres://postgres:02031122@138.2.119.220:5432/bible_db?sslmode=disable"
	sqliteDB  = "data/bible.db"
	maxRetry  = 3
	batchSize = 50 // 배치 크기 (50곡씩 처리 후 중간 로그)
)

var reTitle = regexp.MustCompile(`hymntitleheader'>(.*?)</div>`)

func main() {
	// --- SQLite 연결 ---
	lite, err := sql.Open("sqlite", sqliteDB)
	if err != nil {
		log.Fatalf("SQLite 열기 실패: %v", err)
	}
	defer lite.Close()
	if err := lite.Ping(); err != nil {
		log.Fatalf("SQLite 연결 실패: %v", err)
	}
	log.Println("SQLite 연결 성공:", sqliteDB)

	// --- PostgreSQL 연결 ---
	pg, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatalf("PostgreSQL 열기 실패: %v", err)
	}
	defer pg.Close()
	if err := pg.Ping(); err != nil {
		log.Fatalf("PostgreSQL 연결 실패: %v", err)
	}
	log.Println("PostgreSQL 연결 성공")

	// --- Prepared Statements ---
	liteStmt, err := lite.Prepare("UPDATE hymns SET title = ? WHERE hymnbook = 'new' AND number = ?")
	if err != nil {
		log.Fatalf("SQLite prepare 실패: %v", err)
	}
	defer liteStmt.Close()

	pgStmt, err := pg.Prepare("UPDATE hymns SET title = $1 WHERE hymnbook = 'new' AND number = $2")
	if err != nil {
		log.Fatalf("PostgreSQL prepare 실패: %v", err)
	}
	defer pgStmt.Close()

	client := &http.Client{Timeout: 15 * time.Second}

	var (
		mu              sync.Mutex
		success, fail   int
		failedNums      []int
	)

	log.Println("=== 새찬송가 1~645번 제목 크롤링 시작 ===")

	for num := 1; num <= 645; num++ {
		title, err := fetchTitle(client, num)
		if err != nil {
			log.Printf("  [%d] 크롤링 실패: %v", num, err)
			mu.Lock()
			fail++
			failedNums = append(failedNums, num)
			mu.Unlock()
			time.Sleep(300 * time.Millisecond)
			continue
		}

		// SQLite 업데이트
		if _, err := liteStmt.Exec(title, num); err != nil {
			log.Printf("  [%d] SQLite 업데이트 실패: %v", num, err)
			mu.Lock()
			fail++
			failedNums = append(failedNums, num)
			mu.Unlock()
			time.Sleep(200 * time.Millisecond)
			continue
		}

		// PostgreSQL 업데이트
		if _, err := pgStmt.Exec(title, num); err != nil {
			log.Printf("  [%d] PostgreSQL 업데이트 실패: %v", num, err)
			mu.Lock()
			fail++
			failedNums = append(failedNums, num)
			mu.Unlock()
			time.Sleep(200 * time.Millisecond)
			continue
		}

		mu.Lock()
		success++
		mu.Unlock()

		if num%batchSize == 0 {
			log.Printf("  진행: %d/645 (성공 %d, 실패 %d)", num, success, fail)
		}

		// 서버 부하 방지
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("=== 1차 완료: 성공 %d, 실패 %d ===", success, fail)

	// --- 실패한 번호 재시도 (최대 2회 추가) ---
	for retry := 1; retry <= 2 && len(failedNums) > 0; retry++ {
		log.Printf("=== 재시도 %d회차: %d곡 ===", retry, len(failedNums))
		time.Sleep(2 * time.Second)

		retryNums := make([]int, len(failedNums))
		copy(retryNums, failedNums)
		failedNums = nil

		for _, num := range retryNums {
			title, err := fetchTitle(client, num)
			if err != nil {
				log.Printf("  [%d] 재시도 크롤링 실패: %v", num, err)
				failedNums = append(failedNums, num)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			liteErr := updateDB(liteStmt, title, num)
			pgErr := updateDB(pgStmt, title, num)

			if liteErr != nil || pgErr != nil {
				if liteErr != nil {
					log.Printf("  [%d] 재시도 SQLite 실패: %v", num, liteErr)
				}
				if pgErr != nil {
					log.Printf("  [%d] 재시도 PostgreSQL 실패: %v", num, pgErr)
				}
				failedNums = append(failedNums, num)
			} else {
				success++
				fail--
			}
			time.Sleep(300 * time.Millisecond)
		}
	}

	log.Printf("=== 최종 결과: 성공 %d, 실패 %d ===", success, fail)
	if len(failedNums) > 0 {
		log.Printf("  실패 목록: %v", failedNums)
	}

	// --- 검증: 샘플 10곡 확인 ---
	log.Println("=== 검증: 샘플 조회 ===")
	sampleNums := []int{1, 50, 100, 200, 300, 400, 500, 600, 640, 645}
	for _, n := range sampleNums {
		var liteTitle, pgTitle string
		lite.QueryRow("SELECT title FROM hymns WHERE hymnbook='new' AND number=?", n).Scan(&liteTitle)
		pg.QueryRow("SELECT title FROM hymns WHERE hymnbook='new' AND number=$1", n).Scan(&pgTitle)
		match := "OK"
		if liteTitle != pgTitle {
			match = "MISMATCH"
		}
		log.Printf("  [%d] SQLite=%-30s PG=%-30s %s", n, liteTitle, pgTitle, match)
	}
}

// fetchTitle: prayertents.com에서 제목 크롤링 (최대 maxRetry회 재시도)
func fetchTitle(client *http.Client, num int) (string, error) {
	url := fmt.Sprintf("https://www.prayertents.com/hymns?nh=%d", num)

	var lastErr error
	for attempt := 1; attempt <= maxRetry; attempt++ {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("요청 실패 (시도 %d): %w", attempt, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("응답 읽기 실패 (시도 %d): %w", attempt, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		html := string(body)
		matches := reTitle.FindStringSubmatch(html)
		if len(matches) < 2 {
			lastErr = fmt.Errorf("hymntitleheader 없음 (시도 %d)", attempt)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		title := strings.TrimSpace(matches[1])
		if title == "" {
			lastErr = fmt.Errorf("빈 제목 (시도 %d)", attempt)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		return title, nil
	}

	return "", lastErr
}

func updateDB(stmt *sql.Stmt, title string, num int) error {
	_, err := stmt.Exec(title, num)
	return err
}
