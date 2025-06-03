package pkg

import (
	"os"
	"strings"
	"unicode"
)

func SplitTwoLines(text string) (result []string) {
	lines := RemoveEmptyNonLetterLines(text)
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

// 중간 공백 및 비문자 제거
func RemoveEmptyNonLetterLines(text string) []string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		filtered := ""

		for _, r := range line {
			if unicode.IsLetter(r) || r == ' ' { // 문자 및 공백만 허용
				filtered += string(r)
			}
		}

		if filtered != "" {
			result = append(result, filtered)
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
