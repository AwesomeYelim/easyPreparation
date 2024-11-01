package contents

import (
	"easyPreparation_1.0/executor/bulletin/cover/colorPalette"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"log"
	"path/filepath"
)

func CreateContents() {
	colorList := colorPalette.GetColorWithSortByLuminance()
	highestLuminaceColor := colorList[len(colorList)-1]

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
	rectangle := struct {
		width  float64
		height float64
	}{
		width:  132,
		height: 71,
	}
	rgba := colorPalette.ConvertToRGBRange(highestLuminaceColor.Color.RGBA())
	xPadding := (bulletinSize.Wd/2 - rectangle.width) / 2
	yPadding := bulletinSize.Ht - (xPadding + rectangle.height)

	objPdf.SetFillColor(int(rgba[0]), int(rgba[1]), int(rgba[2]))
	objPdf.Rect(xPadding, yPadding, rectangle.width, rectangle.height, "F")
	fmt.Println(xPadding, yPadding, rectangle.width, rectangle.height, int(rgba[0]), int(rgba[1]), int(rgba[2]))
	err := objPdf.OutputFileAndClose(filepath.Join(outputDir, "sample.pdf"))
	if err != nil {
		log.Fatalf("PDF 저장 중 에러 발생: %v", err)
	}
}
