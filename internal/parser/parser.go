package parser

import "strings"

func RemoveLineNumberPattern(text string) string {
	// 앞에서부터 숫자와 콜론(:)이 나오면 이를 제거
	i := 0
	for i < len(text) && (text[i] >= '0' && text[i] <= '9' || text[i] == ':') {
		i++
	}
	return strings.TrimSpace(text[i:])
}
