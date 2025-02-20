package forPrint

import (
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/presentation"
	"github.com/jung-kurt/gofpdf/v2"
)

func getSize(config extract.Config) (gofpdf.SizeType, presentation.Size) {
	bulletinSize := gofpdf.SizeType{
		Wd: config.Classification.Bulletin.Print.Width,
		Ht: config.Classification.Bulletin.Print.Height,
	}
	rectangle := presentation.Size{
		Width:  config.Classification.Bulletin.Print.InnerRectangle.Width,
		Height: config.Classification.Bulletin.Print.InnerRectangle.Height,
	}

	return bulletinSize, rectangle
}
