package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"easyPreparation_1.0/executor/bulletin/forPresentation"
	"easyPreparation_1.0/executor/bulletin/forPrint"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
)

func main() {
	go handlers.StartServer() // 서버 실행 고루틴

	for data := range handlers.DataChan {
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

			forPrint.CreatePrint(figmaInfo, target, execPath)
			forPresentation.CreatePresentation(figmaInfo, target, execPath)

			fmt.Println("Finish Data Process !!")
			os.Exit(0)
		}(data)
	}
}
