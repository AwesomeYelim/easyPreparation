package main

import (
	"bufio"
	"easyPreparation_1.0/internal/db"
	"easyPreparation_1.0/pkg"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Bolthold 데이터베이스 열기
	path := "data/local.db"
	err := pkg.CheckDirIs(filepath.Dir(path)) // 경로가 없다면 생성
	if err != nil {
		fmt.Printf("디렉토리 생성 중 오류 발생: %v\n", err)
		return
	}
	store := db.OpenDB(path)
	defer store.Close()

	// 노래 리스트 출력
	songs := db.PrintSongList(store)
	if len(songs) == 0 {
		fmt.Println("저장 된 노래가 없습니다")
		return
	}
	// 노래를 조회할지 여부 묻기
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("노래를 조회 하시겠습니까? (yes(Enter)/no)")
	response, _ := reader.ReadString('\n')
	response = db.CleanInput(response)

	if response == "yes" || response == "" {
		// 화살표로 노래 선택
		selectedID := db.SelectSong(songs)
		if selectedID == 0 {
			err := db.DeleteAllSongs(store)
			if err == nil {
				fmt.Println("노래 리스트가 삭제되었습니다.")
				return
			}
		}
		if selectedID != -1 {
			song, err := db.GetSongFromDB(store, selectedID)
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
