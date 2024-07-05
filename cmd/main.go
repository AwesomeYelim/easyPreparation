package main

import (
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/presentation"
	"fmt"

	"github.com/unidoc/unioffice/common/license"
)

func init() {
	// UniOffice 라이센스 설정
	err := license.SetMeteredKey("468eb71b0f562ed29385b487b55d413ad506b3c48950ead1de75bd736c7c17c4")
	if err != nil {
		panic(fmt.Sprintf("Failed to set UniOffice license key: %v", err))
	}
}

func main() {
	// 가사 검색
	song := &lyrics.SlideData{}
	song.SearchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", "하나님은 너를 지키시는자", false)

	// 프레젠테이션 생성 및 슬라이드 추가
	presentation.CreatePresentation(song, "output.pptx")

	fmt.Println("Presentation saved to output.pptx")
}
