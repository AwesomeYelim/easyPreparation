package size

type Box struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Background struct {
	Print        Box `json:"print"`
	Presentation Box `json:"presentation"`
}

type Size struct {
	Background     Background `json:"background"`
	InnerRectangle Box        `json:"innerRectangle"`
}
