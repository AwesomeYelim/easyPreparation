package quote

import (
	"database/sql"
	"easyPreparation_1.0/internal/parser"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite" // SQLite 드라이버
)

// LoadDSN은 환경변수 DB_PATH → config/db.json 순서로 SQLite DB 경로를 읽습니다.
func LoadDSN(configPath string) (string, error) {
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		return dbPath, nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 설정 파일이 없으면 기본 경로 사용
		execPath := path.ExecutePath("easyPreparation")
		return filepath.Join(execPath, "data", "easyprep.db"), nil
	}
	var cfg struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("db.json 파싱 오류: %v", err)
	}
	if cfg.Path == "" {
		execPath := path.ExecutePath("easyPreparation")
		return filepath.Join(execPath, "data", "easyprep.db"), nil
	}
	execPath := path.ExecutePath("easyPreparation")
	return filepath.Join(execPath, cfg.Path), nil
}

// DB 연결 변수 (전역 또는 의존성 주입으로 관리)
var db *sql.DB

// InitDB initializes the database connection and applies schema if needed
func InitDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")
	db.Exec("PRAGMA busy_timeout=5000")

	// 스키마 자동 초기화 — churches 테이블 존재 여부로 판단
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='churches'").Scan(&name)
	if err != nil {
		execPath := path.ExecutePath("easyPreparation")
		schemaPath := filepath.Join(execPath, "data", "schema.sql")
		schema, readErr := os.ReadFile(schemaPath)
		if readErr != nil {
			log.Printf("[DB] schema.sql 없음: %v (수동 초기화 필요)", readErr)
		} else {
			if _, execErr := db.Exec(string(schema)); execErr != nil {
				log.Printf("[DB] 스키마 초기화 실패: %v", execErr)
			} else {
				log.Println("[DB] 스키마 자동 초기화 완료")
			}
		}
	}

	// 기존 DB 마이그레이션 — 누락된 컬럼 추가 (ALTER TABLE은 이미 존재하면 무시)
	migrations := []string{
		"ALTER TABLE licenses ADD COLUMN last_verified TEXT",
		"ALTER TABLE licenses ADD COLUMN signature TEXT DEFAULT ''",
	}
	for _, m := range migrations {
		if _, execErr := db.Exec(m); execErr != nil {
			// "duplicate column name" 에러는 이미 존재하므로 무시
			if !strings.Contains(execErr.Error(), "duplicate column") {
				log.Printf("[DB] 마이그레이션 실패: %v", execErr)
			}
		}
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// GetDB returns the shared DB connection
func GetDB() *sql.DB {
	return db
}

func ProcessQuote(worshipTitle string, bulletin *[]map[string]interface{}) {
	i := 0
	for i < len(*bulletin) {
		el := (*bulletin)[i]

		title, tOk := el["title"].(string)
		info, iOk := el["info"].(string)
		obj, oOk := el["obj"].(string)

		if !tOk || !oOk {
			// title 또는 obj가 string이 아닐 경우 다음으로
			i++
			continue
		}

		// "b_"로 시작하는 info 필드가 있을 때만 처리
		if iOk && strings.HasPrefix(info, "b_") {
			// "성경봉독" 제목 뒤에 "말씀내용" 항목 삽입
			if title == "성경봉독" {
				newItem := map[string]interface{}{
					"key":   fmt.Sprintf("%d.1", i),
					"title": "말씀내용",
					"info":  "c_edit",
					"obj":   "-",
				}

				*bulletin = append((*bulletin)[:i+1], append([]map[string]interface{}{newItem}, (*bulletin)[i+1:]...)...)
			}

			var sb strings.Builder
			var objRange string

			if strings.Contains(obj, ",") {
				refs := strings.Split(obj, ",")
				for _, qObj := range refs {
					qObj = strings.TrimSpace(qObj)
					parts := strings.SplitN(qObj, "_", 2)
					if len(parts) != 2 {
						continue // 포맷 이상 시 무시
					}
					korName, codeAndRange := parts[0], parts[1]

					// codeAndRange에서 책 번호 추출 (예: "45/8:1" -> "45")
					codeParts := strings.SplitN(codeAndRange, "/", 2)
					if len(codeParts) != 2 {
						continue
					}

					quoteText, err := GetQuote(codeAndRange)
					if err != nil {
						log.Printf("성경 구절 조회 오류 (%s): %v", codeAndRange, err)
						continue
					}
					sb.WriteString(quoteText)
					sb.WriteString("\n")

					objRange += fmt.Sprintf(", %s %s", korName, parser.CompressVerse(codeParts[1]))
				}
			} else {
				parts := strings.SplitN(obj, "_", 2)
				if len(parts) == 2 {
					korName, codeAndRange := parts[0], parts[1]

					// codeAndRange에서 책 번호 추출 (예: "45/8:1" -> "45")
					codeParts := strings.SplitN(codeAndRange, "/", 2)
					if len(codeParts) != 2 {
						continue
					}

					quoteText, err := GetQuote(codeAndRange)
					if err != nil {
						log.Printf("성경 구절 조회 오류 (%s): %v", codeAndRange, err)
					} else {
						sb.WriteString(quoteText)
						objRange = fmt.Sprintf("%s %s", korName, parser.CompressVerse(codeParts[1]))
					}
				}
			}

			objRange = strings.TrimPrefix(objRange, ", ")
			(*bulletin)[i]["contents"] = sb.String()
			(*bulletin)[i]["obj"] = objRange
		}

		// "말씀내용"으로 끝나는 title은 바로 앞 항목의 contents 복사 (범위 검사 포함)
		if strings.HasSuffix(title, "말씀내용") {
			if i-1 >= 0 {
				if prevContents, ok := (*bulletin)[i-1]["contents"]; ok {
					(*bulletin)[i]["contents"] = prevContents
				}
			}
		}

		i++
	}

	execPath := path.ExecutePath("easyPreparation")

	sample, _ := json.MarshalIndent(bulletin, "", "  ")
	_ = utils.CheckDirIs(filepath.Join(execPath, "config"))
	_ = os.WriteFile(filepath.Join(execPath, "config", worshipTitle+".json"), sample, 0644)
}

// GetQuote — 기본 버전(개역개정, versionID=1)으로 성경 구절 조회
func GetQuote(codeAndRange string) (string, error) {
	return GetQuoteWithVersion(codeAndRange, 1)
}

// GetQuoteWithVersion — 지정된 버전으로 성경 구절 조회
func GetQuoteWithVersion(codeAndRange string, versionID int) (string, error) {
	var startChapter, startVerse, endChapter, endVerse int

	// codeAndRange 형태: "45/8:1" 또는 "45/8:1-3"
	referBible := strings.Split(codeAndRange, "/")
	if len(referBible) < 2 {
		return "", fmt.Errorf("잘못된 입력 형식: %s (예: 45/8:1-3)", codeAndRange)
	}

	bookOrder := referBible[0]  // "45" (book_order)
	quoteRange := referBible[1] // "8:1" 또는 "8:1-3"

	// book_order를 정수로 변환
	bookOrderInt, err := strconv.Atoi(bookOrder)
	if err != nil {
		return "", fmt.Errorf("잘못된 책 번호: %s", bookOrder)
	}

	if strings.Contains(quoteRange, "-") {
		qCVs := strings.Split(quoteRange, "-")
		start := strings.Split(qCVs[0], ":")
		end := strings.Split(qCVs[1], ":")

		startChapter, _ = strconv.Atoi(start[0])
		startVerse, _ = strconv.Atoi(start[1])

		if len(end) == 2 {
			// "8:1-10:5" 형태
			endChapter, _ = strconv.Atoi(end[0])
			endVerse, _ = strconv.Atoi(end[1])
		} else {
			// "8:1-5" 형태 (같은 장 내)
			endChapter = startChapter
			endVerse, _ = strconv.Atoi(end[0])
		}
	} else {
		start := strings.Split(quoteRange, ":")
		startChapter, _ = strconv.Atoi(start[0])
		startVerse, _ = strconv.Atoi(start[1])
		endChapter, endVerse = startChapter, startVerse
	}

	versesText, err := getBibleVersesFromDB(versionID, bookOrderInt, startChapter, startVerse, endChapter, endVerse)
	if err != nil {
		return "", fmt.Errorf("성경 구절 가져오기 오류 (%s): %v", codeAndRange, err)
	}

	log.Printf("📖 %s (v%d) 조회 완료 (%d절)", codeAndRange, versionID, len(strings.Split(versesText, "\n")))
	return versesText, nil
}

func getBibleVersesFromDB(versionID, bookOrder, startChapter, startVerse, endChapter, endVerse int) (string, error) {
	if db == nil {
		return "", fmt.Errorf("데이터베이스 연결이 초기화되지 않았습니다")
	}

	query := `
		SELECT v.chapter, v.verse, v.text, b.name_kor
		FROM verses v
		JOIN books b ON v.book_id = b.id
		WHERE v.version_id = ?
		  AND b.book_order = ?
		  AND (
		    (v.chapter > ?) OR
		    (v.chapter = ? AND v.verse >= ?)
		  )
		  AND (
		    (v.chapter < ?) OR
		    (v.chapter = ? AND v.verse <= ?)
		  )
		ORDER BY v.chapter, v.verse
	`

	rows, err := db.Query(query, versionID, bookOrder, startChapter, startChapter, startVerse, endChapter, endChapter, endVerse)
	if err != nil {
		return "", fmt.Errorf("쿼리 실행 오류: %v", err)
	}
	defer rows.Close()

	var result []string
	var bookName string

	for rows.Next() {
		var chapter, verse int
		var text, name string

		err := rows.Scan(&chapter, &verse, &text, &name)
		if err != nil {
			return "", fmt.Errorf("데이터 스캔 오류: %v", err)
		}

		if bookName == "" {
			bookName = name
		}

		result = append(result, fmt.Sprintf("%d:%d %s", chapter, verse, text))
	}

	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("행 반복 오류: %v", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("해당 구절을 찾을 수 없습니다: %s %d:%d", bookName, startChapter, startVerse)
	}

	return strings.Join(result, "\n"), nil
}

// GetBibleVersions 사용 가능한 성경 번역본 목록을 반환
func GetBibleVersions() ([]map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("데이터베이스 연결이 초기화되지 않았습니다")
	}

	query := "SELECT id, name, code FROM bible_versions ORDER BY id"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []map[string]interface{}
	for rows.Next() {
		var id int
		var name, code string

		err := rows.Scan(&id, &name, &code)
		if err != nil {
			return nil, err
		}

		versions = append(versions, map[string]interface{}{
			"id":   id,
			"name": name,
			"code": code,
		})
	}

	return versions, rows.Err()
}

