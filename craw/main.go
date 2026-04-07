package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

// GoodTV API 버전 코드 → DB 버전 ID 매핑
type VersionMapping struct {
	GoodTVCode int
	DBID       int
	Name       string
}

var versionMappings = []VersionMapping{
	{0, 1, "개역개정"},
	{1, 2, "개역한글"},
	{2, 3, "공동번역"},
	{3, 4, "표준새번역"},
	{5, 5, "NIV"},
	{6, 6, "KJV"},
	{7, 7, "우리말성경"},
}

type BibleAPIResponse struct {
	Data struct {
		Testament   string `json:"testament"`
		Bookname    string `json:"bookname"`
		BooknameAbb string `json:"bookname_abb"`
		Data        struct {
			Version1 struct {
				Version     int    `json:"version"`
				Jang        int    `json:"jang"`
				VersionName string `json:"version_name"`
				Bookname    string `json:"bookname"`
				BooknameAbb string `json:"bookname_abb"`
				Content     []struct {
					Jul  int    `json:"jul"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"version1"`
		} `json:"data"`
	} `json:"data"`
}

type BibleBook struct {
	Index    int    `json:"index"`
	Kor      string `json:"kor"`
	Eng      string `json:"eng"`
	Chapters []int  `json:"chapters"`
}

func main() {
	versionFlag := flag.String("version", "all", "크롤링할 버전: all 또는 DB ID (예: 2)")
	bibleInfoFlag := flag.String("bible-info", "bible_info.json", "bible_info.json 경로")
	dsnFlag := flag.String("dsn", "postgres://postgres:02031122@138.2.119.220:5432/bible_db?sslmode=disable", "PostgreSQL DSN")
	flag.Parse()

	// PostgreSQL 연결
	db, err := sql.Open("postgres", *dsnFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("DB 연결 실패: %v", err)
	}
	db.SetMaxOpenConns(5)
	log.Println("PostgreSQL 연결 성공")

	bibleBooks, err := LoadBibleBooksFromJSON(*bibleInfoFlag)
	if err != nil {
		log.Fatalf("책 정보 로딩 실패: %v", err)
	}

	// 크롤링할 버전 결정
	var targets []VersionMapping
	if *versionFlag == "all" {
		targets = versionMappings
	} else {
		var dbID int
		if _, err := fmt.Sscanf(*versionFlag, "%d", &dbID); err != nil {
			log.Fatalf("잘못된 버전 ID: %s", *versionFlag)
		}
		for _, vm := range versionMappings {
			if vm.DBID == dbID {
				targets = append(targets, vm)
				break
			}
		}
		if len(targets) == 0 {
			log.Fatalf("버전 ID %d를 찾을 수 없습니다", dbID)
		}
	}

	// bible_versions 테이블에 버전 등록
	ensureBibleVersions(db)

	for _, vm := range targets {
		// 이미 크롤링된 절 수 확인
		var existing int
		db.QueryRow("SELECT COUNT(*) FROM verses WHERE version_id = $1", vm.DBID).Scan(&existing)
		if existing >= 30000 {
			log.Printf("=== [%d] %s — 이미 %d절 존재, 스킵 ===", vm.DBID, vm.Name, existing)
			continue
		}
		log.Printf("=== 버전 크롤링 시작: [%d] %s (GoodTV code=%d, 기존 %d절) ===", vm.DBID, vm.Name, vm.GoodTVCode, existing)
		crawlVersion(db, bibleBooks, vm)
	}

	log.Println("=== 크롤링 완료 ===")
}

func ensureBibleVersions(db *sql.DB) {
	versions := []struct {
		ID   int
		Name string
		Code string
	}{
		{1, "개역개정", "NKRV"},
		{2, "개역한글", "KRV"},
		{3, "공동번역", "COMMON"},
		{4, "표준새번역", "NKSV"},
		{5, "NIV", "NIV"},
		{6, "KJV", "KJV"},
		{7, "우리말성경", "URNB"},
	}
	for _, v := range versions {
		_, err := db.Exec(
			`INSERT INTO bible_versions (id, name, code) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`,
			v.ID, v.Name, v.Code,
		)
		if err != nil {
			log.Printf("버전 등록 실패 [%d] %s: %v", v.ID, v.Name, err)
		}
	}
}

func crawlVersion(db *sql.DB, bibleBooks map[string]BibleBook, vm VersionMapping) {
	stmt, err := db.Prepare(`
		INSERT INTO verses (version_id, book_id, chapter, verse, text)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (version_id, book_id, chapter, verse) DO NOTHING
	`)
	if err != nil {
		log.Fatalf("prepared statement 실패: %v", err)
	}
	defer stmt.Close()

	for _, book := range bibleBooks {
		for chapter := 1; chapter <= len(book.Chapters); chapter++ {
			// 이미 해당 장이 크롤링되었는지 확인
			var cnt int
			db.QueryRow("SELECT COUNT(*) FROM verses WHERE version_id=$1 AND book_id=$2 AND chapter=$3",
				vm.DBID, book.Index, chapter).Scan(&cnt)
			if cnt > 0 {
				continue
			}

			url := fmt.Sprintf(
				"https://goodtvbible.goodtv.co.kr/api/onlinebible/bibleread/read-all?version1=%d&version2=&version3=&bible_code=%d&jang=%d",
				vm.GoodTVCode, book.Index, chapter,
			)

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("  요청 실패 (%s %d장): %v", book.Kor, chapter, err)
				time.Sleep(300 * time.Millisecond)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("  응답 읽기 실패: %v", err)
				time.Sleep(300 * time.Millisecond)
				continue
			}

			var res BibleAPIResponse
			if err := json.Unmarshal(body, &res); err != nil {
				log.Printf("  JSON 파싱 실패 (%s %d장): %v", book.Kor, chapter, err)
				time.Sleep(300 * time.Millisecond)
				continue
			}

			contents := res.Data.Data.Version1.Content
			for _, verse := range contents {
				_, err := stmt.Exec(vm.DBID, book.Index, chapter, verse.Jul, verse.Text)
				if err != nil {
					log.Printf("  DB 저장 실패 (%s %d:%d): %v", book.Kor, chapter, verse.Jul, err)
				}
			}

			log.Printf("  [%s] %s %d장 — %d절 저장", vm.Name, book.Kor, chapter, len(contents))
			time.Sleep(150 * time.Millisecond)
		}
	}
}

func LoadBibleBooksFromJSON(path string) (map[string]BibleBook, error) {
	absPath, _ := filepath.Abs(path)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var books map[string]BibleBook
	if err := json.Unmarshal(data, &books); err != nil {
		return nil, err
	}
	return books, nil
}
