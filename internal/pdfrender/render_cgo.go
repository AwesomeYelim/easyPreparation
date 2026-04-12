//go:build cgo

package pdfrender

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	fitz "github.com/gen2brain/go-fitz"
)

// PDFToImages converts a PDF file to numbered PNG images (1.png, 2.png, ...) in outDir.
func PDFToImages(pdfPath, outDir string, dpi int) error {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return fmt.Errorf("PDF 열기 실패: %w", err)
	}
	defer doc.Close()

	for i := 0; i < doc.NumPage(); i++ {
		img, err := doc.ImageDPI(i, float64(dpi))
		if err != nil {
			return fmt.Errorf("페이지 %d 렌더링 실패: %w", i+1, err)
		}
		f, err := os.Create(filepath.Join(outDir, fmt.Sprintf("%d.png", i+1)))
		if err != nil {
			return err
		}
		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()
	}
	return nil
}
