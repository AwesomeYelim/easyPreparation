package pkg

import (
	"os"
	"strings"
)

func SplitTwoLines(text string) (result []string) {
	// 공백 제거
	lines := RemoveEmptyLines(text)
	// 두 줄씩 자르기
	for lineIndex := 0; lineIndex < len(lines); lineIndex += 2 {
		if lineIndex+1 < len(lines) {
			result = append(result, lines[lineIndex]+"\n"+lines[lineIndex+1])
		} else {
			result = append(result, lines[lineIndex]) // 마지막줄 홀수 인경우에만
		}
	}

	return result
}

// RemoveEmptyLines 함수는 중간 공백을 제거합니다.
func RemoveEmptyLines(text string) []string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, strings.TrimSpace(line))
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
