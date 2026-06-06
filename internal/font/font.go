package font

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed fonts/*.ttf
var embeddedFonts embed.FS

// GetFontBytes — 내장 폰트를 바이트 배열로 반환합니다.
// OS별 경로 문제 없이 바이너리에 포함된 폰트를 직접 사용합니다.
func GetFontBytes(name, weight string) (fontName string, data []byte, err error) {
	saveName := strings.Replace(name, " ", "", 1)
	saveName = fmt.Sprintf("%s-%s.ttf", saveName, weight)

	data, err = embeddedFonts.ReadFile("fonts/" + saveName)
	if err != nil {
		return "", nil, fmt.Errorf("내장 폰트 없음 (%s): %w", saveName, err)
	}
	return saveName, data, nil
}

// GetFontByFileName — 파일명으로 내장 폰트를 바이트 배열로 반환합니다.
// 썸네일 등에서 파일명 기반으로 폰트를 로드할 때 사용합니다.
func GetFontByFileName(fileName string) ([]byte, error) {
	data, err := embeddedFonts.ReadFile("fonts/" + fileName)
	if err != nil {
		return nil, fmt.Errorf("내장 폰트 없음 (%s): %w", fileName, err)
	}
	return data, nil
}
