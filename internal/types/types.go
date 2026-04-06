package types

type DataEnvelope struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// WorshipInfo — 예배 순서 항목 구조체 (이전: internal/gui/define.go)
type WorshipInfo struct {
	Key      string        `json:"key"`
	Title    string        `json:"title"`
	Obj      string        `json:"obj"`
	Info     string        `json:"info"`
	Contents string        `json:"contents"`
	Lead     string        `json:"lead"`
	Children []WorshipInfo `json:"children"`
}
