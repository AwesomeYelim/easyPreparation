package presentation

import (
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type PDF struct {
	*gofpdf.Fpdf
	Title    string
	FullSize gofpdf.SizeType
	BoxSize  Size
	Contents []string
	Path     string
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
func (pdf *PDF) ForEdit(con get.Children, config extract.Config, execPath string) {
	hLColor := colorPalette.HexToRGBA(config.Color.BoxColor) // 가장 채도가 낮음
	var textSize float64 = 25
	var textW float64 = 230

	pdf.SetText(textSize, hLColor)
	trimmedText := pkg.RemoveEmptyLines(con.Obj)
	pdf.Contents = strings.Split(trimmedText, "\n")

	switch pdf.Title {
	case "예배의 부름":
		pdf.setBegin(con, textW, textSize, 4)
	case "말씀내용":
		pdf.setBody(textW, textSize, 3)
	case "찬송", "헌금봉헌":
		pdf.SetText(27, hLColor)
		pdf.WriteText(148.5, 110, con.Content)
		pdf.setOutDirFiles(filepath.Join(execPath, "data", "hymn"), con.Content)
	case "성시교독":
		pdf.setOutDirFiles(filepath.Join(execPath, "data", "responsive_reading"), con.Content)
	default:
		pdf.SetText(27, hLColor)
		pdf.WriteText(148.5, 110, con.Content)
	}
}

func (pdf *PDF) setBegin(con get.Children, textW float64, textSize float64, lines int) {
	var tmpEl string
	for i, _ := range pdf.Contents {

		// 앞에 장 : 절 삭제
		//pdf.Contents[i] = parser.RemoveLineNumberPattern(pdf.Contents[i])
		if i == 0 {
			pdf.Contents[i] = fmt.Sprintf("%s\n%s", con.Content, pdf.Contents[i])
		}
		// 라인 기준으로 페이지를 생성
		if i != 0 && i%lines == 0 {
			if i/lines != 1 {
				pdf.AddPage()
				pdf.CheckImgPlaced(pdf.FullSize, pdf.Path, 0)
			}
			tmpEl += pdf.Contents[i] + "\n"
			pdf.SetXY((pdf.FullSize.Wd-textW)/2, textSize*3)
			pdf.MultiCell(textW, textSize/2, tmpEl, "", "L", false)
			tmpEl = ""
			// 잉여 라인이 생기는 경우 마지막 페이지를 추가
		} else if len(pdf.Contents)%lines < lines && i == len(pdf.Contents)-1 {
			tmpEl += pdf.Contents[i]
			//pdf.AddPage()
			//pdf.CheckImgPlaced(pdf.FullSize, pdf.Path, 0)
			pdf.SetXY(textSize, textSize*3)
			pdf.MultiCell(textW, textSize/2, tmpEl, "", "L", false)
		} else {
			tmpEl += pdf.Contents[i] + "\n"
		}
	}
}

func (pdf *PDF) setBody(textW float64, textSize float64, lines int) {
	var tmpEl string

	for i, _ := range pdf.Contents {

		if i != 0 && i%lines == 0 {
			if i/lines != 1 {
				pdf.AddPage()
				pdf.CheckImgPlaced(pdf.FullSize, pdf.Path, 0)
			}
			if i%2 == 0 {
				tmpEl += "▶ " + pdf.Contents[i] + "\n\n"
			} else {
				tmpEl += pdf.Contents[i] + "\n\n"
			}
			pdf.SetXY(textSize, textSize)
			pdf.MultiCell(textW, textSize/2, tmpEl, "", "L", false)
			tmpEl = ""
		} else if len(pdf.Contents)%lines < lines && i == len(pdf.Contents)-1 {
			if i%2 == 0 {
				tmpEl += "▶ " + pdf.Contents[i]
			} else {
				tmpEl += pdf.Contents[i]
			}
			pdf.AddPage()
			pdf.CheckImgPlaced(pdf.FullSize, pdf.Path, 0)
			pdf.SetXY(textSize, textSize)
			pdf.MultiCell(textW, textSize/2, tmpEl, "", "L", false)
		} else {
			if i%2 == 0 {
				tmpEl += "▶ " + pdf.Contents[i] + "\n\n"
			} else {
				tmpEl += pdf.Contents[i] + "\n\n"
			}
		}
	}
}

func (pdf *PDF) setOutDirFiles(pdfPath, target string) {
	// gs 를 사용해 pdf 파일을 png로 변환
	//cmd := fmt.Sprintf("soffice --headless --convert-to pdf %s", pdfPath)
	//gs -sDEVICE=pngalpha -o 56.png -r96 output/bulletin/presentation/202412_5.pdf

	var splitNum string
	if strings.Contains(pdfPath, "hymn") {
		splitNum = strings.TrimSuffix(target, "장")
	} else {
		splitNum = strings.Split(target, ".")[0]
	}
	// 캐싱 방지
	outputPath := filepath.Join(pdfPath, fmt.Sprintf("temp_%s", splitNum))
	_ = pkg.CheckDirIs(outputPath)

	tempPngPtah := filepath.Join(outputPath, "%d.png")

	targetNum := fmt.Sprintf("%03s.pdf", splitNum)

	var cmdStr string
	var cmd *exec.Cmd
	osType := runtime.GOOS

	switch osType {
	case "windows":
		cmdStr = fmt.Sprintf("gswin64c -sDEVICE=pngalpha -o \"%s\" -r96 \"%s\"", tempPngPtah, filepath.Join(pdfPath, targetNum))
		cmd = exec.Command("cmd", "/C", cmdStr)
	default:
		cmdStr = fmt.Sprintf("gs -sDEVICE=pngalpha -o \"%s\" -r96 \"%s\"", tempPngPtah, filepath.Join(pdfPath, targetNum))
		cmd = exec.Command("bash", "-c", cmdStr)
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatalf("명령어 실행 실패: %s, 에러: %v", string(output), err)
	}
	defer func() {
		_ = os.RemoveAll(outputPath)
	}()
	err = pdf.AddImagesToPDF(outputPath)
	if err != nil {
		log.Fatalf("이미지 PDF 추가 실패: %v", err)
	}
}

func (pdf *PDF) AddImagesToPDF(imageDir string) error {
	files, err := os.ReadDir(imageDir)
	if err != nil {
		return fmt.Errorf("이미지 디렉토리 읽기 실패: %v", err)
	}

	var imageFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".png" {
			imageFiles = append(imageFiles, filepath.Join(imageDir, file.Name()))
		}
	}

	sort.Strings(imageFiles)

	for _, imgPath := range imageFiles {
		pdf.AddPage()
		pdf.CheckImgPlaced(pdf.FullSize, imgPath, 0)
	}

	return nil
}
