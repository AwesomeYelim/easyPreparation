package main

import (
	"easyPreparation_1.0/internal/app"
	"easyPreparation_1.0/internal/bulletin"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/version"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// 빌드 시 ldflags로 주입됩니다:
// -X main.Version=v1.0.0 -X main.Commit=abc1234 -X main.BuildTime=2026-01-01T00:00:00Z
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	version.Set(Version, Commit, BuildTime)
	log.Printf("easyPreparation %s (commit: %s, built: %s)", Version, Commit, BuildTime)
	a := app.Initialize(app.Config{
		FrontendFS:     getFrontendFS(),
		EmbeddedDataFS: getEmbeddedDataFS(),
	})
	defer a.Shutdown()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("시그널 수신 (%v), 서버 종료 중...", sig)
		close(a.DataChan)
	}()

	for data := range a.DataChan {
		switch data.Type {
		case "submit":
			go bulletin.CreateBulletin(data.Payload)
		case "submitLyrics":
			go lyrics.CreateLyricsPDF(data.Payload)
		}
	}

	log.Println("서버 종료 완료")
}
