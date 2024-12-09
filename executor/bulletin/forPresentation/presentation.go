package forPresentation

import (
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	//files, _ := os.ReadDir(outputDir)
	objPdf := presentation.New(bulletinSize)

	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	//sorted.ToIntSort(files, "- ", ".png", 0)

	var contents []map[string]string
	custom, err := os.ReadFile(filepath.Join(execPath, "config", "main_worship.json"))
	err = json.Unmarshal(custom, &contents)

	for _, con := range contents {
		title := strings.Split(con["title"], "_")[1]
		log.Print("title ", title, " ", figmaInfo.PathInfo)

		if path, ok := figmaInfo.PathInfo[title]; ok {
			imgPath := filepath.Join(outputDir, filepath.Base(path))
			objPdf.AddPage()
			objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)
		}
	}
	//for _, file := range files {
	//	imgPath := filepath.Join(outputDir, file.Name())
	//
	//	objPdf.AddPage()
	//	objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)
	//
	//}
	outputBtPath := filepath.Join(execPath, config.OutputPath.Bulletin, "presentation")
	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, outputFilename)
	err = objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}

}
