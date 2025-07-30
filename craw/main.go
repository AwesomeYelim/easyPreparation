package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// 구조체 정의
type Verse struct {
	Verse int    `json:"verse"`
	Text  string `json:"text"`
}

type BibleAPIResponse struct {
	Data struct {
		Testament   string `json:"testament"`
		Bookname    string `json:"bookname"`
		BooknameAbb string `json:"bookname_abb"`
		Data        struct {
			Version1 struct {
				Version        int    `json:"version"`
				Jang           int    `json:"jang"`
				VersionName    string `json:"version_name"`
				SoundtrackYn   string `json:"soundtrack_yn"`
				TranslationIdx int    `json:"translation_idx"`
				Bookname       string `json:"bookname"`
				BooknameAbb    string `json:"bookname_abb"`
				Theme          []struct {
					StartJul int    `json:"start_jul"`
					Cont     string `json:"cont"`
				} `json:"theme"`
				Content []struct {
					Jul            int    `json:"jul"`
					Text           string `json:"text"`
					DictionaryList []struct {
						Idx   int    `json:"idx"`
						Word  string `json:"word"`
						Word2 string `json:"word2"`
						Cont  string `json:"cont"`
					} `json:"dictionaryList"`
				} `json:"content"`
			} `json:"version1"`
		} `json:"data"`
	} `json:"data"`
}

// 책 정보 구조체 (필요 시 JSON으로 불러와도 됨)
type BibleBook struct {
	Index    int    `json:"index"`
	Kor      string `json:"kor"`
	Eng      string `json:"eng"`
	Chapters []int  `json:"chapters"`
}

// DB 설정
const (
	DB_HOST     = "138.2.119.220"
	DB_PORT     = 5432
	DB_USER     = "postgres"
	DB_PASSWORD = "02031122"
	DB_NAME     = "postgres"
)

func main() {
	// 연결
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bibleBooks, err := LoadBibleBooksFromJSON("bible_info.json")
	if err != nil {
		log.Fatalf("책 정보 로딩 실패: %v", err)
	}

	versionID := 1 // 개역개정

	for _, book := range bibleBooks {
		for chapter := 1; chapter <= len(book.Chapters); chapter++ {
			// API 요청 URL
			url := fmt.Sprintf("https://goodtvbible.goodtv.co.kr/api/onlinebible/bibleread/read-all?version1=0&version2=&version3=&bible_code=%d&jang=%d", book.Index, chapter)

			// API 호출
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("❌ 요청 실패 (%s %d장): %v", book.Kor, chapter, err)
				continue
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("❌ 응답 읽기 실패: %v", err)
				continue
			}

			// JSON 파싱
			var res BibleAPIResponse

			if err := json.Unmarshal(body, &res); err != nil {
				log.Printf("❌ JSON 파싱 실패: %v\n본문: %s", err, string(body))
				continue
			}

			contents := res.Data.Data.Version1.Content

			// 절 단위로 INSERT
			for _, verse := range contents {
				_, err := db.Exec(`
					INSERT INTO verses (version_id, book_id, chapter, verse, text)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT DO NOTHING;
				`, versionID, book.Index, chapter, verse.Jul, verse.Text)

				if err != nil {
					log.Printf("❌ DB 저장 실패 (%s %d:%d): %v", book.Kor, chapter, verse.Jul, err)
				}
			}

			log.Printf("✅ 저장 완료: %s %d장", book.Kor, chapter)
		}
	}
}

func LoadBibleBooksFromJSON(path string) (map[string]BibleBook, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var books map[string]BibleBook
	err = json.Unmarshal(bytes, &books)
	if err != nil {
		return nil, err
	}

	return books, nil
}
