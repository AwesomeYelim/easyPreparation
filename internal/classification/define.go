package classification

type Size struct {
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Padding float64 `json:"padding"`
}

type FontInfo struct {
	FontSize   float64 `json:"fontSize"`
	FontFamily string  `json:"fontFamily"`
}

type Color struct {
	BoxColor   string `json:"boxColor"`
	LineColor  string `json:"lineColor"`
	FontColor  string `json:"fontColor"`
	DateColor  string `json:"dateColor"`
	PrintColor string `json:"printColor"`
}
type ResultInfo struct {
	Size
	FontInfo
	InnerRectangle Size  `json:"innerRectangle"`
	Color          Color `json:"color"`
}
type Bulletin struct {
	Print        ResultInfo `json:"print"`
	Presentation ResultInfo `json:"presentation"`
}
type Lyrics struct {
	Presentation ResultInfo `json:"presentation"`
}
type BackgroundInfo struct {
	Bulletin Bulletin `json:"bulletin"`
	Lyrics   Lyrics   `json:"lyrics"`
}
