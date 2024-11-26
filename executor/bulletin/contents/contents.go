package contents

import (
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"os"
	"path/filepath"
)

func CreateContents() {
	execPath, _ := os.Getwd()
	execPath = path.ExecutePath(execPath, "easyPreparation")
	token, key, ui := gui.Connector()
	configPath := filepath.Join(execPath, "config/custom.json")
	config := extract.ExtCustomOption(configPath)

	defer func() {
		_ = ui.Close()
	}()
	figmaInfo := figma.New(&token, &key, execPath)

	outputDir := filepath.Join(execPath, config.OutputPath.Bulletin, "tmp")
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	figmaInfo.GetNodes()
	figmaInfo.GetContents()
	figmaInfo.GetFigmaImage(outputDir, "forPrint")

	highestLuminaceColor := colorPalette.HexToRGBA(config.Color.BoxColor)
	bulletinSize, rectangle := getSize(config)

	files, _ := os.ReadDir(outputDir)
	objPdf := presentation.New(bulletinSize)

	yearMonth, weekFormatted := date.SetDateTitle()

	// 파일명 생성: "202411_3.pdf"
	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	for i, file := range files {
		imgPath := filepath.Join(outputDir, file.Name())

		objPdf.AddPage()
		objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)
		padding := (bulletinSize.Wd/2 - rectangle.Width) / 2

		if i == 0 {
			sunDatText := date.SetThisSunDay()
			objPdf.WriteText("right", rectangle, sunDatText, padding, "end", 10, execPath, highestLuminaceColor)
		}

	}
	outputBtPath := filepath.Join(execPath, config.OutputPath.Bulletin)

	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, outputFilename)
	err := objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		msg := fmt.Sprintf(`document.getElementById("responseMessage").textContent = "PDF 저장 중 에러 발생: %v"`, err)
		ui.Eval(msg)
	}
}

// A4 기준
func getSize(config extract.Config) (gofpdf.SizeType, presentation.BoxSize) {
	bulletinSize := gofpdf.SizeType{
		Wd: config.Size.Background.Width,
		Ht: config.Size.Background.Height,
	}
	rectangle := presentation.BoxSize{
		Width:  config.Size.InnerRectangle.Width,
		Height: config.Size.InnerRectangle.Height,
	}

	return bulletinSize, rectangle
}
