package forPresentation

import (
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"os"
	"path/filepath"
	"strings"
)

func CreatePresentation(figmaInfo *get.Info, target, execPath string) {
	config := extract.ConfigMem
	outputDir := filepath.Join(execPath, config.OutputPath.Bulletin, "presentation", "tmp")
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	figmaInfo.GetFigmaImage(outputDir, "forShowing")

	yearMonth, weekFormatted := date.SetDateTitle()
	bulletinSize, rectangle := getSize(config)

	objPdf := presentation.New(bulletinSize)
	objPdf.FullSize = bulletinSize
	objPdf.BoxSize = rectangle

	objPdf.ExecutePath = execPath

	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	var contents []gui.WorshipInfo

	worshipContents, err := os.ReadFile(filepath.Join(execPath, "config", target+".json"))
	err = json.Unmarshal(worshipContents, &contents)

	for _, con := range contents {
		splitTitle := strings.Split(con.Title, "_")
		objPdf.Title = splitTitle[1]

		if path, ok := figmaInfo.PathInfo[objPdf.Title]; ok {
			objPdf.Path = filepath.Join(outputDir, filepath.Base(path))
			if !strings.Contains(objPdf.Title, "성시교독") {
				objPdf.AddPage()
				objPdf.CheckImgPlaced(objPdf.FullSize, objPdf.Path, 0)
			}
			if strings.Contains(con.Info, "edit") {
				objPdf.ForEdit(con, config, execPath)
			}
		}
	}

	outputBtPath := filepath.Join(execPath, config.OutputPath.Bulletin, "presentation")
	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, outputFilename)
	err = objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}

}

func getSize(config extract.Config) (gofpdf.SizeType, presentation.Size) {
	bulletinSize := gofpdf.SizeType{
		Wd: config.Size.Background.Presentation.Width,
		Ht: config.Size.Background.Presentation.Height,
	}
	rectangle := presentation.Size{
		Width:  config.Size.Background.Presentation.InnerRectangle.Width,
		Height: config.Size.Background.Presentation.InnerRectangle.Height,
	}

	return bulletinSize, rectangle
}
