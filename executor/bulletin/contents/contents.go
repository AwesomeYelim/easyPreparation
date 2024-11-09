package contents

import (
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

type Config struct {
	Color struct {
		BoxColor  string `json:"boxColor"`
		LineColor string `json:"lineColor"`
		FontColor string `json:"fontColor"`
		DateColor string `json:"dateColor"`
	} `json:"color"`
}

func CreateContents() {
	configPath := "./config/custom.json"
	var config Config
	custom, err := os.ReadFile(configPath)
	_ = json.Unmarshal(custom, &config)

	highestLuminaceColor := hexToRGBA(config.Color.BoxColor) // 옅은색상
	lowestLuminaceColor := hexToRGBA(config.Color.FontColor) // 진한색상

	// A4 기준
	bulletinSize := gofpdf.SizeType{
		Wd: 297.0,
		Ht: 210.0,
	}

	imgPath := "./public/images/coverdesign.png"
	outputDir := "./output/bulletin"

	pkg.CheckDirIs(outputDir)

	objPdf := presentation.New(bulletinSize)

	objPdf.AddPage()
	objPdf.CheckImgPlaced(bulletinSize, imgPath, 0.5)
	rectangle := presentation.BoxSize{
		Width:  132,
		Height: 71,
	}
	text := "교회 소식"

	padding := (bulletinSize.Wd/2 - rectangle.Width) / 2
	yPadding := bulletinSize.Ht - (padding + rectangle.Height)

	objPdf.DrawBox(rectangle, padding, yPadding, highestLuminaceColor)
	objPdf.WriteText("left", rectangle, text, padding, "center", 16, lowestLuminaceColor)
	objPdf.DrawLine(rectangle.Width, padding, padding+3, lowestLuminaceColor)

	// 이번주 주일 날짜 계산
	currentDate := time.Now()
	daysUntilSunday := (7 - int(currentDate.Weekday())) % 7
	thisSunday := currentDate.AddDate(0, 0, daysUntilSunday)

	// PDF에 날짜 추가
	dateText := thisSunday.Format("2006년 01월 02일")
	objPdf.WriteText("right", rectangle, dateText, padding, "end", 10, highestLuminaceColor)

	err = objPdf.OutputFileAndClose(filepath.Join(outputDir, "sample.pdf"))
	if err != nil {
		log.Fatalf("PDF 저장 중 에러 발생: %v", err)
	}
}

func hexToRGBA(hex string) color.RGBA {
	var r, g, b uint8
	_, _ = fmt.Sscanf(hex, "#%02X%02X%02X", &r, &g, &b)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
