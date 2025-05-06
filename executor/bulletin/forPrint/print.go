package forPrint

import (
	"easyPreparation_1.0/executor/bulletin/define"
	"easyPreparation_1.0/internal/classification"
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
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

type PdfInfo struct {
	*define.PdfInfo
}

func (pi PdfInfo) Create() {
	config := extract.ConfigMem
	outputDir := filepath.Join(pi.ExecPath, config.OutputPath.Bulletin, "print", "tmp")
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	pi.FigmaInfo.GetFigmaImage(outputDir, "forPrint")

	files, _ := os.ReadDir(outputDir)
	instanceSize := gofpdf.SizeType{
		Wd: config.Classification.Bulletin.Print.Width,
		Ht: config.Classification.Bulletin.Print.Height,
	}
	objPdf := presentation.New(instanceSize)
	objPdf.Config = config.Classification.Bulletin.Print
	objPdf.GetConversionRatio()

	sorted.ToIntSort(files, "- ", ".png", 0)

	var elements []gui.WorshipInfo
	var newsCon gui.WorshipInfo

	worshipContents, err := os.ReadFile(filepath.Join(pi.ExecPath, "config", pi.Target+".json"))
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
	outputBtPath := filepath.Join(pi.ExecPath, config.OutputPath.Bulletin, "print")

	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, pi.OutputFilename)
	err = objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}
}
