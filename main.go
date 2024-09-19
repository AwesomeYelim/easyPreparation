package main

import (
	"bufio"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/presentation"
	"fmt"
	"os"

	"github.com/unidoc/unioffice/common/license"
)

// 라이센스 설정 함수
func setUniOfficeLicense() {
	licenseKey := os.Getenv("UNIDOC_LICENSE_API_KEY")
	if err := license.SetMeteredKey(licenseKey); err != nil {
		panic(fmt.Sprintf("UniOffice 라이센스 설정 실패: %v", err))
	}
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
	// 입력받은 제목에서 마지막 개행문자를 제거
	return songTitle[:len(songTitle)-1]
}

func main() {
	// 라이센스 설정
	setUniOfficeLicense()

	// 노래 제목 입력받기
	songTitle := getSongTitle()
	if songTitle == "" {
		fmt.Println("노래 제목을 입력하지 않았습니다. 프로그램을 종료합니다.")
		return
	}

	// 가사 검색
	song := &lyrics.SlideData{}
	song.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", songTitle, false)

	// 프레젠테이션 생성 및 슬라이드 추가
	presentation.CreatePresentation(song, "output.pptx")

	fmt.Println("프레젠테이션이 'output.pptx'에 저장되었습니다.")
}