// GetBooks 성경 책 목록을 반환
func GetBooks() ([]map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("데이터베이스 연결이 초기화되지 않았습니다")
	}

	query := `
		SELECT id, name_kor, name_eng, abbr_kor, abbr_eng, book_order 
		FROM books 
		ORDER BY book_order
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []map[string]interface{}
	for rows.Next() {
		var id, bookOrder int
		var nameKor, nameEng, abbrKor, abbrEng string

		err := rows.Scan(&id, &nameKor, &nameEng, &abbrKor, &abbrEng, &bookOrder)
		if err != nil {
			return nil, err
		}

		books = append(books, map[string]interface{}{
			"id":         id,
			"name_kor":   nameKor,
			"name_eng":   nameEng,
			"abbr_kor":   abbrKor,
			"abbr_eng":   abbrEng,
			"book_order": bookOrder,
		})
	}

	return books, rows.Err()
}

// SearchBibleVerses 키워드로 성경 본문 검색
func SearchBibleVerses(keyword string, versionID int, limit int) ([]map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("데이터베이스 연결이 초기화되지 않았습니다")
	}
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT b.name_kor, b.book_order, v.chapter, v.verse, v.text
		FROM verses v
		JOIN books b ON v.book_id = b.id
		WHERE v.version_id = ? AND v.text LIKE '%' || ? || '%'
		ORDER BY b.book_order, v.chapter, v.verse
		LIMIT ?
	`
	rows, err := db.Query(query, versionID, keyword, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var bookName string
		var bookOrder, chapter, verse int
		var text string
		if err := rows.Scan(&bookName, &bookOrder, &chapter, &verse, &text); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"book_name":  bookName,
			"book_order": bookOrder,
			"chapter":    chapter,
			"verse":      verse,
			"text":       text,
		})
	}
	return results, rows.Err()
}

