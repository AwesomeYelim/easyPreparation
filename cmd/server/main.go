package main

import (
	"easyPreparation_1.0/internal/api"
	"easyPreparation_1.0/internal/types"
	"easyPreparation_1.0/internal/bulletin"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/lyrics"
)

func main() {
	dataChan := make(chan types.DataEnvelope, 100)
	go api.StartServer(dataChan)
	go handlers.StartKeepAliveBroadcast()

	for data := range dataChan {
		switch data.Type {
		case "submit":
			go bulletin.CreateBulletin(data.Payload)
		case "submitLyrics":
			go lyrics.CreateLyricsPDF(data.Payload)
		}
	}
}
