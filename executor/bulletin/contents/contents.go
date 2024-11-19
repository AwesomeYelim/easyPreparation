package contents

import (
	"easyPreparation_1.0/figma"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Color struct {
	BoxColor  string `json:"boxColor"`
	LineColor string `json:"lineColor"`
	FontColor string `json:"fontColor"`
	DateColor string `json:"dateColor"`
}

type Size struct {
	Background     gofpdf.SizeType      `json:"background"`
	InnerRectangle presentation.BoxSize `json:"innerRectangle"`
}

type Config struct {
	Color Color `json:"color"`
	Size  Size  `json:"size"`
}

func CreateContents() {
	token, key, ui := gui.Connector()

	defer func() {
		_ = ui.Close()
	}()

	ui.Eval(`document.getElementById("responseMessage").textContent = "Setting up data ~"`)

	figmaInfo := figma.New(&token, &key)

	outputDir := "./output/bulletin/tmp"
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	figmaInfo.GetNodes()
	figmaInfo.GetContents()
	figmaInfo.GetFigmaImage(outputDir)

	execPath, _ := os.Getwd()
	log.Println(execPath)

	configPath := "./config/custom.json"
	var config Config
	custom, err := os.ReadFile(configPath)

	if err != nil {
		err = json.Unmarshal(custom, &config)
	} else {
		config.Color.BoxColor = "#FFFFFF"
	}

	highestLuminaceColor := hexToRGBA(config.Color.BoxColor) // 옅은색상

	// A4 기준
	bulletinSize := gofpdf.SizeType{
		Wd: config.Size.Background.Wd,
		Ht: config.Size.Background.Ht,
	}
	rectangle := presentation.BoxSize{
		Width:  config.Size.InnerRectangle.Width,
		Height: config.Size.InnerRectangle.Height,
	}
	files, _ := os.ReadDir(outputDir)
	objPdf := presentation.New(bulletinSize)

	// 현재 날짜와 주차 정보를 계산
	currentDate := time.Now()
	firstDayOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
	_, firstWeek := firstDayOfMonth.ISOWeek()
	_, currentWeek := currentDate.ISOWeek()
	weekInMonth := currentWeek - firstWeek + 1
	yearMonth := currentDate.Format("200601")
	weekFormatted := fmt.Sprintf("%d", weekInMonth)

	// 파일명 생성: "202411_3.pdf"
	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	for i, file := range files {
		imgPath := fmt.Sprintf(outputDir+"/%s", file.Name())

		objPdf.AddPage()
		objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)
		padding := (bulletinSize.Wd/2 - rectangle.Width) / 2

		if i == 0 {
			// 이번주 주일 날짜 계산
			daysUntilSunday := (7 - int(currentDate.Weekday())) % 7
			thisSunday := currentDate.AddDate(0, 0, daysUntilSunday)

			// PDF에 날짜 추가
			dateText := thisSunday.Format("2006년 01월 02일")
			objPdf.WriteText("right", rectangle, dateText, padding, "end", 10, highestLuminaceColor)
		}

	}
	outputBtPath := "./output/bulletin"

	_ = pkg.CheckDirIs(outputBtPath)

	err = objPdf.OutputFileAndClose(filepath.Join(outputBtPath, outputFilename))
	if err != nil {
		msg := fmt.Sprintf(`document.getElementById("responseMessage").textContent = "PDF 저장 중 에러 발생: %v"`, err)
		ui.Eval(msg)
	}
}

func hexToRGBA(hex string) color.RGBA {
	var r, g, b uint8
	_, _ = fmt.Sscanf(hex, "#%02X%02X%02X", &r, &g, &b)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
