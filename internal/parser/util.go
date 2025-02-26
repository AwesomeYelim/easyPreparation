package parser

import "strings"

func RemoveLineNumberPattern(text string) string {
	var result strings.Builder
	i := 0

	for i < len(text) {
		if i+2 < len(text) && text[i] >= '0' && text[i] <= '9' && text[i+1] == ':' && text[i+2] >= '0' && text[i+2] <= '9' {
			i += 3
			for i < len(text) && text[i] >= '0' && text[i] <= '9' {
				i++
			}
			for i < len(text) && text[i] == ' ' {
				i++
			}
		} else {
			result.WriteByte(text[i])
			i++
		}
	}

	return strings.TrimSpace(result.String())
}
