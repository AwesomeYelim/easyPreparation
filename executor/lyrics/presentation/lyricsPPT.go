package main

import (
	"easyPreparation_1.0/internal/classification"
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
	lpm := NewLyricsPresentationManager()
	defer lpm.Cleanup()
	lyricsInfo, figmaInfo := gui.SetLyricsGui(lpm.execPath)
	figmaInfo.GetFigmaImage(lpm.outputDir, "forLyrics")
	lpm.CreatePresentation(lyricsInfo)
}

func NewLyricsPresentationManager() *LyricsPresentationManager {
	execPath := path.ExecutePath("easyPreparation")
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
	hasLyrics := lyricsInfo["hasLyrics"]

	fontInfo := lpm.config.Classification.Lyrics.Presentation.FontInfo

	labelS, labelH := fontInfo.FontSize/2, 28.00
	labelWm, labelHm := 13.00, 10.00
	labelP := 15.00

	if songTitle == "" {
		fmt.Println("노래 제목을 입력하지 않았습니다. 프로그램을 종료합니다.")
		return
	}

	songTitles := strings.Split(songTitle, ",")
	backgroundImages, _ := os.ReadDir(lpm.outputDir)
	instanceSize := gofpdf.SizeType{
		Wd: extract.ConfigMem.Classification.Lyrics.Presentation.Width,
		Ht: extract.ConfigMem.Classification.Lyrics.Presentation.Height,
	}

	for _, title := range songTitles {
		song := &parser.SlideData{}

		if hasLyrics != "" {
			song.Content = pkg.SplitTwoLines(hasLyrics)
		} else {
			song.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", title, false)
		}

		fileName := filepath.Join(strings.TrimSuffix(lpm.outputDir, "tmp"), sanitize.FileName(title)+".pdf")

		objPdf := presentation.New(instanceSize)
		objPdf.Config = extract.ConfigMem.Classification.Lyrics.Presentation

		for _, content := range song.Content {
			objPdf.AddPage()
			objPdf.CheckImgPlaced(filepath.Join(lpm.outputDir, backgroundImages[0].Name()), 0)
			// 가운데 배치
			objPdf.SetXY((objPdf.Config.Width-objPdf.Config.InnerRectangle.Width)/2, (objPdf.Config.Height-fontInfo.FontSize)/2)
			objPdf.SetText(fontInfo, true, color.RGBA{R: 255, G: 255, B: 255})
			objPdf.MultiCell(objPdf.Config.InnerRectangle.Width, fontInfo.FontSize/2, content, "", "C", false)

			// label - 400*70
			// margin - 20, 15
			textWidth := objPdf.GetStringWidth(label)
			objPdf.SetXY(objPdf.Config.Width-(textWidth+labelWm+labelP), objPdf.Config.Height-(labelH+labelHm+labelP))
			objPdf.SetText(classification.FontInfo{
				FontFamily: "Jacques Francois", FontSize: labelS,
			}, false, color.RGBA{R: 255, G: 255, B: 255})
			objPdf.MultiCell(textWidth, labelH, label, "", "R", false)
		}
		_ = pkg.ReplaceDirPath(fileName, "./")

		if err := objPdf.OutputFileAndClose(fileName); err != nil {
			log.Fatalf("PDF 저장 중 에러 발생: %v", err)
		}
		fmt.Printf("프레젠테이션이 '%s'에 저장되었습니다.\n", fileName)

	}
}
