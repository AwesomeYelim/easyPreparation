package global

type DataEnvelope struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}
