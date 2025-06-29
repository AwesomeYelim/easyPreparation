package main

import (
	"easyPreparation_1.0/apiServer/bulletin"
	"easyPreparation_1.0/apiServer/lyrics/presentation"
	"easyPreparation_1.0/internal/api"
	"easyPreparation_1.0/internal/api/global"
	"easyPreparation_1.0/internal/handlers"
)

func main() {
	dataChan := make(chan global.DataEnvelope, 100)
	go api.StartServer(dataChan)
	go handlers.StartKeepAliveBroadcast()

	for data := range dataChan {
		switch data.Type {
		case "submit":
			go bulletin.CreateBulletin(data.Payload)
		case "submitLyrics":
			go presentation.CreateLyricsPDF(data.Payload)
		}
	}
}
