package pkg

import (
	"os"
	"strings"
)

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
