package main

import (
	"bufio"
	"easyPreparation_1.0/internal/db"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 노래 제목 입력받기
	songTitle := getSongTitle()
	if songTitle == "" {
		fmt.Println("노래 제목을 입력하지 않았습니다. 프로그램을 종료합니다.")
		return
	}

	// 노래 목록에 대한 프레젠테이션 생성 및 DB 저장
	createPresentationForSongs(songTitle)

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

func sanitizeFileName(fileName string) string {
	replacer := strings.NewReplacer(
		"<", "_",
		">", "_",
		":", "_",
		`"`, "_",
		"/", "_",
		`\`, "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)

	safeName := replacer.Replace(fileName)
	return strings.TrimSpace(safeName)
}

// 노래 제목에 대한 프레젠테이션 생성 함수
func createPresentationForSongs(songTitle string) {

	songTitles := strings.Split(songTitle, ",")

	for _, title := range songTitles {
		// 가사 검색
		song := &lyrics.SlideData{}
		song.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", title, false)

		outPutDir := "./output/pdf"
		_ = pkg.CheckDirIs(outPutDir)

		fileName := filepath.Join(outPutDir, sanitizeFileName(title)+".pdf")
		execPath, _ := os.Getwd()
		configPath := filepath.Join(execPath, "config/custom.json")
		config := extract.ExtCustomOption(configPath)
		pdfSize := gofpdf.SizeType{
			Wd: config.Size.Background.Presentation.Width,
			Ht: config.Size.Background.Presentation.Height,
		}

		objPdf := presentation.New(pdfSize)
		objPdf.SetText(40, true, color.RGBA{R: 255, G: 255, B: 255})

		for _, content := range song.Content {
			objPdf.AddPage()

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
