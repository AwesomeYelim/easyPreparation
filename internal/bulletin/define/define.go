package define

type Service interface {
	Create()
}

type PdfInfo struct {
	Target         string
	ExecPath       string
	OutputFilename string
	MarkName       string
}
