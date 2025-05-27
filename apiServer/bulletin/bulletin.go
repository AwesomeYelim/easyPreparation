package bulletin

import (
	"easyPreparation_1.0/apiServer/bulletin/define"
	"easyPreparation_1.0/apiServer/bulletin/forPresentation"
	"easyPreparation_1.0/apiServer/bulletin/forPrint"
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma"
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

	//  figmaInfo 파싱
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

	figmaInfo := figma.New(&token, &key, execPath)
	if err := figmaInfo.GetNodes(); err != nil {
		handlers.BroadcastProgress("Get Nodes Error", -1, fmt.Sprintf("GetNodes Error: %s", err))
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
	if rawTargetInfo, err := json.Marshal(data["targetInfo"]); err == nil {
		_ = json.Unmarshal(rawTargetInfo, &targetInfo)
	} else {
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
