package parser

import (
	"fmt"
	"strings"
)

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

// verse 정리 함수
// "30:3-30:5", "10:4-10:7", "1:3-3:5", "2:1-2:1"
func CompressVerse(verse string) string {
	parts := strings.Split(verse, "-")
	if len(parts) != 2 {
		return verse // 잘못된 포맷은 그대로 반환
	}

	start := parts[0]
	end := parts[1]

	startParts := strings.Split(start, ":")
	endParts := strings.Split(end, ":")

	if len(startParts) != 2 || (len(endParts) != 1 && len(endParts) != 2) {
		return verse // 포맷 오류
	}

	startChap := startParts[0]
	startVerse := startParts[1]

	var endChap, endVerse string
	if len(endParts) == 2 {
		endChap = endParts[0]
		endVerse = endParts[1]
	} else {
		endChap = startChap // 장 정보가 없으면 시작 장과 같음
		endVerse = endParts[0]
	}

	// 같은 장일 경우 장 번호 생략
	if startChap == endChap {
		return fmt.Sprintf("%s:%s-%s", startChap, startVerse, endVerse)
	}
	return verse
}
