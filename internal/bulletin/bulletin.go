package bulletin

import (
	"easyPreparation_1.0/internal/bulletin/define"
	"easyPreparation_1.0/internal/bulletin/forPresentation"
	"easyPreparation_1.0/internal/bulletin/forPrint"
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	figmapkg "easyPreparation_1.0/internal/figma"
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

	// figmaInfo 파싱
	var key, token string
	if rawFigma, ok := data["figmaInfo"]; ok {
		if figmaMap, ok := rawFigma.(map[string]interface{}); ok {
			if k, ok := figmaMap["key"].(string); ok {
				key = k
			}
			if t, ok := figmaMap["token"].(string); ok {
				token = t
			}
		}
	}

	figmaInfo, err := figmapkg.New(&token, &key, execPath)
	if err != nil {
		handlers.BroadcastProgress("Figma Init Error", -1, fmt.Sprintf("Figma 초기화 실패: %s", err))
		return
	}

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

	quote.ProcessQuote(target, &targetInfo)

	configPath := filepath.Join(execPath, "config", "custom.json")
	extract.ExtCustomOption(configPath)

	// 파일명 생성: "202411_3.pdf"
	yearMonth, weekFormatted := date.SetDateTitle()
	outputFilename := fmt.Sprintf("%s_%s", yearMonth, weekFormatted)
	outputFilenameExe := fmt.Sprintf("%s.pdf", outputFilename)
	PdfInfo := &define.PdfInfo{
		FigmaInfo:      figmaInfo,
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

	printData.Create()
	presentationData.Create()

	handlers.BroadcastProcessDone(target, outputFilename)
	handlers.BroadcastProgress("Finish Data Process", 1, "Finish Data Process !!")

}
