// +build ignore

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	_ "modernc.org/sqlite"
)

type BibleBook struct {
	Index    int    `json:"index"`
	Kor      string `json:"kor"`
	Eng      string `json:"eng"`
	Chapters []int  `json:"chapters"`
}

var engFullNames = map[int]string{
	1: "Genesis", 2: "Exodus", 3: "Leviticus", 4: "Numbers", 5: "Deuteronomy",
	6: "Joshua", 7: "Judges", 8: "Ruth", 9: "1 Samuel", 10: "2 Samuel",
	11: "1 Kings", 12: "2 Kings", 13: "1 Chronicles", 14: "2 Chronicles",
	15: "Ezra", 16: "Nehemiah", 17: "Esther", 18: "Job", 19: "Psalms",
	20: "Proverbs", 21: "Ecclesiastes", 22: "Song of Solomon", 23: "Isaiah",
	24: "Jeremiah", 25: "Lamentations", 26: "Ezekiel", 27: "Daniel",
	28: "Hosea", 29: "Joel", 30: "Amos", 31: "Obadiah", 32: "Jonah",
	33: "Micah", 34: "Nahum", 35: "Habakkuk", 36: "Zephaniah", 37: "Haggai",
	38: "Zechariah", 39: "Malachi", 40: "Matthew", 41: "Mark", 42: "Luke",
	43: "John", 44: "Acts", 45: "Romans", 46: "1 Corinthians",
	47: "2 Corinthians", 48: "Galatians", 49: "Ephesians", 50: "Philippians",
	51: "Colossians", 52: "1 Thessalonians", 53: "2 Thessalonians",
	54: "1 Timothy", 55: "2 Timothy", 56: "Titus", 57: "Philemon",
	58: "Hebrews", 59: "James", 60: "1 Peter", 61: "2 Peter",
	62: "1 John", 63: "2 John", 64: "3 John", 65: "Jude", 66: "Revelation",
}

func main() {
	dbPath := "data/easyprep.db"
	if p := os.Getenv("DB_PATH"); p != "" {
		dbPath = p
	}

	bibleInfoPath := "config/bible_info.json"
	if len(os.Args) > 1 {
		bibleInfoPath = os.Args[1]
	}

	// Read bible_info.json
	data, err := os.ReadFile(bibleInfoPath)
	if err != nil {
		log.Fatalf("bible_info.json 읽기 실패: %v", err)
	}

	var books map[string]BibleBook
	if err := json.Unmarshal(data, &books); err != nil {
		log.Fatalf("JSON 파싱 실패: %v", err)
	}

	// Sort by index
	type namedBook struct {
		Name string
		BibleBook
	}
	var sorted []namedBook
	for name, book := range books {
		sorted = append(sorted, namedBook{name, book})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Index < sorted[j].Index })

	// Open DB
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("DB 열기 실패: %v", err)
	}
	defer db.Close()
	db.Exec("PRAGMA journal_mode=WAL")

	// Insert books
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("INSERT OR IGNORE INTO books (id, name_kor, name_eng, abbr_kor, abbr_eng, book_order) VALUES (?, ?, ?, ?, ?, ?)")
	defer stmt.Close()

	for _, b := range sorted {
		engName := engFullNames[b.Index]
		if engName == "" {
			engName = b.Eng
		}
		_, err := stmt.Exec(b.Index, b.Name, engName, b.Kor, b.Eng, b.Index)
		if err != nil {
			log.Printf("  [SKIP] %s: %v", b.Name, err)
		} else {
			log.Printf("  [OK] %d. %s (%s)", b.Index, b.Name, engName)
		}
	}
	tx.Commit()

	// Verify
	var count int
	db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	fmt.Printf("\n=== books 테이블: %d권 시드 완료 ===\n", count)
}
