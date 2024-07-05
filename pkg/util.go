package pkg

import "strings"

// RemoveEmptyLines 함수는 중간 공백을 제거합니다.
func RemoveEmptyLines(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, strings.TrimSpace(line))
		}
	}
	return strings.Join(result, "\n")
}
