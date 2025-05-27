package main

import (
	"easyPreparation_1.0/apiServer/bulletin"
	"easyPreparation_1.0/internal/api"
)

func main() {
	var dataChan = make(chan map[string]interface{}, 100)
	go api.StartServer(dataChan)

	for data := range dataChan {
		go bulletin.CreateBulletin(data)
	}
}
