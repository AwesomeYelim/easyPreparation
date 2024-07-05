package presentation

import (
	"easyPreparation_1.0/internal/lyrics"
	"github.com/unidoc/unioffice/presentation"
	"log"
)

// CreatePresentation 함수는 프레젠테이션을 생성하고 슬라이드를 추가합니다.
func CreatePresentation(slidesData *lyrics.SlideData, filePath string) {
	ppt := presentation.New()
	defer ppt.Close()

	for _, content := range slidesData.Content {
		slide := ppt.AddSlide()
		// 제목 설정
		titleBox := slide.AddTextBox()
		titlePara := titleBox.AddParagraph()
		titleRun := titlePara.AddRun()
		titleRun.SetText(slidesData.Title)
		titleBox.Properties().SetPosition(50, 50)
		titleBox.Properties().SetSize(50, 50)

		// 내용 설정
		contentBox := slide.AddTextBox()
		contentPara := contentBox.AddParagraph()
		contentRun := contentPara.AddRun()
		contentRun.SetText(content)
		contentBox.Properties().SetPosition(50, 50)
		contentBox.Properties().SetSize(600, 400)
	}

	if err := ppt.SaveToFile(filePath); err != nil {
		log.Fatalf("Error saving presentation: %v", err)
	}
}
