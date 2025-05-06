package main

import (
	"easyPreparation_1.0/executor/bulletin/define"
	"easyPreparation_1.0/executor/bulletin/forPresentation"
	"easyPreparation_1.0/executor/bulletin/forPrint"
	"easyPreparation_1.0/internal/api"
	"easyPreparation_1.0/internal/date"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
)

func main() {
	var dataChan = make(chan map[string]interface{}, 100)

	go api.StartServer(dataChan) // 서버 실행 고루틴

	for data := range dataChan {
		go func(data map[string]interface{}) {
			execPath := path.ExecutePath("easyPreparation")
			fmt.Println("Start Data Process !!")

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
				log.Println("GetNodes Error:", err)
				return
			}

			// target
			target, ok := data["target"].(string)
			if !ok {
				log.Println("target is not a string")
				return
			}

			// targetInfo 파싱
			var targetInfo []map[string]interface{}
			if rawTargetInfo, err := json.Marshal(data["targetInfo"]); err == nil {
				_ = json.Unmarshal(rawTargetInfo, &targetInfo)
			} else {
				log.Println("Failed to parse targetInfo:", err)
				return
			}

			quote.ProcessQuote(target, &targetInfo)

			configPath := filepath.Join(execPath, "config", "custom.json")
			extract.ExtCustomOption(configPath)

			// 파일명 생성: "202411_3.pdf"
			yearMonth, weekFormatted := date.SetDateTitle()
			outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)
			PdfInfo := &define.PdfInfo{
				FigmaInfo:      figmaInfo,
				ExecPath:       execPath,
				Target:         target,
				OutputFilename: outputFilename,
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
			fmt.Println("Finish Data Process !!")
		}(data)
	}
}
