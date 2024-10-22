package main

import (
	"bufio"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/timshannon/bolthold"
	"log"
	"os"
	"strings"
	"time"
)

// Song 구조체 정의 (Bolthold 데이터 저장용)
type Song struct {
	ID      int `boltholdKey:"ID"` // Primary Key 역할
	Title   string
	Content []string
}

func main() {
	// Bolthold 데이터베이스 열기
	store, err := bolthold.Open("data/local.db", 0666, nil)
	if err != nil {
		log.Fatalf("DB 열기 오류: %v", err)
	}
	defer store.Close()

	// 노래 리스트 출력
	songs := printSongList(store)

	// 노래를 조회할지 여부 묻기
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("노래를 조회 하시겠습니까? (yes/no): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response)) // 공백 제거 후 소문자로 변환

	if response == "yes" {
		// 화살표로 노래 선택
		selectedID := selectSong(songs)
		if selectedID != -1 {
			song, err := getSongFromDB(store, selectedID)
			if err != nil {
				fmt.Printf("노래 조회 실패: %v\n", err)
			} else {
				fmt.Printf("조회된 노래: %s - %s\n", song.Title, song.Content)
			}
		} else {
			fmt.Println("노래 선택이 취소되었습니다.")
		}
	} else {
		fmt.Println("노래 조회를 건너뜁니다.")
	}
}

// 노래 리스트를 출력하는 함수
func printSongList(store *bolthold.Store) []Song {
	var songs []Song
	err := store.Find(&songs, nil) // 모든 Song 항목 조회
	if err != nil {
		log.Printf("노래 리스트 조회 실패: %v\n", err)
		return nil
	}

	fmt.Println("저장된 노래 목록:")
	for _, song := range songs {
		fmt.Printf("ID: %d, Title: %s\n", song.ID, song.Title)
	}
	return songs
}

// 화살표로 노래를 선택하는 함수
func selectSong(songs []Song) int {
	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	cursor := 0 // 커서 위치
	selectedID := -1

	for {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		for i, song := range songs {
			if i == cursor {
				termbox.SetCell(0, i, '>', termbox.ColorGreen, termbox.ColorDefault) // 화살표 표시
			}
			printSongDetails(i, song.Title)
		}

		termbox.Flush()

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowDown:
				cursor = (cursor + 1) % len(songs) // 아래로 이동
			case termbox.KeyArrowUp:
				cursor = (cursor - 1 + len(songs)) % len(songs) // 위로 이동
			case termbox.KeyEnter:
				selectedID = songs[cursor].ID // 선택된 노래 ID 저장
				return selectedID
			case termbox.KeyEsc:
				return -1 // 선택 취소
			}
		case termbox.EventError:
			log.Fatal(ev.Err)
		}
		time.Sleep(50 * time.Millisecond) // 50초 제한
	}
}

// 노래 상세정보 출력
func printSongDetails(i int, title string) {
	for j, char := range title {
		termbox.SetCell(j+2, i, char, termbox.ColorDefault, termbox.ColorDefault) // 제목 출력
	}
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
