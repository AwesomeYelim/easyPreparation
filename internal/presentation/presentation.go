package presentation

import (
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/parser"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"os"
	"strings"
)

type PDF struct {
	*gofpdf.Fpdf
	Title      string
	FullSize   gofpdf.SizeType
	BoxSize    Size
	Contents   []string
	Path       string
	CommonPath string
}

type Size struct {
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
func (pdf *PDF) DrawBox(boxSize Size, x, y float64, color ...color.Color) {
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

func (pdf *PDF) WriteText(x, y float64, text string) {
	textWidth := pdf.GetStringWidth(text)

	if x > 10 {
		x = x - (textWidth / 2)
	}
	pdf.Text(x, y, text)
}

func (pdf *PDF) SetText(fontSize float64, textColor ...color.Color) {
	fontPath := "./public/font/NotoSansKR-Bold.ttf"
	// 폰트 설정
	pdf.AddUTF8Font("NotoSansKR-Bold", "", fontPath)
	pdf.SetFont("NotoSansKR-Bold", "", fontSize)

	// 텍스트 색상 설정
	rgba := pdf.getColorRGB(textColor)

	// 텍스트 출력
	pdf.SetTextColor(int(rgba[0]), int(rgba[1]), int(rgba[2]))
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

func (pdf *PDF) ForEdit(con get.Children, config extract.Config) {
	highestLuminaceColor := colorPalette.HexToRGBA(config.Color.BoxColor)
	switch pdf.Title {
	case "예배의 부름":
		var textSize float64 = 25
		var tmpEl string

		pdf.SetText(textSize, highestLuminaceColor)
		// 공백 제거
		trimmedText := pkg.RemoveEmptyLines(con.Obj)
		lines := strings.Split(trimmedText, "\n")

		for i, el := range lines {
			if strings.HasPrefix(el, "Bible Quote") {
				lines[i] = strings.TrimPrefix(el, "Bible Quote")
			}
			lines[i] = parser.RemoveLineNumberPattern(lines[i])
			if i == 0 {
				// 앞에 말씀 범위 표시하기
				lines[i] = fmt.Sprintf("%s\n%s", con.Content, lines[i])
			}
			// 4개씩 묶기
			if i != 0 && i%4 == 0 {
				// FIXME: 위에 한번 추가 해줘서
				if i/4 != 1 {
					pdf.AddPage()
					pdf.CheckImgPlaced(pdf.FullSize, pdf.Path, 0)
				}

				tmpEl += lines[i] + "\n"

				var textW float64 = 230
				// 텍스트 추가 (내용 설정)
				pdf.SetXY((pdf.FullSize.Wd-textW)/2, textSize*3)
				pdf.MultiCell(textW, textSize/2, tmpEl, "", "C", false)
				tmpEl = ""
			} else {
				tmpEl += lines[i] + "\n"
			}
		}
	case "성경봉독":

	default:
		pdf.SetText(27, highestLuminaceColor)
		pdf.WriteText(148.5, 110, con.Content)
	}

}
