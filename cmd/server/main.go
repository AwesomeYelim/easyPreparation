package main

import (
	"easyPreparation_1.0/internal/api"
	"easyPreparation_1.0/internal/bulletin"
	"easyPreparation_1.0/internal/googleCloud"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/internal/types"
	"log"
	"path/filepath"
)

func main() {
	// DB 연결 — 서버 시작 시 한 번만 초기화
	execPath := path.ExecutePath("easyPreparation")
	dsn, err := quote.LoadDSN(filepath.Join(execPath, "config", "db.json"))
	if err != nil {
		log.Printf("DB 설정 로드 실패 (성경 조회 비활성): %v", err)
	} else {
		if err := quote.InitDB(dsn); err != nil {
			log.Printf("DB 연결 실패 (성경 조회 비활성): %v", err)
		} else {
			handlers.InitAPIDB(quote.GetDB())
			log.Println("DB 연결 성공")
			defer func() { _ = quote.CloseDB() }()
		}
	}

	// OBS WebSocket 연결
	obs.Init(filepath.Join(execPath, "config", "obs.json"))
	defer func() {
		if m := obs.Get(); m != nil {
			m.Disconnect()
		}
	}()

	// Display 상태 복원 (이전 세션)
	handlers.LoadDisplayState()

	// Google Drive 진행 콜백 연결
	googleCloud.ProgressFunc = handlers.BroadcastProgress

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
