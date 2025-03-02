package presentation

import (
	"easyPreparation_1.0/internal/classification"
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/font"
	"easyPreparation_1.0/internal/format"
	"easyPreparation_1.0/internal/googleCloud"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/parser"
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
	"strconv"
	"strings"
	"unicode/utf8"
)

type PDF struct {
	*gofpdf.Fpdf
	Title       string
	BibleVerse  []string
	Path        string
	ExecutePath string
	Config      classification.ResultInfo
}

type InnerSizeInfo struct {
	Width   float64
	Height  float64
	Padding float64
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

func (pdf *PDF) CheckImgPlaced(path string, place float32) {
	if _, err := os.Stat(path); err == nil {
		switch place {
		case 0:
			pdf.ImageOptions(path, 0, 0, pdf.Config.Width, pdf.Config.Height, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		case 0.5:
			pdf.ImageOptions(path, pdf.Config.Width/2, 0, pdf.Config.Width/2, pdf.Config.Height, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		default:
			log.Print("해당되지 않은 값")
		}
	} else {
		log.Printf("배경 이미지 파일이 존재하지 않음: %s\n", err)
	}
}

// 박스 그려주는 함수
func (pdf *PDF) DrawBox(x, y float64, color ...color.Color) {
	colorList := colorPalette.GetColorWithSortByLuminance()
	highestLuminaceColor := colorList[len(colorList)-1] // 채도 가장 낮은 색상 - background

	var rgba []uint32
	if len(color) <= 0 {
		rgba = colorPalette.ConvertToRGBRange(highestLuminaceColor.Color.RGBA())
	} else {
		rgba = colorPalette.ConvertToRGBRange(color[0].RGBA())
	}

	pdf.SetFillColor(int(rgba[0]), int(rgba[1]), int(rgba[2]))
	pdf.Rect(x, y, pdf.Config.InnerRectangle.Width, pdf.Config.InnerRectangle.Height, "F")

}

// 선 그려주는 함수
func (pdf *PDF) DrawLine(length, x, y float64, color color.Color) {

	rgba := colorPalette.ConvertToRGBRange(color.RGBA())

	pdf.SetDrawColor(int(rgba[0]), int(rgba[1]), int(rgba[2]))
	pdf.SetLineWidth(0.5)
	pdf.Line(x, y, x+length, y)
}

func (pdf *PDF) WriteText(text, position string, custom ...float64) {
	textWidth := pdf.GetStringWidth(text)

	var x, y float64
	switch position {
	case "center":
		x = pdf.Config.Size.Width / 2
		y = pdf.Config.Size.Height / 2
		if x > 10 {
			x = x - (textWidth / 2)
		}
	case "right":
		padding := (pdf.Config.Size.Width/2 - pdf.Config.InnerRectangle.Width) / 2
		x = pdf.Config.Size.Width - (textWidth + padding)
		y = padding
	case "custom":
		x = custom[0]
		y = custom[1]
	}
	pdf.Text(x, y, text)
}

func (pdf *PDF) SetText(fontInfo classification.FontInfo, isB bool, textColor ...color.Color) {
	var fontPath string
	var err error

	if isB {
		fontPath, err = font.GetFont(fontInfo.FontFamily, "800", isB)
	} else {
		fontPath, err = font.GetFont(fontInfo.FontFamily, "regular", isB)
	}

	if err != nil {
		fmt.Println("폰트 다운로드 에러:", err)
		return
	}

	pdf.AddUTF8Font(filepath.Base(fontPath), "B", fontPath)
	if err != nil {
		fmt.Println("폰트 추가 실패:", err)
		return
	}
	pdf.SetFont(filepath.Base(fontPath), "B", fontInfo.FontSize)

	// 텍스트 색상 설정
	if len(textColor) > 0 {
		rgba := textColor[0].(color.RGBA)
		pdf.SetTextColor(int(rgba.R), int(rgba.G), int(rgba.B))
	}
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

// 텍스트 간격
func (pdf *PDF) TextSpacingFormat(text string, targetWidth, x, y float64) {
	textWidth := pdf.GetStringWidth(text)
	charCount := utf8.RuneCountInString(text) // 한 문자가 아닌 한 글자씩
	if charCount > 1 {
		charSpacing := (targetWidth - textWidth) / float64(charCount-1)
		currentX := x

		for _, char := range text {
			charStr := string(char)
			pdf.SetXY(currentX, y)
			pdf.CellFormat(pdf.GetStringWidth(charStr), 0, charStr, "", 0, "L", false, 0, "")
			currentX += pdf.GetStringWidth(charStr) + charSpacing
		}
	} else {
		pdf.SetXY(x, y)
		pdf.CellFormat(targetWidth, 0, text, "", 0, "L", false, 0, "")
	}
}

func (pdf *PDF) ForComposeBuiltin(elements []gui.WorshipInfo, limit string) (ym float64) {
	// figma 디자인 기준
	var xm float64 = 95
	ym = 202
	var line float64 = 272
	var lineM float64 = 19
	var targetWidth float64 = 100
	var fontInfo = pdf.Config.FontInfo

	printColor := colorPalette.HexToRGBA(pdf.Config.Color.PrintColor)

	pdf.SetText(fontInfo, false, printColor)

	for _, order := range elements {
		// 하위 목록인 경우 skip
		if strings.Contains(order.Title, ".") {
			continue
		}

		pdf.SetXY(xm, ym)
		title := strings.Split(order.Title, "_")
		strOW := pdf.GetStringWidth(order.Obj)
		editLine := (line - (strOW + (lineM * 2))) / 2
		firstPlacedLine := xm + targetWidth + lineM
		secondPlacedLine := firstPlacedLine + editLine + strOW + (lineM * 2)

		if strings.Contains(title[1], "기도") {
			title[1] = "기도"
		}
		pdf.TextSpacingFormat(title[1], targetWidth, xm, ym)

		if strings.HasSuffix(order.Info, "c_edit") || strings.HasSuffix(order.Info, "b_edit") {
			pdf.SetXY(xm, ym)
			pdf.DrawLine(editLine, firstPlacedLine, ym, printColor)
			pdf.MultiCell(pdf.Config.InnerRectangle.Width, 0, order.Obj, "", "C", false)
			pdf.DrawLine(editLine, secondPlacedLine, ym, printColor)
		} else {
			pdf.DrawLine(line, firstPlacedLine, ym, printColor)
		}
		pdf.TextSpacingFormat(order.Lead, targetWidth, secondPlacedLine+editLine+lineM, ym)
		ym += fontInfo.FontSize / 1.8
		if title[0] == limit {
			break
		}
	}
	return ym
}

func (pdf *PDF) ForReferNext(elements []gui.WorshipInfo, strLimit string, nextStart float64) {

	pdf.Config.FontSize *= 0.8
	fontInfo := pdf.Config.FontInfo
	printColor := colorPalette.HexToRGBA(pdf.Config.Color.PrintColor)
	var xm float64 = 427
	var innerBoxWidth float64 = 180

	for idx, element := range elements {
		titles := strings.Split(element.Title, "_")
		conTitle, _ := strconv.Atoi(titles[0])
		limit, _ := strconv.Atoi(strLimit)
		if conTitle <= limit {
			continue
		}
		if idx == len(elements)-1 {
			break
		}
		pdf.SetText(fontInfo, true, printColor)
		pdf.SetXY(xm, nextStart)
		pdf.MultiCell(innerBoxWidth, 0, fmt.Sprintf("%s:", titles[1]), "", "L", false)
		pdf.SetXY(xm, nextStart)
		pdf.MultiCell(innerBoxWidth, 0, element.Obj, "", "R", false)
		nextStart += fontInfo.FontSize / 2
	}
}

func (pdf *PDF) ForTodayVerse(element gui.WorshipInfo) {
	pdf.Config.FontInfo.FontSize = pdf.Config.FontInfo.FontSize * 0.9
	fontInfo := pdf.Config.FontInfo

	printColor := colorPalette.HexToRGBA(pdf.Config.Color.PrintColor)
	var xm float64 = 190
	var ym float64 = 820
	var innerBoxW float64 = 330

	element.Contents = parser.RemoveLineNumberPattern(element.Contents)
	element.Contents += "\n\n" + element.Obj

	pdf.SetText(fontInfo, false, printColor)
	pdf.SetXY(xm, ym)
	pdf.MultiCell(innerBoxW, fontInfo.FontSize/2, element.Contents, "", "C", false)
}

func (pdf *PDF) ForEdit(con gui.WorshipInfo, config extract.Config, execPath string) {
	hLColor := colorPalette.HexToRGBA(pdf.Config.Color.BoxColor) // 박스 색상 설정
	fontInfo := config.Classification.Bulletin.Presentation.FontInfo

	pdf.SetText(fontInfo, true, hLColor)
	//trimmedText := pkg.RemoveEmptyLines(con.BibleVerse)
	pdf.BibleVerse = strings.Split(con.Contents, "\n")

	switch pdf.Title {
	case "예배의 부름":
		pdf.setBegin(con, 4)
	case "말씀내용":
		pdf.setBody(3) // 3 줄씩 끊기
	case "찬송", "헌금봉헌":
		pdf.SetText(fontInfo, true, hLColor)
		pdf.WriteText(con.Obj, "center")
		pdf.setOutDirFiles("hymn", con.Obj)
	case "성시교독":
		pdf.setOutDirFiles("responsive_reading", con.Obj)
	case "교회소식":
		pdf.DrawChurchNews(fontInfo, con, hLColor, pdf.Config.Padding, pdf.Config.Padding*2)
	case "참회의 기도":
		pdf.SetText(fontInfo, true, hLColor)
		pdf.SetXY(pdf.Config.Padding, pdf.Config.Padding*2)
		pdf.MultiCell(pdf.Config.InnerRectangle.Width, pdf.Config.FontSize/1.7, con.Obj, "", "R", false)
	default:
		if con.Obj == "-" {
			pdf.WriteText(con.Lead, "center")
		} else {
			pdf.WriteText(con.Obj, "center")
		}
	}
}

func (pdf *PDF) DrawChurchNews(fontInfo classification.FontInfo, con gui.WorshipInfo, hLColor color.RGBA, x, y float64) {
	// 재귀적으로 교회소식과 그 내부 children 데이터를 처리하는 함수
	var draw func(items []gui.WorshipInfo, depth int)

	var tmpData string
	pdf.SetText(fontInfo, false, hLColor)

	draw = func(items []gui.WorshipInfo, depth int) {
		for i, item := range items {
			tab := strings.Repeat("\t", depth-1)

			if item.Obj == "-" {
				item.Obj = ""
			}

			// 인덱스 포맷 적용
			index := format.IndexFormat(i, depth)
			if len(item.Children) == 0 {
				item.Title = fmt.Sprintf("%s:", item.Title)
			}
			// 데이터 추가
			tmpData += tab + fmt.Sprintf("%s %s %s", index, item.Title, item.Obj) + "\n"

			// children이 있는 경우 재귀 호출
			if len(item.Children) > 0 {
				draw(item.Children, depth+1) // depth 증가
			}
		}
	}

	if strings.Contains(con.Title, "교회소식") {
		// children 처리
		if len(con.Children) > 0 {
			draw(con.Children, 1)
		}
	}

	// 최종 출력
	pdf.SetXY(x, y)
	pdf.MultiCell(pdf.Config.InnerRectangle.Width, fontInfo.FontSize/1.8, tmpData, "", "L", false)
}

func (pdf *PDF) setBegin(con gui.WorshipInfo, lines int) {
	var tmpEl string

	for i, _ := range pdf.BibleVerse {
		// 첫 번째 콘텐츠에 추가 정보 삽입 (중복 방지)
		if i == 0 && !strings.Contains(pdf.BibleVerse[i], con.Obj) {
			pdf.BibleVerse[i] = fmt.Sprintf("%s\n%s", con.Obj, pdf.BibleVerse[i])
		}

		// 텍스트 추가
		tmpEl += pdf.BibleVerse[i] + "\n"

		// 페이지 처리 조건
		if (i+1)%lines == 0 || i == len(pdf.BibleVerse)-1 {
			pdf.SetXY(pdf.Config.Padding, pdf.Config.Padding*2)
			pdf.MultiCell(pdf.Config.InnerRectangle.Width, pdf.Config.FontSize/2, tmpEl, "", "L", false)
			tmpEl = ""

			// 다음 페이지 추가
			if i != len(pdf.BibleVerse)-1 {
				pdf.AddPage()
				pdf.CheckImgPlaced(pdf.Path, 0)
			}
		}
	}
}

func (pdf *PDF) setBody(lines int) {
	var tmpEl string

	for i, content := range pdf.BibleVerse {
		tmpEl += content + "\n\n"

		// 페이지 처리 조건
		if (i+1)%lines == 0 || i == len(pdf.BibleVerse)-1 {
			pdf.SetXY(pdf.Config.Padding, pdf.Config.Padding*2)
			pdf.MultiCell(pdf.Config.InnerRectangle.Width, pdf.Config.FontSize/2, tmpEl, "", "L", false)
			tmpEl = ""

			// 마지막 줄이 아니라면 새로운 페이지 추가
			if i != len(pdf.BibleVerse)-1 {
				pdf.AddPage()
				pdf.CheckImgPlaced(pdf.Path, 0)
			}
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
	extStandard := ".png"
	files, err := os.ReadDir(imageDir)
	if err != nil {
		return fmt.Errorf("이미지 디렉토리 읽기 실패: %v", err)
	}

	var imageFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == extStandard {
			imageFiles = append(imageFiles, filepath.Join(imageDir, file.Name()))
		}
	}

	// 숫자 기준 정렬
	sort.Slice(imageFiles, func(i, j int) bool {
		baseName1 := strings.TrimSuffix(filepath.Base(imageFiles[i]), extStandard)
		baseName2 := strings.TrimSuffix(filepath.Base(imageFiles[j]), extStandard)

		num1, _ := strconv.Atoi(baseName1)
		num2, _ := strconv.Atoi(baseName2)

		return num1 < num2
	})

	for _, imgPath := range imageFiles {
		pdf.AddPage()
		pdf.CheckImgPlaced(imgPath, 0)
	}

	return nil
}
