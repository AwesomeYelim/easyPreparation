package contents

type Color struct {
	BoxColor  string `json:"boxColor"`
	LineColor string `json:"lineColor"`
	FontColor string `json:"fontColor"`
	DateColor string `json:"dateColor"`
}
type Box struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Size struct {
	Background     Box `json:"background"`
	InnerRectangle Box `json:"innerRectangle"`
}

type Config struct {
	Color Color `json:"color"`
	Size  Size  `json:"size"`
}
