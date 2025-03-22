package forPrint

import (
	"easyPreparation_1.0/internal/classification"
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/internal/sorted"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"os"
	"path/filepath"
	"strings"
)

func CreatePrint(figmaInfo *get.Info, target, execPath string) {
	config := extract.ConfigMem
	outputDir := filepath.Join(execPath, config.OutputPath.Bulletin, "print", "tmp")
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	figmaInfo.GetFigmaImage(outputDir, "forPrint")

	files, _ := os.ReadDir(outputDir)
	instanceSize := gofpdf.SizeType{
		Wd: config.Classification.Bulletin.Print.Width,
		Ht: config.Classification.Bulletin.Print.Height,
	}
	objPdf := presentation.New(instanceSize)
	objPdf.Config = config.Classification.Bulletin.Print
	yearMonth, weekFormatted := date.SetDateTitle()
	objPdf.GetConversionRatio()
	// 파일명 생성: "202411_3.pdf"
	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	sorted.ToIntSort(files, "- ", ".png", 0)

	var elements []gui.WorshipInfo
	var newsCon gui.WorshipInfo

	worshipContents, err := os.ReadFile(filepath.Join(execPath, "config", target+".json"))
	err = json.Unmarshal(worshipContents, &elements)
	fontInfo := config.Classification.Bulletin.Print.FontInfo
	hLColor := colorPalette.HexToRGBA(objPdf.Config.Color.LineColor)

	for _, el := range elements {
		if strings.Contains(el.Title, "교회소식") {
			newsCon = el
			break
		}
	}
	for i, file := range files {
		imgPath := filepath.Join(outputDir, file.Name())

		objPdf.AddPage()
		objPdf.CheckImgPlaced(imgPath, 0)
		if i == 0 {
			sunDateText := date.SetThisSunDay()
			objPdf.SetText(classification.FontInfo{
				FontSize:   fontInfo.FontSize * 1.4,
				FontFamily: fontInfo.FontFamily,
			}, false, colorPalette.HexToRGBA(objPdf.Config.Color.DateColor))
			objPdf.WriteText(sunDateText, "right")
			objPdf.DrawChurchNews(config.Classification.Bulletin.Print.FontInfo, newsCon, hLColor, 70.0, 200.0)
		} else {
			ym := objPdf.ForComposeBuiltin(elements)
			objPdf.ForReferNext(elements, ym)
			objPdf.ForTodayVerse(elements[len(elements)-1])
		}

	}
	outputBtPath := filepath.Join(execPath, config.OutputPath.Bulletin, "print")

	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, outputFilename)
	err = objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}
}
