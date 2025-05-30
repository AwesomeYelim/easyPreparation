package db

import (
	"easyPreparation_1.0/internal/parser"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/timshannon/bolthold"
	"log"
	"strings"
)

// 터미널에서 항목 선택 함수
func SelectSong(songs []parser.SlideData) int {
	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	songs = append(songs, parser.SlideData{TrackID: 0, Title: "전체 삭제", Content: make([]string, 0)})
	cursor := 0
	for {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		for i, song := range songs {
			if i == cursor {
				termbox.SetCell(0, i, '>', termbox.ColorGreen, termbox.ColorDefault) // 화살표 표시
				SetCellWithColor(i, song.Title, termbox.ColorGreen, termbox.ColorDefault)
			} else {
				SetCellWithColor(i, song.Title, termbox.ColorDefault, termbox.ColorDefault)
			}
		}
		termbox.Flush()

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowDown:
				cursor = (cursor + 1) % len(songs)
			case termbox.KeyArrowUp:
				cursor = (cursor - 1 + len(songs)) % len(songs)
			case termbox.KeyEnter:
				return songs[cursor].TrackID
			case termbox.KeyEsc:
				return -1
			}
		case termbox.EventError:
			log.Fatal(ev.Err)
		}
	}
}

// 셀에 텍스트와 색상을 설정하는 함수
func SetCellWithColor(i int, text string, fg, bg termbox.Attribute) {
	for j, char := range text {
		termbox.SetCell(j+2, i, char, fg, bg)
	}
}

// 입력 문자열 정리 함수
func CleanInput(input string) string {
	return strings.TrimSpace(strings.ToLower(input))
}

// 노래 리스트 출력 함수
func PrintSongList(store *bolthold.Store) []parser.SlideData {
	var songs []parser.SlideData
	if err := store.Find(&songs, nil); err != nil {
		log.Printf("노래 리스트 조회 실패: %v\n", err)
		return nil
	}
	if len(songs) > 0 {
		fmt.Println("저장된 노래 목록:")
		for _, song := range songs {
			fmt.Printf("ID: %d, Title: %s\n", song.TrackID, song.Title)
		}
	}

	return songs
}

// 노래 데이터 조회 함수
func GetSongFromDB(store *bolthold.Store, id int) (*parser.SlideData, error) {
	var song parser.SlideData
	if err := store.Get(id, &song); err != nil {
		return nil, err
	}
	return &song, nil
}

func DeleteAllSongs(store *bolthold.Store) error {
	var allSongs []parser.SlideData

	// 모든 항목 조회
	err := store.Find(&allSongs, nil)
	if err != nil {
		return fmt.Errorf("삭제할 모든 노래 조회 실패: %v", err)
	}

	// 각 항목 삭제
	for _, song := range allSongs {
		err := store.Delete(song.TrackID, &parser.SlideData{})
		if err != nil {
			return fmt.Errorf("노래 삭제 실패: %v", err)
		}
	}

	return nil
}
