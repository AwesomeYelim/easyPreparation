package size

type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type FontInfo struct {
	FontSize   float64 `json:"fontSize"`
	FontOption string  `json:"fontOption"`
}
type ResultInfo struct {
	Size
	FontInfo
	InnerRectangle Size `json:"innerRectangle"`
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