// GetChapterVerses 특정 장의 모든 절 조회
func GetChapterVerses(versionID, bookOrder, chapter int) ([]map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("데이터베이스 연결이 초기화되지 않았습니다")
	}

	query := `
		SELECT v.verse, v.text
		FROM verses v
		JOIN books b ON v.book_id = b.id
		WHERE v.version_id = ? AND b.book_order = ? AND v.chapter = ?
		ORDER BY v.verse
	`
	rows, err := db.Query(query, versionID, bookOrder, chapter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var verse int
		var text string
		if err := rows.Scan(&verse, &text); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"verse": verse,
			"text":  text,
		})
	}
	return results, rows.Err()
}

// GetBookChapterCount 특정 책의 장 수 조회
func GetBookChapterCount(versionID, bookOrder int) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("데이터베이스 연결이 초기화되지 않았습니다")
	}

	query := `
		SELECT MAX(v.chapter)
		FROM verses v
		JOIN books b ON v.book_id = b.id
		WHERE v.version_id = ? AND b.book_order = ?
	`
	var maxChapter int
	err := db.QueryRow(query, versionID, bookOrder).Scan(&maxChapter)
	return maxChapter, err
}

// GetBibleVersesWithVersion 특정 번역본의 성경 구절을 가져오는 함수
func GetBibleVersesWithVersion(versionID, bookOrder, startChapter, startVerse, endChapter, endVerse int) (string, error) {
	if db == nil {
		return "", fmt.Errorf("데이터베이스 연결이 초기화되지 않았습니다")
	}

	query := `
		SELECT v.chapter, v.verse, v.text, b.name_kor
		FROM verses v
		JOIN books b ON v.book_id = b.id
		WHERE v.version_id = ?
		  AND b.book_order = ?
		  AND (
		    (v.chapter > ?) OR
		    (v.chapter = ? AND v.verse >= ?)
		  )
		  AND (
		    (v.chapter < ?) OR
		    (v.chapter = ? AND v.verse <= ?)
		  )
		ORDER BY v.chapter, v.verse
	`

	rows, err := db.Query(query, versionID, bookOrder, startChapter, startChapter, startVerse, endChapter, endChapter, endVerse)
	if err != nil {
		return "", fmt.Errorf("쿼리 실행 오류: %v", err)
	}
	defer rows.Close()

	var result []string
	var bookName string

	for rows.Next() {
		var chapter, verse int
		var text, name string

		err := rows.Scan(&chapter, &verse, &text, &name)
		if err != nil {
			return "", fmt.Errorf("데이터 스캔 오류: %v", err)
		}

		if bookName == "" {
			bookName = name
		}

		result = append(result, fmt.Sprintf("%d:%d %s", chapter, verse, text))
	}

	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("행 반복 오류: %v", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("해당 구절을 찾을 수 없습니다: %s %d:%d", bookName, startChapter, startVerse)
	}

	return strings.Join(result, "\n"), nil
}
