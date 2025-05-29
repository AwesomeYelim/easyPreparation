package presentation

import (
	"easyPreparation_1.0/internal/classification"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma"
	"easyPreparation_1.0/internal/handlers"
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
	ExecPath  string
	Config    extract.Config
	OutputDir string
}

func CreateLyricsPDF(data map[string]interface{}) {
	execPath := path.ExecutePath("easyPreparation")

	lpm := NewLyricsPresentationManager()
	defer lpm.Cleanup()

	var key, token string
	if rawFigma, ok := data["figmaInfo"]; ok {
		if figmaMap, ok := rawFigma.(map[string]interface{}); ok {
			if k, ok := figmaMap["key"].(string); ok {
				key = k
			}
			if t, ok := figmaMap["token"].(string); ok {
				token = t
			}
		}
	}

	figmaInfo := figma.New(&token, &key, execPath)
	if err := figmaInfo.GetNodes(); err != nil {
		handlers.BroadcastProgress("Get Nodes Error", -1, fmt.Sprintf("GetNodes Error: %s", err))
		return
	}

	figmaInfo.GetFigmaImage(lpm.OutputDir, "forLyrics")

	lpm.CreatePresentation(data)
}

func NewLyricsPresentationManager() *LyricsPresentationManager {
	execPath := path.ExecutePath("easyPreparation")
	configPath := filepath.Join(execPath, "config/custom.json")
	extract.ExtCustomOption(configPath)

	outputDir := filepath.Join(execPath, extract.ConfigMem.OutputPath.Lyrics, "tmp")
	_ = pkg.CheckDirIs(outputDir)

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

	var songs []struct {
		Title  string
		Lyrics string
	}

	for _, item := range rawSongs {
		songMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := songMap["title"].(string)
		lyrics, _ := songMap["lyrics"].(string)

		songs = append(songs, struct {
			Title  string
			Lyrics string
		}{
			Title:  title,
			Lyrics: lyrics,
		})
	}
	label := data["mark"].(string)

	fontInfo := lpm.Config.Classification.Lyrics.Presentation.FontInfo

	labelS, labelH := fontInfo.FontSize/2, 28.00
	labelWm, labelHm := 13.00, 10.00
	labelP := 15.00

	//songTitles := strings.Split(songs, ",")
	backgroundImages, _ := os.ReadDir(lpm.OutputDir)
	instanceSize := gofpdf.SizeType{
		Wd: extract.ConfigMem.Classification.Lyrics.Presentation.Width,
		Ht: extract.ConfigMem.Classification.Lyrics.Presentation.Height,
	}

	for _, song := range songs {
		newSong := &parser.SlideData{}
		newSong.Title = song.Title
		newSong.Content = pkg.SplitTwoLines(song.Lyrics)

		//newSong.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", song.title, false)

		fileName := filepath.Join(strings.TrimSuffix(lpm.OutputDir, "tmp"), sanitize.FileName(song.Title)+".pdf")

		objPdf := presentation.New(instanceSize)
		objPdf.Config = extract.ConfigMem.Classification.Lyrics.Presentation

		for _, content := range newSong.Content {
			objPdf.AddPage()
			objPdf.CheckImgPlaced(filepath.Join(lpm.OutputDir, backgroundImages[0].Name()), 0)
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
