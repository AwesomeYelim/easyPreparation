package forPresentation

import (
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"os"
	"path/filepath"
)

func CreatePresentation(figmaInfo *get.Info, execPath string, config extract.Config) {
	outputDir := filepath.Join(execPath, config.OutputPath.Bulletin, "presentation", "tmp")
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	figmaInfo.GetFigmaImage(outputDir, "forShowing")

	bulletinSize := gofpdf.SizeType{
		Wd: config.Size.Background.Width,
		Ht: config.Size.Background.Height,
	}
	yearMonth, weekFormatted := date.SetDateTitle()

	files, _ := os.ReadDir(outputDir)
	objPdf := presentation.New(bulletinSize)

	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	for _, file := range files {
		imgPath := filepath.Join(outputDir, file.Name())

		objPdf.AddPage()
		objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)

	}
	outputBtPath := filepath.Join(execPath, config.OutputPath.Bulletin, "presentation")
	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, outputFilename)
	err := objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}

}
