package path

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ExecutePath(baseDir string) string {
	fullPath, err := os.Getwd()
	if err != nil {
		log.Printf("작업 디렉터리를 가져오는 중 오류 발생: %v", err)
		return ""
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		log.Printf("절대 경로 변환 오류: %v", err)
		return ""
	}

	if index := strings.Index(absPath, baseDir); index != -1 {
		return absPath[:index+len(baseDir)]
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Printf("실행 파일 경로를 가져오는 중 오류 발생: %v", err)
		return ""
	}

	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		log.Printf("실행 파일 절대 경로 변환 오류: %v", err)
		return ""
	}

	return absExecPath
}
