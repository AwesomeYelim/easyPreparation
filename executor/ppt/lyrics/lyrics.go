package main

import (
	"bufio"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/presentation"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/timshannon/bolthold"
)

// Song 구조체 정의 (Bolthold 데이터 저장용)
type Song struct {
	ID      int `boltholdKey:"ID"` // Primary Key 역할
	Title   string
	Content []string
}

// 사용자로부터 노래 제목을 입력받는 함수
func getSongTitle() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("노래 제목을 입력하세요: ")
	songTitle, err := reader.ReadString('\n') // Enter 키 입력 시까지 읽음
	if err != nil {
		fmt.Printf("입력 에러: %v\n", err)
		return ""
	}
	return strings.TrimSpace(songTitle) // 개행문자 및 공백 제거
}

// 파일 이름으로 안전하게 변환하는 함수
func sanitizeFileName(fileName string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*]+`)      // 파일 이름에 사용할 수 없는 문자 정규 표현식
	safeName := re.ReplaceAllString(fileName, "_") // 안전한 문자로 대체
	return strings.TrimSpace(safeName)             // 공백 제거
}

// Bolthold 데이터베이스에 노래 저장 함수
func saveSongToDB(store *bolthold.Store, title string, content []string) error {
	var existingSong Song
	// 제목으로 기존 노래 조회
	err := store.FindOne(&existingSong, bolthold.Where("Title").Eq(title))
	if err == nil {
		// 기존 노래가 있는 경우, 내용을 업데이트
		existingSong.Content = append(existingSong.Content, content...) // 기존 내용에 추가
		return store.Update(existingSong.ID, &existingSong)             // 업데이트
	}

	// 새로운 노래 저장
	// 현재 데이터베이스의 항목 수를 기반으로 새로운 ID 생성
	count, err := store.Count(&Song{}, nil) // Song 타입의 항목 수를 계산
	if err != nil {
		return fmt.Errorf("ID 생성 실패: %v", err)
	}
	newID := count + 1 // ID는 현재 항목 수 + 1

	song := Song{ID: newID, Title: title, Content: content}
	return store.Insert(newID, &song)
}

// 노래 데이터 조회 함수
func getSongFromDB(store *bolthold.Store, id int) (*Song, error) {
	var song Song
	err := store.Get(id, &song)
	if err != nil {
		return nil, err
	}
	return &song, nil
}

// 노래 제목에 대한 프레젠테이션 생성 함수
func createPresentationForSongs(songTitles []string, store *bolthold.Store) {
	for _, title := range songTitles {
		// 가사 검색
		song := &lyrics.SlideData{}
		song.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", title, false)

		// 파일 이름 만들기
		fileName := sanitizeFileName(title) + ".pdf"

		// PDF 프레젠테이션 생성
		presentation.CreatePresentation(song, fileName)
		fmt.Printf("프레젠테이션이 '%s'에 저장되었습니다.\n", fileName)

		// DB에 노래 저장
		err := saveSongToDB(store, title, song.Content)
		if err != nil {
			log.Printf("노래 저장 실패: %v\n", err)
		} else {
			fmt.Printf("'%s' 노래가 데이터베이스에 저장되었습니다.\n", title)
		}
	}
}

// 노래 리스트를 출력하는 함수
func printSongList(store *bolthold.Store) {
	var songs []Song
	err := store.Find(&songs, nil) // 모든 Song 항목 조회
	if err != nil {
		log.Printf("노래 리스트 조회 실패: %v\n", err)
		return
	}

	fmt.Println("저장된 노래 목록:")
	for _, song := range songs {
		fmt.Printf("ID: %d, Title: %s\n", song.ID, song.Title)
	}
}

func main() {
	// Bolthold 데이터베이스 열기
	store, err := bolthold.Open("local.db", 0666, nil)
	if err != nil {
		log.Fatalf("DB 열기 오류: %v", err)
	}
	defer store.Close()

	// 노래 리스트 출력
	printSongList(store) // 추가된 부분

	// 노래 제목 입력받기
	songTitle := getSongTitle()
	if songTitle == "" {
		fmt.Println("노래 제목을 입력하지 않았습니다. 프로그램을 종료합니다.")
		return
	}

	// 노래 제목을 쉼표로 구분하여 배열로 변환
	songTitles := strings.Split(songTitle, ",")

	// 노래 목록에 대한 프레젠테이션 생성 및 DB 저장
	createPresentationForSongs(songTitles, store)

	// 데이터베이스에서 노래 조회 예제
	fmt.Print("조회할 노래 ID를 입력하세요: ")
	var id int
	fmt.Scan(&id)
	song, err := getSongFromDB(store, id)
	if err != nil {
		fmt.Printf("노래 조회 실패: %v\n", err)
	} else {
		fmt.Printf("조회된 노래: %s - %s\n", song.Title, song.Content)
	}
}
