package size

type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ResultInfo struct {
	Size
	FontSize       float64 `json:"fontSize"`
	InnerRectangle Size    `json:"innerRectangle"`
}

type Background struct {
	Print        ResultInfo `json:"print"`
	Presentation ResultInfo `json:"presentation"`
}

type BackgroundInfo struct {
	Background Background `json:"background"`
}
