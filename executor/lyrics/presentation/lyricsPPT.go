package main

import (
	"easyPreparation_1.0/internal/db"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/parser"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/internal/sanitize"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type LyricsPresentationManager struct {
	execPath  string
	config    extract.Config
	outputDir string
}

func main() {
	execPath, _ := os.Getwd()
	lpm := NewLyricsPresentationManager(execPath)
	defer lpm.Cleanup()

	lyricsInfo, figmaInfo := gui.SetLyricsGui(lpm.execPath)

	figmaInfo.GetFigmaImage(lpm.outputDir, "forLyrics")

	lpm.CreatePresentation(lyricsInfo)
}

func NewLyricsPresentationManager(execPath string) *LyricsPresentationManager {
	execPath = path.ExecutePath(execPath, "easyPreparation")
	configPath := filepath.Join(execPath, "config/custom.json")
	extract.ExtCustomOption(configPath)

	outputDir := filepath.Join(execPath, extract.ConfigMem.OutputPath.Lyrics, "tmp")
	_ = pkg.CheckDirIs(outputDir)

	return &LyricsPresentationManager{
		execPath:  execPath,
		config:    extract.ConfigMem,
		outputDir: outputDir,
	}
}

func (lpm *LyricsPresentationManager) Cleanup() {
	_ = os.RemoveAll(lpm.outputDir)
}

func (lpm *LyricsPresentationManager) CreatePresentation(lyricsInfo map[string]string) {
	songTitle := lyricsInfo["songTitle"]
	label := lyricsInfo["label"]

	fontSize := lpm.config.Classification.Lyrics.Presentation.FontSize + 30
	fontOption := lpm.config.Classification.Bulletin.Print.FontOption

	labelS, labelH := fontSize/2, 28.00
	labelWm, labelHm := 13.00, 10.00
	labelP := 15.00

	if songTitle == "" {
		fmt.Println("노래 제목을 입력하지 않았습니다. 프로그램을 종료합니다.")
		return
	}

	songTitles := strings.Split(songTitle, ",")

	pdfSize, rectangle := getSize(extract.ConfigMem)

	backgroundImages, _ := os.ReadDir(lpm.outputDir)

	for _, title := range songTitles {
		song := &parser.SlideData{}
		song.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", title, false)

		fileName := filepath.Join(strings.TrimSuffix(lpm.outputDir, "tmp"), sanitize.FileName(title)+".pdf")

		objPdf := presentation.New(pdfSize)
		for _, content := range song.Content {
			objPdf.AddPage()
			objPdf.CheckImgPlaced(pdfSize, filepath.Join(lpm.outputDir, backgroundImages[0].Name()), 0)
			// 가운데 배치
			objPdf.SetXY((pdfSize.Wd-rectangle.Width)/2, (pdfSize.Ht-fontSize)/2)
			objPdf.SetText(fontOption, fontSize, true, color.RGBA{R: 255, G: 255, B: 255})
			objPdf.MultiCell(lpm.config.Classification.Lyrics.Presentation.InnerRectangle.Width, fontSize/2, content, "", "C", false)

			// label - 400*70
			// margin - 20, 15
			textWidth := objPdf.GetStringWidth(label)
			objPdf.SetXY(pdfSize.Wd-(textWidth+labelWm+labelP), pdfSize.Ht-(labelH+labelHm+labelP))
			objPdf.SetText("Jacques Francois", labelS, false, color.RGBA{R: 255, G: 255, B: 255})
			objPdf.MultiCell(textWidth, labelH, label, "", "R", false)
		}
		_ = pkg.ReplaceDirPath(fileName, "./")

		if err := objPdf.OutputFileAndClose(fileName); err != nil {
			log.Fatalf("PDF 저장 중 에러 발생: %v", err)
		}
		fmt.Printf("프레젠테이션이 '%s'에 저장되었습니다.\n", fileName)

		lpm.saveToDB(title, song)
	}
}

func (lpm *LyricsPresentationManager) saveToDB(title string, song *parser.SlideData) {
	dbPath := "data/local.db"
	if err := pkg.CheckDirIs(filepath.Dir(dbPath)); err != nil {
		fmt.Printf("디렉토리 생성 중 오류 발생: %v\n", err)
		return
	}

	store := db.OpenDB(dbPath)
	defer func() {
		_ = store.Close()
	}()

	song.Title = title
	if err := db.SaveSongToDB(store, song); err != nil {
		log.Printf("노래 저장 실패: %v\n", err)
	} else {
		fmt.Printf("'%s' 노래가 데이터베이스에 저장되었습니다.\n", title)
	}
}

func getSize(config extract.Config) (gofpdf.SizeType, presentation.Size) {
	bulletinSize := gofpdf.SizeType{
		Wd: config.Classification.Lyrics.Presentation.Width,
		Ht: config.Classification.Lyrics.Presentation.Height,
	}
	rectangle := presentation.Size{
		Width:  config.Classification.Lyrics.Presentation.InnerRectangle.Width,
		Height: config.Classification.Lyrics.Presentation.InnerRectangle.Height,
	}

	return bulletinSize, rectangle
}
