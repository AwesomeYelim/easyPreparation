package main

import (
	"bufio"
	"easyPreparation_1.0/internal/db"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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

		outPutDir := "./output/pdf"
		pkg.CheckDirIs(outPutDir)

		fileName := filepath.Join(outPutDir, sanitizeFileName(title)+".pdf")

		pdfSize := gofpdf.SizeType{
			Wd: 297.0,
			Ht: 167.0,
		}
		objPdf := presentation.New(pdfSize)
		objPdf.SetText(40, color.RGBA{R: 255, G: 255, B: 255})

		for _, content := range song.Content {
			objPdf.AddPage()
			// 배경 이미지 추가 (배경 이미지를 추가하려면 파일이 필요합니다)
			backgroundImage := "./public/images/ppt_background.png"
			objPdf.CheckImgPlaced(pdfSize, backgroundImage, 0)

			var textW float64 = 250
			var textH float64 = 20
			// 텍스트 추가 (내용 설정)
			objPdf.SetXY((pdfSize.Wd-textW)/2, (pdfSize.Ht-textH)/2)
			objPdf.MultiCell(textW, textH, content, "", "C", false)
		}
		err := objPdf.OutputFileAndClose(fileName)
		if err != nil {
			log.Fatalf("PDF 저장 중 에러 발생: %v", err)
		}

		fmt.Printf("프레젠테이션이 '%s'에 저장되었습니다.\n", fileName)

		path := "data/local.db"
		err = pkg.CheckDirIs(filepath.Dir(path)) // 경로가 없다면 생성
		if err != nil {
			fmt.Printf("디렉토리 생성 중 오류 발생: %v\n", err)
			return
		}

		store := db.OpenDB(path)
		defer store.Close()

		// DB에 노래 저장
		song.Title = title
		err = db.SaveSongToDB(store, song)
		if err != nil {
			log.Printf("노래 저장 실패: %v\n", err)
		} else {
			fmt.Printf("'%s' 노래가 데이터베이스에 저장되었습니다.\n", title)
		}
	}
}
