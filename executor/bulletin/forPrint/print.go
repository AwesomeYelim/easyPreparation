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
	var targetWidth float64 = 100

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
				if strings.Contains(order.Title, "참회의 기도") || order.Obj == "-" {
					continue
				}
				objPdf.SetXY(xm, ym)
				title := strings.Split(order.Title, "_")
				objPdf.MultiCell(objPdf.BoxSize.Width, 0, title[1], "", "L", false)
				strTW := objPdf.GetStringWidth(title[1])
				strOW := objPdf.GetStringWidth(order.Obj)
				//strLW := objPdf.GetStringWidth(order.Lead)
				//line := objPdf.BoxSize.Width - (strTW + strLW + (lineM * 2))
				editLine := (line - (strOW + lineM)) / 2
				firstPlacedLine := xm + targetWidth + lineM
				secondPlacedLine := firstPlacedLine + editLine + strOW + (lineM * 2)

				if strings.HasSuffix(order.Info, "edit") && title[1] != "교회소식" {
					objPdf.SetXY(xm, ym)
					objPdf.DrawLine(editLine, firstPlacedLine, ym, printColor)

					if len(title[1]) > 1 {
						charSpacing := (targetWidth - strTW) / float64(len(title[1])-1)
						fmt.Println(charSpacing)
						objPdf.SetWordSpacing(charSpacing)
					}
					objPdf.MultiCell(objPdf.BoxSize.Width, 0, order.Obj, "", "C", false)
					objPdf.DrawLine(editLine, secondPlacedLine, ym, printColor)
				} else {
					objPdf.DrawLine(line, firstPlacedLine, ym, printColor)
				}
				objPdf.SetXY(xm, ym)
				objPdf.MultiCell(objPdf.BoxSize.Width, 0, order.Lead, "", "R", false)
				ym += fontSize / 1.5
				if title[0] == "17" {
					break
				}
				objPdf.SetWordSpacing(0)

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
