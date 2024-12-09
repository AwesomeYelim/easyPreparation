package presentation

import (
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/lyrics"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"os"
)

type PDF struct {
	*gofpdf.Fpdf
	BoxSize
}

type BoxSize struct {
	Width  float64
	Height float64
}

func New(size gofpdf.SizeType) PDF {
	// PDF 객체 생성
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P", // "P"는 세로, "L"은 가로
		UnitStr:        "mm",
		Size:           size,
	})

	return PDF{Fpdf: pdf}
}

func CreatePresentation(slidesData *lyrics.SlideData, filePath string) {
	pdfSize := gofpdf.SizeType{
		Wd: 297.0,
		Ht: 167.0,
	}
	objPdf := New(pdfSize)

	fontPath := "./public/font/NotoSansKR-Bold.ttf"
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		log.Fatalf("폰트 파일을 찾을 수 없습니다: %s", err)
	}
	objPdf.AddUTF8Font("NotoSansKR-Bold", "", fontPath)
	objPdf.SetFont("NotoSansKR-Bold", "", 40)
	// 글씨 색상 변경 (RGB 색상 지정)
	objPdf.SetTextColor(255, 255, 255)

	// PDF 페이지 추가
	for _, content := range slidesData.Content {
		objPdf.AddPage()
		// 배경 이미지 추가 (배경 이미지를 추가하려면 파일이 필요합니다)
		backgroundImage := "./public/images/ppt_background.png"
		objPdf.CheckImgPlaced(pdfSize, backgroundImage, 0)

		var textW float64 = 250
		var textH float64 = 20
		// 텍스트 추가 (내용 설정)
		objPdf.SetXY((pdfSize.Wd-textW)/2, (pdfSize.Ht-textH)/2)
		objPdf.MultiCell(textW, textH, content, "", "C", false)

	}
	// PDF 저장
	err := objPdf.OutputFileAndClose(filePath)
	if err != nil {
		log.Fatalf("PDF 저장 중 에러 발생: %v", err)
	}
}

// img 경로 확인 로직

func (pdf *PDF) CheckImgPlaced(pdfSize gofpdf.SizeType, path string, place float32) {
	if _, err := os.Stat(path); err == nil {
		switch place {
		case 0:
			pdf.ImageOptions(path, 0, 0, pdfSize.Wd, pdfSize.Ht, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		case 0.5:
			pdf.ImageOptions(path, pdfSize.Wd/2, 0, pdfSize.Wd/2, pdfSize.Ht, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		default:
			log.Print("해당되지 않은 값")
		}
	} else {
		log.Printf("배경 이미지 파일이 존재하지 않음: %s\n", err)
	}
}

// 박스 그려주는 함수
func (pdf *PDF) DrawBox(boxSize BoxSize, x, y float64, color ...color.Color) {
	colorList := colorPalette.GetColorWithSortByLuminance()
	highestLuminaceColor := colorList[len(colorList)-1] // 채도 가장 낮은 색상 - background

	var rgba []uint32
	if len(color) <= 0 {
		rgba = colorPalette.ConvertToRGBRange(highestLuminaceColor.Color.RGBA())
	} else {
		rgba = colorPalette.ConvertToRGBRange(color[0].RGBA())
	}

	pdf.SetFillColor(int(rgba[0]), int(rgba[1]), int(rgba[2]))
	pdf.Rect(x, y, boxSize.Width, boxSize.Height, "F")

}

// 선 그려주는 함수
func (pdf *PDF) DrawLine(length, x, y float64, color ...color.Color) {

	colorList := colorPalette.GetColorWithSortByLuminance()
	lowestLuminaceColor := colorList[0] // 채도 가장 낮은 색상 - background

	var rgba []uint32
	if len(color) <= 0 {
		rgba = colorPalette.ConvertToRGBRange(lowestLuminaceColor.Color.RGBA())
	} else {
		rgba = colorPalette.ConvertToRGBRange(color[0].RGBA())
	}

	pdf.SetDrawColor(int(rgba[0]), int(rgba[1]), int(rgba[2]))
	pdf.SetLineWidth(0.5)
	pdf.Line(x, y, x+length, y)
}

func (pdf *PDF) WriteText(x, y, fontSize float64, text string, textColor ...color.Color) {
	fontPath := "./public/font/NotoSansKR-Bold.ttf"
	// 폰트 설정
	pdf.AddUTF8Font("NotoSansKR-Bold", "", fontPath)
	pdf.SetFont("NotoSansKR-Bold", "", fontSize)
	textWidth := pdf.GetStringWidth(text)

	// 텍스트 색상 설정
	rgba := pdf.getColorRGB(textColor)

	// 텍스트 출력
	pdf.SetTextColor(int(rgba[0]), int(rgba[1]), int(rgba[2]))
	if x > 10 {
		x = x - (textWidth / 2)
	}
	pdf.Text(x, y, text)

}

// 텍스트 색상 설정 함수
func (pdf *PDF) getColorRGB(textColor []color.Color) []uint32 {
	var rgba []uint32
	if len(textColor) == 0 {
		// 기본 색상 사용
		colorList := colorPalette.GetColorWithSortByLuminance()
		lowestLuminanceColor := colorList[0]
		rgba = colorPalette.ConvertToRGBRange(lowestLuminanceColor.Color.RGBA())
	} else {
		rgba = colorPalette.ConvertToRGBRange(textColor[0].RGBA())
	}
	return rgba
}

// 텍스트 위치 계산 함수
func (pdf *PDF) calculateTextPosition(side, xAlignment, yAlignment string, padding, boxWidth, textWidth float64) (float64, float64) {
	var x float64
	var y float64

	leftOffset, centerOffset, rightOffset := 0.0, 99.0, 148.0

	switch side {
	case "right":
		x = rightOffset
		break
	case "left":
		x = leftOffset
		break
	case "center":
		x = centerOffset + (textWidth / 2)
		break
	}

	switch yAlignment {
	case "top":
		y = padding
		break
	case "center":
		y = 105.0
		break
	case "bottom":
		y = 210 - padding
		break
	}

	// 정렬 설정
	switch xAlignment {
	case "start":
		x += padding
		break
	case "center":
		if side != "center" {
			x += padding + (boxWidth-textWidth)/2
		}
		break
	case "end":
		x += padding + (boxWidth - textWidth)
		break
	}
	return x, y
}
