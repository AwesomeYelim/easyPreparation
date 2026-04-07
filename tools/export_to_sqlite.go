// tools/export_to_sqlite.go
// PostgreSQL bible_db → SQLite data/bible.db 전체 마이그레이션 스크립트
//
// 사용법: go run tools/export_to_sqlite.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

const (
	pgDSN      = "postgres://postgres:02031122@138.2.119.220:5432/bible_db?sslmode=disable"
	sqlitePath = "data/bible.db"
)

func main() {
	start := time.Now()

	// 1. 기존 SQLite 파일 삭제 (WAL/SHM 포함)
	for _, ext := range []string{"", "-wal", "-shm"} {
		os.Remove(sqlitePath + ext)
	}
	fmt.Println("[1/6] 기존 SQLite 파일 삭제 완료")

	// 2. SQLite 열기 + 스키마 생성
	lite, err := sql.Open("sqlite", sqlitePath+"?_pragma=journal_mode(wal)&_pragma=synchronous(normal)&_pragma=busy_timeout(10000)")
	if err != nil {
		log.Fatalf("SQLite 열기 실패: %v", err)
	}
	defer lite.Close()

	if err := createSchema(lite); err != nil {
		log.Fatalf("스키마 생성 실패: %v", err)
	}
	fmt.Println("[2/6] SQLite 스키마 생성 완료")

	// 3. PostgreSQL 연결
	pg, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatalf("PostgreSQL 연결 실패: %v", err)
	}
	defer pg.Close()
	if err := pg.Ping(); err != nil {
		log.Fatalf("PostgreSQL Ping 실패: %v", err)
	}
	fmt.Println("[3/6] PostgreSQL 연결 성공")

	// 4. bible_versions 복사
	n, err := copyBibleVersions(pg, lite)
	if err != nil {
		log.Fatalf("bible_versions 복사 실패: %v", err)
	}
	fmt.Printf("[4/6] bible_versions: %d건 복사 완료\n", n)

	// 5. books 복사
	n, err = copyBooks(pg, lite)
	if err != nil {
		log.Fatalf("books 복사 실패: %v", err)
	}
	fmt.Printf("[5/6] books: %d건 복사 완료\n", n)

	// 6. hymns 복사
	n, err = copyHymns(pg, lite)
	if err != nil {
		log.Fatalf("hymns 복사 실패: %v", err)
	}
	fmt.Printf("[6/6] hymns: %d건 복사 완료\n", n)

	// 7. verses 복사 (대용량 — 버전별 배치)
	total, err := copyVerses(pg, lite)
	if err != nil {
		log.Fatalf("verses 복사 실패: %v", err)
	}
	fmt.Printf("[완료] verses: 총 %d건 복사 완료\n", total)

	// 8. 검증
	fmt.Println("\n=== 데이터 검증 ===")
	verify(lite)

	fmt.Printf("\n총 소요 시간: %v\n", time.Since(start).Round(time.Millisecond))
}

func createSchema(db *sql.DB) error {
	schema := `
CREATE TABLE bible_versions (
    id   INTEGER PRIMARY KEY,
    name VARCHAR(50)  NOT NULL,
    code VARCHAR(10)  NOT NULL UNIQUE
);

CREATE TABLE books (
    id         INTEGER PRIMARY KEY,
    name_kor   VARCHAR(20)  NOT NULL,
    name_eng   VARCHAR(50)  NOT NULL,
    abbr_kor   VARCHAR(10)  NOT NULL,
    abbr_eng   VARCHAR(10)  NOT NULL,
    book_order INTEGER      NOT NULL UNIQUE
);
CREATE INDEX idx_books_order ON books(book_order);

CREATE TABLE verses (
    version_id INTEGER NOT NULL REFERENCES bible_versions(id) ON DELETE CASCADE,
    book_id    INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    chapter    INTEGER NOT NULL,
    verse      INTEGER NOT NULL,
    text       TEXT    NOT NULL,
    PRIMARY KEY (version_id, book_id, chapter, verse)
);
CREATE INDEX idx_verses_lookup ON verses(version_id, book_id, chapter, verse);

CREATE TABLE hymns (
    id         INTEGER PRIMARY KEY,
    hymnbook   VARCHAR(20)  NOT NULL DEFAULT 'new',
    number     INTEGER      NOT NULL,
    title      VARCHAR(200) NOT NULL,
    first_line VARCHAR(500),
    category   VARCHAR(100),
    lyrics     TEXT,
    has_pdf    INTEGER      NOT NULL DEFAULT 0,
    created_at DATETIME     NOT NULL DEFAULT (datetime('now')),
    UNIQUE(hymnbook, number)
);
CREATE INDEX idx_hymns_title  ON hymns(title);
CREATE INDEX idx_hymns_number ON hymns(number);
`
	_, err := db.Exec(schema)
	return err
}

func copyBibleVersions(pg, lite *sql.DB) (int, error) {
	rows, err := pg.Query("SELECT id, name, code FROM bible_versions ORDER BY id")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	tx, err := lite.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare("INSERT INTO bible_versions (id, name, code) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int
		var name, code string
		if err := rows.Scan(&id, &name, &code); err != nil {
			tx.Rollback()
			return 0, err
		}
		if _, err := stmt.Exec(id, name, code); err != nil {
			tx.Rollback()
			return 0, err
		}
		count++
	}
	return count, tx.Commit()
}

