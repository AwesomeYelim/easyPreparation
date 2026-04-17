package presentation

import (
	"easyPreparation_1.0/internal/bulletin/define"
	"easyPreparation_1.0/internal/classification"
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/font"
	"easyPreparation_1.0/internal/format"
	"easyPreparation_1.0/internal/assets"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/parser"
	"easyPreparation_1.0/internal/types"
	"easyPreparation_1.0/internal/utils"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"golang.org/x/text/unicode/norm"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type PDF struct {
	*gofpdf.Fpdf
	Title      string
	BibleVerse []string
	Path       string
	Config     classification.ResultInfo
	*define.PdfInfo
}

type InnerSizeInfo struct {
	Width   float64
	Height  float64
	Padding float64
}

// Text, MultiCell, CellFormat, GetStringWidth — NFC 정규화 래퍼
// macOS에서 한글이 NFD(자모 분리)로 들어올 경우 NFC(완성형)로 변환
func (pdf *PDF) Text(x, y float64, txtStr string) {
	pdf.Fpdf.Text(x, y, norm.NFC.String(txtStr))
}

func (pdf *PDF) MultiCell(w, h float64, txtStr, borderStr, alignStr string, fill bool) {
	pdf.Fpdf.MultiCell(w, h, norm.NFC.String(txtStr), borderStr, alignStr, fill)
}

func (pdf *PDF) CellFormat(w, h float64, txtStr, borderStr string, ln int, alignStr string, fill bool, link int, linkStr string) {
	pdf.Fpdf.CellFormat(w, h, norm.NFC.String(txtStr), borderStr, ln, alignStr, fill, link, linkStr)
}

func (pdf *PDF) GetStringWidth(s string) float64 {
	return pdf.Fpdf.GetStringWidth(norm.NFC.String(s))
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
		imgType := "PNG"
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" {
			imgType = "JPG"
		}
		switch place {
		case 0:
			pdf.ImageOptions(path, 0, 0, pdf.Config.Width, pdf.Config.Height, false, gofpdf.ImageOptions{ImageType: imgType}, 0, "")
		case 0.5:
			pdf.ImageOptions(path, pdf.Config.Width/2, 0, pdf.Config.Width/2, pdf.Config.Height, false, gofpdf.ImageOptions{ImageType: imgType}, 0, "")
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

func (pdf *PDF) ForComposeBuiltin(elements []types.WorshipInfo) (ym float64) {
	// 디자인 기준
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
		if strings.Contains(order.Key, ".") {
			continue
		}
		// 행사인 경우 skip
		if strings.Contains(order.Title, "행사") {
			continue
		}
		pdf.SetXY(xm, ym)
		strOW := pdf.GetStringWidth(order.Obj)
		editLine := (line - (strOW + (lineM * 2))) / 2
		firstPlacedLine := xm + targetWidth + lineM
		secondPlacedLine := firstPlacedLine + editLine + strOW + (lineM * 2)

		if strings.Contains(order.Title, "기도") {
			order.Title = "기도"
		}
		pdf.TextSpacingFormat(order.Title, targetWidth, xm, ym)

		if strings.HasSuffix(order.Info, "c_edit") || strings.HasSuffix(order.Info, "b_edit") {
			pdf.SetXY((pdf.Config.Width/2-pdf.Config.InnerRectangle.Width)/2, ym)
			pdf.DrawLine(editLine, firstPlacedLine, ym, printColor)
			pdf.MultiCell(pdf.Config.InnerRectangle.Width, 0, order.Obj, "", "C", false)
			pdf.DrawLine(editLine, secondPlacedLine, ym, printColor)
		} else {
			pdf.DrawLine(line, firstPlacedLine, ym, printColor)
		}
		pdf.TextSpacingFormat(order.Lead, targetWidth, secondPlacedLine+editLine+lineM, ym)
		ym += fontInfo.FontSize / 1.8
		if order.Title == "축도" {
			break
		}
	}
	return ym
}

func (pdf *PDF) ForReferNext(elements []types.WorshipInfo, nextStart float64) float64 {

	pdf.Config.FontSize *= 0.8
	fontInfo := pdf.Config.FontInfo
	printColor := colorPalette.HexToRGBA(pdf.Config.Color.PrintColor)
	var xm float64 = 427
	var innerBoxWidth float64 = 180
	var isIteratorRange bool

	for idx, element := range elements {
		if element.Title == "축도" {
			isIteratorRange = true
			continue
		}
		if !isIteratorRange {
			continue
		}
		if idx == len(elements)-1 {
			break
		}
		pdf.SetText(fontInfo, true, printColor)
		pdf.SetXY(xm, nextStart)
		pdf.MultiCell(innerBoxWidth, 0, fmt.Sprintf("%s:", element.Title), "", "L", false)
		pdf.SetXY(xm, nextStart)
		pdf.MultiCell(innerBoxWidth, 0, element.Obj, "", "R", false)
		nextStart += fontInfo.FontSize / 2
	}
	return nextStart
}

func (pdf *PDF) ForExtraWorship(label string, elements []types.WorshipInfo, startY float64) float64 {
	fontInfo := pdf.Config.FontInfo
	printColor := colorPalette.HexToRGBA(pdf.Config.Color.PrintColor)
	var xm float64 = 427
	var innerBoxWidth float64 = 180

	startY += fontInfo.FontSize
	pdf.SetText(classification.FontInfo{
		FontSize:   fontInfo.FontSize * 1.1,
		FontFamily: fontInfo.FontFamily,
	}, true, printColor)
	pdf.SetXY(xm, startY)
	pdf.MultiCell(innerBoxWidth, 0, label, "", "L", false)
	startY += fontInfo.FontSize / 1.5

	pdf.SetText(fontInfo, false, printColor)
	for _, el := range elements {
		if strings.Contains(el.Key, ".") {
			continue
		}
		t := el.Title
		if strings.Contains(t, "기도") {
			t = "기도"
		}
		pdf.SetXY(xm, startY)
		pdf.MultiCell(innerBoxWidth, 0, fmt.Sprintf("%s:", t), "", "L", false)
		if el.Obj != "" && el.Obj != "-" {
			pdf.SetXY(xm, startY)
			pdf.MultiCell(innerBoxWidth, 0, el.Obj, "", "R", false)
		}
		startY += fontInfo.FontSize / 2
	}
	return startY
}

func (pdf *PDF) ForTodayVerse(element types.WorshipInfo) {
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

func (pdf *PDF) ForEdit(con types.WorshipInfo, config extract.Config) {
	hLColor := colorPalette.HexToRGBA(pdf.Config.Color.BoxColor) // 박스 색상 설정
	fontInfo := config.Classification.Bulletin.Presentation.FontInfo

	pdf.SetText(fontInfo, true, hLColor)
	//trimmedText := utils.RemoveEmptyLines(con.BibleVerse)
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
		fontInfo.FontSize = fontInfo.FontSize * 0.8
		pdf.DrawChurchNews(fontInfo, con, hLColor, pdf.Config.Padding, pdf.Config.Padding*2.5)
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

func (pdf *PDF) MarkName() {
	labelW := 340.00
	labelWm, labelHm := 13.00, 18.00
	labelP := 11.00

	fontSize := pdf.Config.FontInfo.FontSize / 1.5
	pdf.SetText(classification.FontInfo{
		FontFamily: "Jacques Francois", FontSize: fontSize,
	}, false, color.RGBA{R: 255, G: 255, B: 255})

	x := pdf.Config.Width - (labelW + labelWm + labelP)
	y := pdf.Config.Height - (labelHm + labelP)

	pdf.SetXY(x, y)
	pdf.MultiCell(labelW, 0, pdf.PdfInfo.MarkName, "", "R", false)
}
func (pdf *PDF) DrawChurchNews(fontInfo classification.FontInfo, con types.WorshipInfo, hLColor color.RGBA, x, y float64) {
	// 재귀적으로 교회소식과 그 내부 children 데이터를 처리하는 함수
	var draw func(items []types.WorshipInfo, depth int)

	var tmpData string
	pdf.SetText(fontInfo, false, hLColor)

	draw = func(items []types.WorshipInfo, depth int) {
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
	pdf.MultiCell(pdf.Config.InnerRectangle.Width, fontInfo.FontSize/2.3, tmpData, "", "L", false)
}

func (pdf *PDF) setBegin(con types.WorshipInfo, lines int) {
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
			pdf.SetXY(pdf.Config.Padding, pdf.Config.Padding*2.5)
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
	hLColor := colorPalette.HexToRGBA(pdf.Config.Color.BoxColor) // 박스 색상 설정
	fontInfo := pdf.Config.FontInfo

	for i, content := range pdf.BibleVerse {
		tmpEl += content + "\n\n"

		// 페이지 처리 조건
		if (i+1)%lines == 0 || i == len(pdf.BibleVerse)-1 {
			pdf.SetText(fontInfo, true, hLColor)
			pdf.SetXY(pdf.Config.Padding, pdf.Config.Padding*2.5)
			pdf.MultiCell(pdf.Config.InnerRectangle.Width, pdf.Config.FontSize/2, tmpEl, "", "L", false)
			tmpEl = ""

			// 마지막 줄이 아니라면 새로운 페이지 추가
			if i != len(pdf.BibleVerse)-1 {
				pdf.AddPage()
				pdf.CheckImgPlaced(pdf.Path, 0)
				pdf.MarkName()
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
		// "31장", "495장" → 숫자만 추출 / "사명", "하나님 품으로" → 원문 그대로
		for _, r := range target {
			if r >= '0' && r <= '9' {
				splitNum += string(r)
			}
		}
		if splitNum == "" {
			splitNum = target
		}
	case "responsive_reading":
		splitNum = strings.Split(target, ".")[0]
	}
	// %03d 로 숫자 0-패딩 ("31" → "031.pdf")
	num, err := strconv.Atoi(splitNum)
	var targetNum string
	if err == nil {
		targetNum = fmt.Sprintf("%03d.pdf", num)
	} else {
		targetNum = fmt.Sprintf("%s.pdf", splitNum)
	}

	cacheRoot := filepath.Join(pdf.ExecPath, "data", "cache")
	_ = utils.CheckDirIs(cacheRoot)
	if pngPaths := assets.DownloadPNGPages(category, targetNum, cacheRoot); len(pngPaths) > 0 {
		handlers.BroadcastProgress("PNG cache", 1, fmt.Sprintf("[PNG] %s/%s 적용", category, targetNum))
		for _, imgPath := range pngPaths {
			pdf.AddPage()
			pdf.CheckImgPlaced(imgPath, 0)
		}
		return
	}
	log.Printf("[presentation] PNG 없음 — %s/%s 건너뜀", category, targetNum)
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
