package pkg

import (
	"os"
	"strings"
	"unicode"
)

func SplitTwoLines(text string) (result []string) {
	lines := RemoveEmptyNonLetterLines(text, 20)
	seen := make(map[string]bool)

	// 중복 가사 제거
	for i := 0; i < len(lines); i += 2 {
		var block string
		if i+1 < len(lines) {
			block = lines[i] + "\n" + lines[i+1]
		} else {
			block = lines[i]
		}

		if !seen[block] {
			result = append(result, block)
			seen[block] = true
		}
	}

	return result
}

func RemoveEmptyNonLetterLines(text string, maxLineLength int) []string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		filtered := ""

		// 문자 및 공백만 남기기
		for _, r := range line {
			if unicode.IsLetter(r) || r == ' ' {
				filtered += string(r)
			}
		}

		if filtered == "" {
			continue
		}

		words := strings.Fields(filtered)
		currentLine := ""

		for _, word := range words {
			//if len([]rune(word)) > maxLineLength {
			//	// 단어 자체가 너무 길면 단독 줄로
			//	if currentLine != "" {
			//		result = append(result, currentLine)
			//		currentLine = ""
			//	}
			//	result = append(result, word)
			//	continue
			//}

			// 새로운 줄에 단어를 추가할 수 있는지 확인
			testLine := word
			if currentLine != "" {
				testLine = currentLine + " " + word
			}

			if len([]rune(testLine)) <= maxLineLength {
				currentLine = testLine
			} else {
				result = append(result, currentLine)
				currentLine = word
			}
		}

		if currentLine != "" {
			result = append(result, currentLine)
		}
	}

	return result
}

func CheckDirIs(dirPath string) (err error) {
	if _, err = os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0700)
	}
	return err
}

func ReplaceDirPath(dirPath, replacePath string) (err error) {
	if _, err = os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(replacePath, 0700)
	}
	return err
}
