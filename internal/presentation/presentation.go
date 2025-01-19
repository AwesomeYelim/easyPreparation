package presentation

import (
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/googleCloud"
	"easyPreparation_1.0/internal/gui"
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
	Title       string
	FullSize    gofpdf.SizeType
	BoxSize     Size
	Contents    []string
	Path        string
	ExecutePath string
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

func (pdf *PDF) WriteText(text, position string, custom ...float64) {
	textWidth := pdf.GetStringWidth(text)

	var x, y float64
	switch position {
	case "center":
		x = pdf.FullSize.Wd / 2
		y = pdf.FullSize.Ht/2 + 5
		if x > 10 {
			x = x - (textWidth / 2)
		}
	case "right":
		padding := (pdf.FullSize.Wd/2 - pdf.BoxSize.Width) / 2
		x = pdf.FullSize.Wd - (textWidth + padding)
		y = padding
	case "custom":
		x = custom[0]
		y = custom[1]
	}

	pdf.Text(x, y, text)
}

func (pdf *PDF) SetText(fontSize float64, isB bool, textColor ...color.Color) {
	fontBPath := "./public/font/NanumGothic-ExtraBold.ttf"
	fontPath := "./public/font/NanumGothic-Regular.ttf"

	if isB {
		pdf.AddUTF8Font("NanumGothic-ExtraBold", "B", fontBPath)
		pdf.SetFont("NanumGothic-ExtraBold", "B", fontSize)
	} else {
		pdf.AddUTF8Font("NanumGothic-Regular", "", fontPath)
		pdf.SetFont("NanumGothic-Regular", "", fontSize)
	}

	// 텍스트 색상 설정
	rgba := pdf.getColorRGB(textColor)
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
func (pdf *PDF) ForEdit(con gui.WorshipInfo, config extract.Config, execPath string) {
	hLColor := colorPalette.HexToRGBA(config.Color.BoxColor) // 박스 색상 설정
	var textSize float64 = 25
	var textW float64 = 230

	pdf.SetText(textSize, true, hLColor)
	trimmedText := pkg.RemoveEmptyLines(con.Contents)
	pdf.Contents = trimmedText

	switch pdf.Title {
	case "예배의 부름":
		pdf.setBegin(con, textW, textSize, 4)
	case "말씀내용":
		pdf.setBody(textW, textSize, 3)
	case "찬송", "헌금봉헌":
		pdf.SetText(27, true, hLColor)
		pdf.WriteText(con.Obj, "center")
		pdf.setOutDirFiles("hymn", con.Obj)
	case "성시교독":
		pdf.setOutDirFiles("responsive_reading", con.Obj)
	case "교회소식":
		pdf.DrawChurchNews(con, hLColor)
	default:
		pdf.WriteText(con.Obj, "center")
	}
}

func (pdf *PDF) DrawChurchNews(con gui.WorshipInfo, hLColor color.RGBA) {
	// 재귀적으로 교회소식과 그 내부 children 데이터를 처리하는 함수
	var draw func(items []gui.WorshipInfo, depth int)

	x, y := 10.0, 50.0
	fontSize := 27.0
	var tmpData string
	pdf.SetText(fontSize, false, hLColor)

	draw = func(items []gui.WorshipInfo, depth int) {
		for i, item := range items {
			tab := strings.Repeat("\t", depth-1)
			// 데이터 추가
			tmpData += tab + fmt.Sprintf("%d. %s: %s", i+1, item.Title, item.Obj) + "\n"

			// children이 있는 경우 재귀 호출
			if len(item.Children) > 0 {
				draw(item.Children, depth+1) // depth 증가
			}
		}
	}

	if con.Title == "13_교회소식" {
		// children 처리
		if len(con.Children) > 0 {
			draw(con.Children, 1)
		}
	}

	// 최종 출력
	pdf.SetXY(x, y)
	pdf.MultiCell(pdf.BoxSize.Width, fontSize/2, tmpData, "", "L", false)
}

func (pdf *PDF) setBegin(con gui.WorshipInfo, textW float64, textSize float64, lines int) {
	var tmpEl string
	for i, _ := range pdf.Contents {

		if i == 0 {
			pdf.Contents[i] = fmt.Sprintf("%s\n%s", con.Obj, pdf.Contents[i])
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

	for i, content := range pdf.Contents {
		if i != 0 && i%lines == 0 {
			if i/lines != 1 {
				pdf.AddPage()
				pdf.CheckImgPlaced(pdf.FullSize, pdf.Path, 0)
			}
			tmpEl += content + "\n\n"
			pdf.SetXY(textSize, textSize)
			pdf.MultiCell(textW, textSize/2, tmpEl, "", "L", false)
			tmpEl = ""
		} else if len(pdf.Contents)%lines < lines && i == len(pdf.Contents)-1 {
			tmpEl += content
			pdf.AddPage()
			pdf.CheckImgPlaced(pdf.FullSize, pdf.Path, 0)
			pdf.SetXY(textSize, textSize)
			pdf.MultiCell(textW, textSize/2, tmpEl, "", "L", false)
		} else {
			tmpEl += content + "\n\n"
		}
	}
}

func (pdf *PDF) setOutDirFiles(category, target string) {
	// gs 를 사용해 pdf 파일을 png로 변환
	//cmd := fmt.Sprintf("soffice --headless --convert-to pdf %s", pdfPath)
	//gs -sDEVICE=pngalpha -o 56.png -r96 output/bulletin/presentation/202412_5.pdf

	var splitNum string
	switch category {
	case "hymn":
		splitNum = strings.TrimSuffix(target, "장")
	case "responsive_reading":
		splitNum = strings.Split(target, ".")[0]
	}
	outputPath := filepath.Join(pdf.ExecutePath, "data", category)

	_ = pkg.CheckDirIs(outputPath)
	defer func() {
		_ = os.RemoveAll(outputPath)
	}()

	targetNum := fmt.Sprintf("%03s.pdf", splitNum)

	googleCloud.GetGoogleCloudInfo(category, targetNum, outputPath)
	// 캐싱 방지
	tempPath := filepath.Join(outputPath, fmt.Sprintf("temp_%s", splitNum))
	_ = pkg.CheckDirIs(tempPath)
	tempPngPtah := filepath.Join(tempPath, "%d.png")

	var cmdStr string
	var cmd *exec.Cmd
	osType := runtime.GOOS

	switch osType {
	case "windows":
		cmd = exec.Command("gswin64c", "-sDEVICE=pngalpha", "-o", tempPngPtah, "-r96", filepath.Join(outputPath, targetNum))
	default:
		cmdStr = fmt.Sprintf("gs -sDEVICE=pngalpha -o \"%s\" -r96 \"%s\"", tempPngPtah, filepath.Join(outputPath, targetNum))
		cmd = exec.Command("bash", "-c", cmdStr)
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatalf("명령어 실행 실패: %s, 에러: %v", string(output), err)
	}
	defer func() {
		_ = os.RemoveAll(tempPath)
	}()
	err = pdf.AddImagesToPDF(tempPath)
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
