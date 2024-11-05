package contents

import (
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"path/filepath"
)

func CreateContents() {
	//highestLuminaceColor := hexToRGBA("#F8F3EA") // 옅은색상
	//lowestLuminaceColor := hexToRGBA("#BEA07C")  // 진한색상

	// A4 기준
	bulletinSize := gofpdf.SizeType{
		Wd: 297.0,
		Ht: 210.0,
	}

	imgPath := "./public/images/fullcover.png"
	outputDir := "./output/bulletin"

	pkg.CheckDirIs(outputDir)

	objPdf := presentation.New(bulletinSize)

	objPdf.AddPage()
	objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)
	//rectangle := presentation.BoxSize{
	//	Width:  132,
	//	Height: 71,
	//}
	//text := "교회 소식"
	//
	//padding := (bulletinSize.Wd/2 - rectangle.Width) / 2
	//yPadding := bulletinSize.Ht - (padding + rectangle.Height)

	//objPdf.DrawBox(rectangle, padding, yPadding, highestLuminaceColor)
	//objPdf.WriteText(rectangle, text, padding, "center", lowestLuminaceColor)
	//objPdf.DrawLine(rectangle.Width, padding, padding+3, lowestLuminaceColor)

	err := objPdf.OutputFileAndClose(filepath.Join(outputDir, "sample.pdf"))
	if err != nil {
		log.Fatalf("PDF 저장 중 에러 발생: %v", err)
	}
}

func hexToRGBA(hex string) color.RGBA {
	var r, g, b uint8
	_, _ = fmt.Sscanf(hex, "#%02X%02X%02X", &r, &g, &b)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
