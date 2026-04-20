package bulletin

import (
	"easyPreparation_1.0/internal/bulletin/define"
	"easyPreparation_1.0/internal/bulletin/forPresentation"
	"easyPreparation_1.0/internal/bulletin/forPrint"
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"encoding/json"
	"fmt"
	"path/filepath"
)

func CreateBulletin(data map[string]interface{}) {
	execPath := path.ExecutePath("easyPreparation")
	handlers.BroadcastProgress("Start", 1, "Start Data Process !!")

	// target
	target, ok := data["target"].(string)
	if !ok {
		handlers.BroadcastProgress("target string type assertion error", -1, "target is not a string")
		return
	}

	// mark
	mark, ok := data["mark"].(string)
	if !ok {
		handlers.BroadcastProgress("mark string type assertion error", -1, "mark is not a string")
		return
	}

	// targetInfo 파싱
	var targetInfo []map[string]interface{}
	rawTargetInfo, err := json.Marshal(data["targetInfo"])
	if err != nil {
		handlers.BroadcastProgress("TargetInfo parsing error", -1, fmt.Sprintf("Failed to marshal targetInfo: %s", err))
		return
	}
	if err := json.Unmarshal(rawTargetInfo, &targetInfo); err != nil {
		handlers.BroadcastProgress("TargetInfo parsing error", -1, fmt.Sprintf("Failed to parse targetInfo: %s", err))
		return
	}

	// ProcessQuote는 startup()에서 이미 초기화된 bibleDB만 사용
	// 여기서 InitDB/CloseDB 호출하면 bibleDB까지 닫혀서 찬송가·성경 검색이 끊김
	quote.ProcessQuote(target, &targetInfo)

	configPath := filepath.Join(execPath, "config", "custom.json")
	extract.ExtCustomOption(configPath)

	// 파일명 생성: "sun_202411_3.pdf"
	worshipPrefix := map[string]string{
		"main_worship":  "sun",
		"after_worship": "after",
		"wed_worship":   "wed",
		"fri_worship":   "fri",
	}
	prefix := worshipPrefix[target]
	if prefix == "" {
		prefix = target
	}
	yearMonth, weekFormatted := date.SetDateTitle()
	outputFilename := fmt.Sprintf("%s_%s_%s", prefix, yearMonth, weekFormatted)
	outputFilenameExe := fmt.Sprintf("%s.pdf", outputFilename)
	PdfInfo := &define.PdfInfo{
		ExecPath:       execPath,
		Target:         target,
		OutputFilename: outputFilenameExe,
		MarkName:       mark,
	}
	presentationData := forPresentation.PdfInfo{
		PdfInfo: PdfInfo,
	}

	printData := forPrint.PdfInfo{
		PdfInfo: PdfInfo,
	}

	handlers.BroadcastProgress("Print PDF", 1, "A4 인쇄용 PDF 생성 중...")
	printData.Create()
	handlers.BroadcastProgress("Presentation PDF", 1, "프레젠테이션 PDF 생성 중...")
	presentationData.Create()

	handlers.BroadcastProcessDone(target, outputFilename)
	handlers.BroadcastProgress("Finish Data Process", 1, "Finish Data Process !!")

	// 생성 이력 기록
	if email, ok := data["email"].(string); ok && email != "" {
		outputPath := filepath.Join(execPath, "output", "bulletin", "presentation", outputFilenameExe)
		handlers.RecordGeneration(email, "bulletin", outputFilenameExe, outputPath, "success")
	}
}
