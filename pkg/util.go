package pkg

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

func SplitTwoLines(text string) (result []string) {
	lines := RemoveEmptyNonLetterLines(text, 25)
	var prev string
	count := 1

	for i := 0; i < len(lines); i += 2 {
		var block string
		if i+1 < len(lines) {
			block = lines[i] + "\n" + lines[i+1]
		} else {
			block = lines[i]
		}

		if block == prev {
			count++
		} else {
			if count > 1 && len(result) > 0 {
				result[len(result)-1] = fmt.Sprintf("%s\n(x%d)", result[len(result)-1], count)
			}
			result = append(result, block)
			prev = block
			count = 1
		}
	}

	// 마지막 반복된 가사 처리
	if count > 1 && len(result) > 0 {
		result[len(result)-1] = fmt.Sprintf("%s\n(x%d)", result[len(result)-1], count)
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
