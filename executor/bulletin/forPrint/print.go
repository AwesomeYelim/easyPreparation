package forPrint

import (
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
)

func CreatePrint(figmaInfo *get.Info, target, execPath string) {
	config := extract.ConfigMem
	outputDir := filepath.Join(execPath, config.OutputPath.Bulletin, "print", "tmp")
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	figmaInfo.GetFigmaImage(outputDir, "forPrint")

	bulletinPrintSize, rectangle := getSize(config)
	files, _ := os.ReadDir(outputDir)
	objPdf := presentation.New(bulletinPrintSize)
	objPdf.FullSize = bulletinPrintSize
	objPdf.BoxSize = rectangle
	objPdf.Config = config.Classification.Bulletin.Print

	yearMonth, weekFormatted := date.SetDateTitle()

	// 파일명 생성: "202411_3.pdf"
	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	sorted.ToIntSort(files, "- ", ".png", 0)

	var elements []gui.WorshipInfo

	worshipContents, err := os.ReadFile(filepath.Join(execPath, "config", target+".json"))
	err = json.Unmarshal(worshipContents, &elements)

	fontSize := config.Classification.Bulletin.Print.FontSize
	fontOption := config.Classification.Bulletin.Print.FontOption
	for i, file := range files {
		imgPath := filepath.Join(outputDir, file.Name())

		objPdf.AddPage()
		objPdf.CheckImgPlaced(bulletinPrintSize, imgPath, 0)
		contentSectionDivision := "18"
		if i == 0 {
			sunDateText := date.SetThisSunDay()
			objPdf.SetText(fontOption, fontSize*1.2, true, colorPalette.HexToRGBA(objPdf.Config.Color.DateColor))
			objPdf.WriteText(sunDateText, "right")
		} else {
			ym := objPdf.ForComposeBuiltin(elements, contentSectionDivision)
			objPdf.ForReferNext(elements, contentSectionDivision, ym)
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

func getSize(config extract.Config) (gofpdf.SizeType, presentation.Size) {
	bulletinSize := gofpdf.SizeType{
		Wd: config.Classification.Bulletin.Print.Width,
		Ht: config.Classification.Bulletin.Print.Height,
	}
	rectangle := presentation.Size{
		Width:  config.Classification.Bulletin.Print.InnerRectangle.Width,
		Height: config.Classification.Bulletin.Print.InnerRectangle.Height,
	}

	return bulletinSize, rectangle
}
