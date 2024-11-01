package contents

import (
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"github.com/jung-kurt/gofpdf/v2"
	"log"
	"path/filepath"
)

func CreateContents() {
	// 주보 사이즈
	bulletinSize := gofpdf.SizeType{
		Wd: 297.0,
		Ht: 210.0,
	}
	imgPath := "./public/images/coverdesign.png"
	outputDir := "./output/bulletin"
	pkg.CheckDirIs(outputDir)

	objPdf := presentation.New(bulletinSize)
	objPdf.AddPage()
	objPdf.CheckImgPath(bulletinSize, imgPath, 0.5)
	err := objPdf.OutputFileAndClose(filepath.Join(outputDir, "1.pdf"))
	if err != nil {
		log.Fatalf("PDF 저장 중 에러 발생: %v", err)
	}
}
