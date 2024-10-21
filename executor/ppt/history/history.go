package main

import (
	"easyPreparation_1.0/internal/db"
	"fmt"
	"github.com/timshannon/bolthold"
	"log"
)

func main() {
	// Bolthold 데이터베이스 열기
	store, err := bolthold.Open("local.db", 0666, nil)
	if err != nil {
		log.Fatalf("DB 열기 오류: %v", err)
	}
	defer store.Close()

	// 노래 리스트 출력
	printSongList(store)

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

// 노래 리스트를 출력하는 함수
func printSongList(store *bolthold.Store) {
	var songs []db.Song
	err := store.Find(&songs, nil) // 모든 Song 항목 조회
	if err != nil {
		log.Printf("노래 리스트 조회 실패: %v\n", err)
		return
	}

	fmt.Println("저장된 노래 목록:")
	for _, song := range songs {
		fmt.Printf("ID: %d, Title: %s\n", song.ID, song.Title)
	}

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

// 노래 데이터 조회 함수
func getSongFromDB(store *bolthold.Store, id int) (*db.Song, error) {
	var song db.Song
	err := store.Get(id, &song)
	if err != nil {
		return nil, err
	}
	return &song, nil
}
