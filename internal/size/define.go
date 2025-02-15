package size

type ResultInfo struct {
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	FontSize float64 `json:"fontSize"`
}

type Background struct {
	Print        ResultInfo `json:"print"`
	Presentation ResultInfo `json:"presentation"`
}

type Size struct {
	Background     Background `json:"background"`
	InnerRectangle ResultInfo `json:"innerRectangle"`
}
