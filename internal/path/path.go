package path

import (
	"log"
	"path/filepath"
	"strings"
)

func ExecutePath(fullPath, baseDir string) string {
	absPath, _ := filepath.Abs(fullPath)
	index := strings.Index(absPath, baseDir)
	if index == -1 {
		log.Printf("기준 디렉터리 (%s) 없음", baseDir)
		return ""
	}
	return absPath[:index+len(baseDir)]
}
