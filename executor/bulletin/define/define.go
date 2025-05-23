package define

import "easyPreparation_1.0/internal/figma/get"

type Service interface {
	Create()
}

type PdfInfo struct {
	FigmaInfo      *get.Info
	Target         string
	ExecPath       string
	OutputFilename string
	MarkName       string
}
