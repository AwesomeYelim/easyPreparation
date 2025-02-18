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

	fontColor := colorPalette.HexToRGBA(config.Color.FontColor)
	printColor := colorPalette.HexToRGBA(config.Color.PrintColor)

	bulletinSize, rectangle := getSize(config)
	files, _ := os.ReadDir(outputDir)
	objPdf := presentation.New(bulletinSize)
	objPdf.BoxSize = rectangle
	yearMonth, weekFormatted := date.SetDateTitle()

	// 파일명 생성: "202411_3.pdf"
	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	sorted.ToIntSort(files, "- ", ".png", 0)

	var contents []gui.WorshipInfo

	worshipContents, err := os.ReadFile(filepath.Join(execPath, "config", target+".json"))
	err = json.Unmarshal(worshipContents, &contents)

	// 글씨 시작 margin
	var xm float64 = 95
	var ym float64 = 202
	var line float64 = 272
	var lineM float64 = 10

	fontSize := config.Size.Bulletin.Print.FontSize
	fontOption := config.Size.Bulletin.Print.FontOption

	for i, file := range files {
		imgPath := filepath.Join(outputDir, file.Name())

		objPdf.AddPage()
		objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)

		if i == 0 {
			sunDatText := date.SetThisSunDay()
			objPdf.SetText(fontOption, fontSize, true, fontColor)
			objPdf.WriteText(sunDatText, "right")
		} else {
			objPdf.SetText(fontOption, fontSize, false, printColor)
			for _, order := range contents {
				// 하위 목록인 경우 skip
				if strings.Contains(order.Title, ".") {
					continue
				}
				objPdf.SetXY(xm, ym)
				title := strings.Split(order.Title, "_")[1]
				objPdf.MultiCell(objPdf.BoxSize.Width, fontSize/2, title, "", "L", false)
				if strings.Contains(order.Title, "참회의 기도") || order.Obj == "-" {
					continue
				}
				linePlace := ym - (fontSize / 2)

				if strings.HasSuffix(order.Info, "edit") {
					strTW := objPdf.GetStringWidth(title)
					strOW := objPdf.GetStringWidth(order.Obj)
					editLine := (line - (strOW + lineM)) / 2
					firstPlacedLine := xm + strTW + lineM
					secondPlacedLine := firstPlacedLine + editLine + strOW + (lineM * 2)

					objPdf.DrawLine(editLine, firstPlacedLine, linePlace, printColor)
					objPdf.MultiCell(objPdf.BoxSize.Width, fontSize/2, order.Obj, "", "C", false)
					objPdf.DrawLine(editLine, secondPlacedLine, linePlace, printColor)
				} else {
					objPdf.DrawLine(line, xm, linePlace, printColor)
				}
				objPdf.MultiCell(objPdf.BoxSize.Width, fontSize/2, order.Lead, "", "R", false)
				ym += fontSize / 2

			}

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
		Wd: config.Size.Bulletin.Print.Width,
		Ht: config.Size.Bulletin.Print.Height,
	}
	rectangle := presentation.Size{
		Width:  config.Size.Bulletin.Print.InnerRectangle.Width,
		Height: config.Size.Bulletin.Print.InnerRectangle.Height,
	}

	return bulletinSize, rectangle
}
