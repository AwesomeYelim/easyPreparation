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
	lowestLuminaceColor := colorList[0]

	// A4 기준
	bulletinSize := gofpdf.SizeType{
		Wd: 297.0,
		Ht: 210.0,
	}
	imgPath := "./public/images/coverdesign.png"
	outputDir := "./output/bulletin"
	fontPath := "./public/font/NotoSansKR-Bold.ttf" // 폰트 파일 경로 지정

	pkg.CheckDirIs(outputDir)

	objPdf := presentation.New(bulletinSize)
	objPdf.AddUTF8Font("NotoSansKR-Bold", "", fontPath) // UTF-8 폰트 등록
	objPdf.SetFont("NotoSansKR-Bold", "", 16)           // UTF-8 폰트를 기본 폰트로 설정

	objPdf.AddPage()
	objPdf.CheckImgPlaced(bulletinSize, imgPath, 0.5)
	rectangle := struct {
		width  float64
		height float64
	}{
		width:  132,
		height: 71,
	}
	hRgba := colorPalette.ConvertToRGBRange(highestLuminaceColor.Color.RGBA())
	lRgba := colorPalette.ConvertToRGBRange(lowestLuminaceColor.Color.RGBA())
	padding := (bulletinSize.Wd/2 - rectangle.width) / 2
	yPadding := bulletinSize.Ht - (padding + rectangle.height)

	objPdf.SetFillColor(int(hRgba[0]), int(hRgba[1]), int(hRgba[2]))
	objPdf.Rect(padding, yPadding, rectangle.width, rectangle.height, "F")

	// 텍스트 추가
	objPdf.SetTextColor(int(lRgba[0]), int(lRgba[1]), int(lRgba[2]))
	objPdf.Text(padding, padding, "교회 소식") // 한글 텍스트

	objPdf.DrawLine(rectangle.width+padding, padding, lowestLuminaceColor.Color)

	fmt.Println(padding, yPadding, rectangle.width, rectangle.height, int(hRgba[0]), int(hRgba[1]), int(hRgba[2]))
	err := objPdf.OutputFileAndClose(filepath.Join(outputDir, "sample.pdf"))
	if err != nil {
		log.Fatalf("PDF 저장 중 에러 발생: %v", err)
	}
}
