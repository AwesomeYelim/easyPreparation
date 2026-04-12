package lyrics

import (
	"easyPreparation_1.0/internal/classification"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/parser"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/internal/sanitize"
	"easyPreparation_1.0/internal/utils"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type LyricsPresentationManager struct {
	ExecPath  string
	Config    extract.Config
	OutputDir string
}

type stSong struct {
	title  string
	lyrics string
}

func CreateLyricsPDF(data map[string]interface{}) {
	lpm := NewLyricsPresentationManager()

	// 배경 이미지가 OutputDir에 없으면 경고만 출력하고 계속 진행
	entries, err := os.ReadDir(lpm.OutputDir)
	if err != nil || len(entries) == 0 {
		handlers.BroadcastProgress("Background Warning", 1, "배경 이미지 없음 — 템플릿 폴더를 확인하세요")
	}

	lpm.CreatePresentation(data)
}

func NewLyricsPresentationManager() *LyricsPresentationManager {
	execPath := path.ExecutePath("easyPreparation")
	configPath := filepath.Join(execPath, "config/custom.json")
	extract.ExtCustomOption(configPath)

	outputDir := filepath.Join(execPath, "data", "templates", "lyrics")
	_ = utils.CheckDirIs(outputDir)

	return &LyricsPresentationManager{
		ExecPath:  execPath,
		Config:    extract.ConfigMem,
		OutputDir: outputDir,
	}
}

func (lpm *LyricsPresentationManager) Cleanup() {
	_ = os.RemoveAll(lpm.OutputDir)
}

func (lpm *LyricsPresentationManager) CreatePresentation(data map[string]interface{}) {

	rawSongs, ok := data["songs"].([]interface{})
	if !ok {
		log.Println("songs 데이터가 유효하지 않습니다.")
		return
	}

	var songs []stSong

	for _, item := range rawSongs {
		songMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := songMap["title"].(string)
		lyricsText, _ := songMap["lyrics"].(string)

		songs = append(songs, stSong{
			title:  title,
			lyrics: lyricsText,
		})
	}
	label, _ := data["mark"].(string)

	fontInfo := lpm.Config.Classification.Lyrics.Presentation.FontInfo

	labelS, labelH := fontInfo.FontSize/2, 28.00
	labelWm, labelHm := 13.00, 10.00
	labelP := 15.00

	backgroundImages, err := os.ReadDir(lpm.OutputDir)
	if err != nil || len(backgroundImages) == 0 {
		handlers.BroadcastProgress("Lyrics Error", -1, fmt.Sprintf("배경 이미지 없음: %s", lpm.OutputDir))
		return
	}
	// 이미지 파일만 필터 (PNG/JPG)
	var bgImagePath string
	for _, e := range backgroundImages {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
			bgImagePath = filepath.Join(lpm.OutputDir, e.Name())
			break
		}
	}
	if bgImagePath == "" {
		handlers.BroadcastProgress("Lyrics Error", -1, fmt.Sprintf("배경 이미지 없음: %s", lpm.OutputDir))
		return
	}

	instanceSize := gofpdf.SizeType{
		Wd: extract.ConfigMem.Classification.Lyrics.Presentation.Width,
		Ht: extract.ConfigMem.Classification.Lyrics.Presentation.Height,
	}

	for _, song := range songs {
		newSong := &parser.SlideData{}
		newSong.Title = song.title
		newSong.Content = utils.SplitTwoLines(song.lyrics)

		if len(newSong.Content) == 0 {
			log.Printf("가사 없음, 스킵: %s", song.title)
			continue
		}

		fileName := filepath.Join(strings.TrimSuffix(lpm.OutputDir, "tmp"), sanitize.FileName(song.title)+".pdf")

		objPdf := presentation.New(instanceSize)
		objPdf.Config = extract.ConfigMem.Classification.Lyrics.Presentation

		for _, content := range newSong.Content {
			objPdf.AddPage()
			objPdf.CheckImgPlaced(bgImagePath, 0)
			// 가운데 배치
			objPdf.SetXY((objPdf.Config.Width-objPdf.Config.InnerRectangle.Width)/2, (objPdf.Config.Height-fontInfo.FontSize)/2)
			objPdf.SetText(fontInfo, true, color.RGBA{R: 255, G: 255, B: 255})
			objPdf.MultiCell(objPdf.Config.InnerRectangle.Width, fontInfo.FontSize/2, content, "", "C", false)

			// label
			textWidth := objPdf.GetStringWidth(label)
			objPdf.SetXY(objPdf.Config.Width-(textWidth+labelWm+labelP), objPdf.Config.Height-(labelH+labelHm+labelP))
			objPdf.SetText(classification.FontInfo{
				FontFamily: "Jacques Francois", FontSize: labelS,
			}, false, color.RGBA{R: 255, G: 255, B: 255})
			objPdf.MultiCell(textWidth, labelH, label, "", "R", false)
		}
		_ = utils.ReplaceDirPath(fileName, "./")

		if err := objPdf.OutputFileAndClose(fileName); err != nil {
			log.Printf("PDF 저장 중 에러 발생: %v", err)
			continue
		}
		fmt.Printf("프레젠테이션이 '%s'에 저장되었습니다.\n", fileName)
	}
}
