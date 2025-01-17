package forPrint

import (
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/internal/sorted"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"os"
	"path/filepath"
)

func CreatePrint(figmaInfo *get.Info, execPath string, config extract.Config) {
	outputDir := filepath.Join(execPath, config.OutputPath.Bulletin, "print", "tmp")
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	figmaInfo.GetFigmaImage(outputDir, "forPrint")

	highestLuminaceColor := colorPalette.HexToRGBA(config.Color.BoxColor)
	bulletinSize, rectangle := getSize(config)

	files, _ := os.ReadDir(outputDir)
	objPdf := presentation.New(bulletinSize)
	objPdf.FullSize = bulletinSize
	objPdf.BoxSize = rectangle

	yearMonth, weekFormatted := date.SetDateTitle()

	// 파일명 생성: "202411_3.pdf"
	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	sorted.ToIntSort(files, "- ", ".png", 0)

	for i, file := range files {
		imgPath := filepath.Join(outputDir, file.Name())

		objPdf.AddPage()
		objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)

		if i == 0 {
			sunDatText := date.SetThisSunDay()
			objPdf.SetText(10, true, highestLuminaceColor)
			objPdf.WriteText(sunDatText, "right")
		}

	}
	outputBtPath := filepath.Join(execPath, config.OutputPath.Bulletin, "print")

	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, outputFilename)
	err := objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}
}

// A4 기준
func getSize(config extract.Config) (gofpdf.SizeType, presentation.Size) {
	bulletinSize := gofpdf.SizeType{
		Wd: config.Size.Background.Print.Width,
		Ht: config.Size.Background.Print.Height,
	}
	rectangle := presentation.Size{
		Width:  config.Size.InnerRectangle.Width,
		Height: config.Size.InnerRectangle.Height,
	}

	return bulletinSize, rectangle
}
