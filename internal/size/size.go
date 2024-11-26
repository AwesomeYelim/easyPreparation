package size

type Box struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Size struct {
	Background     Box `json:"background"`
	InnerRectangle Box `json:"innerRectangle"`
}
