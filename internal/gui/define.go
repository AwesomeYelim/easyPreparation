package gui

type WorshipInfo struct {
	Title    string        `json:"title"`
	Obj      string        `json:"obj"`
	Info     string        `json:"info"`
	Contents string        `json:"contents"`
	Lead     string        `json:"lead"`
	Children []WorshipInfo `json:"children"`
}
