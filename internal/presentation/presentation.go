package presentation

import (
	"easyPreparation_1.0/internal/lyrics"
	"github.com/jung-kurt/gofpdf/v2"
	"log"
	"os"
)

type PDF struct {
	*gofpdf.Fpdf
}

func New(size gofpdf.SizeType) PDF {
	// PDF 객체 생성
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P", // "P"는 세로, "L"은 가로
		UnitStr:        "mm",
		Size:           size,
	})

	return PDF{pdf}
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
		objPdf.CheckImgPath(pdfSize, backgroundImage, 1)

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

func (pdf *PDF) CheckImgPath(pdfSize gofpdf.SizeType, path string, place float32) {
	if _, err := os.Stat(path); err == nil {
		switch place {
		case 1:
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