func copyBooks(pg, lite *sql.DB) (int, error) {
	rows, err := pg.Query("SELECT id, name_kor, name_eng, abbr_kor, abbr_eng, book_order FROM books ORDER BY id")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	tx, err := lite.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare("INSERT INTO books (id, name_kor, name_eng, abbr_kor, abbr_eng, book_order) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id, bookOrder int
		var nameKor, nameEng, abbrKor, abbrEng string
		if err := rows.Scan(&id, &nameKor, &nameEng, &abbrKor, &abbrEng, &bookOrder); err != nil {
			tx.Rollback()
			return 0, err
		}
		if _, err := stmt.Exec(id, nameKor, nameEng, abbrKor, abbrEng, bookOrder); err != nil {
			tx.Rollback()
			return 0, err
		}
		count++
	}
	return count, tx.Commit()
}

func copyHymns(pg, lite *sql.DB) (int, error) {
	rows, err := pg.Query(`
		SELECT id, hymnbook, number, title,
		       COALESCE(first_line,''), COALESCE(category,''), COALESCE(lyrics,''),
		       has_pdf, created_at
		FROM hymns ORDER BY id`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	tx, err := lite.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(`INSERT INTO hymns (id, hymnbook, number, title, first_line, category, lyrics, has_pdf, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id, number, hasPdf int
		var hymnbook, title, firstLine, category, lyrics string
		var createdAt time.Time
		if err := rows.Scan(&id, &hymnbook, &number, &title, &firstLine, &category, &lyrics, &hasPdf, &createdAt); err != nil {
			tx.Rollback()
			return 0, err
		}
		if _, err := stmt.Exec(id, hymnbook, number, title, firstLine, category, lyrics, hasPdf, createdAt.Format("2006-01-02 15:04:05")); err != nil {
			tx.Rollback()
			return 0, err
		}
		count++
	}
	return count, tx.Commit()
}

func copyVerses(pg, lite *sql.DB) (int, error) {
	// 버전별로 복사하여 진행 상황 표시
	var versions []struct {
		id   int
		name string
	}
	rows, err := pg.Query("SELECT id, name FROM bible_versions ORDER BY id")
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var v struct {
			id   int
			name string
		}
		if err := rows.Scan(&v.id, &v.name); err != nil {
			rows.Close()
			return 0, err
		}
		versions = append(versions, v)
	}
	rows.Close()

	total := 0
	batchSize := 5000

	for _, v := range versions {
		vStart := time.Now()
		vRows, err := pg.Query(
			"SELECT version_id, book_id, chapter, verse, text FROM verses WHERE version_id = $1 ORDER BY book_id, chapter, verse",
			v.id,
		)
		if err != nil {
			return total, fmt.Errorf("version %d 조회 실패: %w", v.id, err)
		}

		tx, err := lite.Begin()
		if err != nil {
			vRows.Close()
			return total, err
		}
		stmt, err := tx.Prepare("INSERT INTO verses (version_id, book_id, chapter, verse, text) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			tx.Rollback()
			vRows.Close()
			return total, err
		}

		count := 0
		for vRows.Next() {
			var versionID, bookID, chapter, verse int
			var text string
			if err := vRows.Scan(&versionID, &bookID, &chapter, &verse, &text); err != nil {
				stmt.Close()
				tx.Rollback()
				vRows.Close()
				return total, err
			}
			if _, err := stmt.Exec(versionID, bookID, chapter, verse, text); err != nil {
				stmt.Close()
				tx.Rollback()
				vRows.Close()
				return total, err
			}
			count++

			// 배치 커밋 (대용량 INSERT 성능 최적화)
			if count%batchSize == 0 {
				stmt.Close()
				if err := tx.Commit(); err != nil {
					vRows.Close()
					return total, err
				}
				tx, err = lite.Begin()
				if err != nil {
					vRows.Close()
					return total, err
				}
				stmt, err = tx.Prepare("INSERT INTO verses (version_id, book_id, chapter, verse, text) VALUES (?, ?, ?, ?, ?)")
				if err != nil {
					tx.Rollback()
					vRows.Close()
					return total, err
				}
			}
		}
		vRows.Close()
		stmt.Close()
		if err := tx.Commit(); err != nil {
			return total, err
		}

		total += count
		fmt.Printf("  - [%d/%d] %s (id=%d): %d건 (%.1fs)\n",
			v.id, len(versions), v.name, v.id, count, time.Since(vStart).Seconds())
	}

	return total, nil
}

func verify(lite *sql.DB) {
	tables := []string{"bible_versions", "books", "verses", "hymns"}
	for _, t := range tables {
		var count int
		err := lite.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", t)).Scan(&count)
		if err != nil {
			fmt.Printf("  %s: 조회 실패 — %v\n", t, err)
		} else {
			fmt.Printf("  %s: %d건\n", t, count)
		}
	}

	// 버전별 verses 수도 확인
	fmt.Println("\n=== 버전별 verses 수 ===")
	rows, err := lite.Query(`
		SELECT bv.name, bv.code, COUNT(v.verse) as cnt
		FROM verses v
		JOIN bible_versions bv ON bv.id = v.version_id
		GROUP BY v.version_id
		ORDER BY v.version_id`)
	if err != nil {
		fmt.Printf("  버전별 조회 실패: %v\n", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var name, code string
		var cnt int
		rows.Scan(&name, &code, &cnt)
		fmt.Printf("  %s (%s): %d건\n", name, code, cnt)
	}
}
