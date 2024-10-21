package main

import (
	"bufio"
	"easyPreparation_1.0/internal/db"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/presentation"
	"fmt"
	"github.com/timshannon/bolthold"
	"log"
	"os"
	"regexp"
	"strings"
)

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

// 노래 제목에 대한 프레젠테이션 생성 함수
func createPresentationForSongs(songTitles []string) {
	for _, title := range songTitles {
		// 가사 검색
		song := &lyrics.SlideData{}
		song.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", title, false)

		// 파일 이름 만들기
		fileName := sanitizeFileName(title) + ".pdf"

		// PDF 프레젠테이션 생성
		presentation.CreatePresentation(song, fileName)
		fmt.Printf("프레젠테이션이 '%s'에 저장되었습니다.\n", fileName)

		// Bolthold 데이터베이스 열기
		store, err := bolthold.Open("data/local.db", 0666, nil)

		defer store.Close()
		// DB에 노래 저장
		err = db.SaveSongToDB(store, title, song.Content)
		if err != nil {
			log.Printf("노래 저장 실패: %v\n", err)
		} else {
			fmt.Printf("'%s' 노래가 데이터베이스에 저장되었습니다.\n", title)
		}
	}
}

func main() {
	// 노래 제목 입력받기
	songTitle := getSongTitle()
	if songTitle == "" {
		fmt.Println("노래 제목을 입력하지 않았습니다. 프로그램을 종료합니다.")
		return
	}

	// 노래 제목을 쉼표로 구분하여 배열로 변환
	songTitles := strings.Split(songTitle, ",")

	// 노래 목록에 대한 프레젠테이션 생성 및 DB 저장
	createPresentationForSongs(songTitles)

}
