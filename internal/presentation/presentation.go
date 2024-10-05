package presentation

import (
	"easyPreparation_1.0/internal/lyrics"
	"github.com/jung-kurt/gofpdf/v2"
	"log"
	"os"
)

// CreatePresentation 함수는 프레젠테이션을 PDF로 생성하고 슬라이드를 추가합니다.
func CreatePresentation(slidesData *lyrics.SlideData, filePath string) {
	pdfSize := gofpdf.SizeType{
		Wd: 297.0,
		Ht: 167.0,
	}
	// PDF 객체 생성
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P", // "P"는 세로, "L"은 가로
		UnitStr:        "mm",
		Size:           pdfSize,
	})

	// 한글 폰트 등록 (나눔고딕 폰트 파일을 사용하는 예시)
	fontPath := "NotoSansKR-Bold.ttf" // 폰트 파일 경로를 지정
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		log.Fatalf("폰트 파일을 찾을 수 없습니다: %s", err)
	}
	pdf.AddUTF8Font("NotoSansKR-Bold", "", fontPath)
	pdf.SetFont("NotoSansKR-Bold", "", 40) // 등록한 폰트 사용
	// 글씨 색상 변경 (RGB 색상 지정)
	pdf.SetTextColor(255, 255, 255) // 흰색

	// PDF 페이지 추가
	for _, content := range slidesData.Content {
		pdf.AddPage()

		// 배경 이미지 추가 (배경 이미지를 추가하려면 파일이 필요합니다)
		backgroundImage := "background.png"
		if _, err := os.Stat(backgroundImage); err == nil {
			pdf.ImageOptions(backgroundImage, 0, 0, 297, 167, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		} else {
			log.Printf("배경 이미지 파일이 존재하지 않음: %s\n", err)
		}
		var textW float64 = 250
		var textH float64 = 20
		// 텍스트 추가 (내용 설정)
		pdf.SetXY((pdfSize.Wd-textW)/2, (pdfSize.Ht-textH)/2)
		pdf.MultiCell(textW, textH, content, "", "C", false)

	}

	// PDF 저장
	err := pdf.OutputFileAndClose(filePath)
	if err != nil {
		log.Fatalf("PDF 저장 중 에러 발생: %v", err)
	}
}
