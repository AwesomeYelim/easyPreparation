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

// HTML 태그 제거 함수
func RemoveTags(input string) string {
	var output strings.Builder
	inTag := false
	for _, char := range input {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			output.WriteRune(char)
		}
	}
	return strings.TrimSpace(output.String())
}

// 공백 정리 함수 (여러 개의 공백 → 단일 공백)
func NormalizeSpaces(input string) string {
	fields := strings.Fields(input) // 여러 개의 공백을 단일 공백으로 변환
	return strings.Join(fields, " ")
}
