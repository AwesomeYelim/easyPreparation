package presentation

import (
	"easyPreparation_1.0/internal/lyrics"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/presentation"
	"log"
)

// CreatePresentation 함수는 프레젠테이션을 생성하고 슬라이드를 추가합니다.
func CreatePresentation(slidesData *lyrics.SlideData, filePath string) {
	ppt := presentation.New()
	defer ppt.Close()

	// 슬라이드 크기 설정 (예: 960x540, 16:9 비율)
	var slideWidth measurement.Distance = 960.0
	var slideHeight measurement.Distance = 540.0

	// 텍스트 상자를 중앙에 배치하기 위한 함수
	centerPosition := func(slideWidth, slideHeight, boxWidth, boxHeight measurement.Distance) (measurement.Distance, measurement.Distance) {
		xPos := (slideWidth - boxWidth) / 2
		yPos := (slideHeight - boxHeight) / 2
		return xPos, yPos
	}

	for _, content := range slidesData.Content {
		slide := ppt.AddSlide()

		// 제목 설정
		titleBox := slide.AddTextBox()
		titlePara := titleBox.AddParagraph()
		titleRun := titlePara.AddRun()
		titleRun.SetText(slidesData.Title)

		// 제목 상자 크기 설정
		var titleBoxWidth measurement.Distance = 300.0
		var titleBoxHeight measurement.Distance = 50.0
		titleXPos, titleYPos := centerPosition(slideWidth, slideHeight/4, titleBoxWidth, titleBoxHeight) // 제목은 상단에 배치
		titleBox.Properties().SetPosition(titleXPos, titleYPos)
		titleBox.Properties().SetSize(titleBoxWidth, titleBoxHeight)

		// 내용 설정
		contentBox := slide.AddTextBox()
		contentPara := contentBox.AddParagraph()
		contentRun := contentPara.AddRun()
		contentRun.SetText(content)

		// 내용 상자 크기 설정
		var contentBoxWidth measurement.Distance = 600.0
		var contentBoxHeight measurement.Distance = 400.0
		contentXPos, contentYPos := centerPosition(slideWidth, slideHeight, contentBoxWidth, contentBoxHeight) // 내용은 중앙에 배치
		contentBox.Properties().SetPosition(contentXPos, contentYPos)
		contentBox.Properties().SetSize(contentBoxWidth, contentBoxHeight)
	}

	// 프레젠테이션 저장
	if err := ppt.SaveToFile(filePath); err != nil {
		log.Fatalf("Error saving presentation: %v", err)
	}
}
