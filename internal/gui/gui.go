package gui

import (
	"embed"
	"fmt"
	"github.com/zserge/lorca"
	"log"
	"os"
	"os/signal"
)

// Embed the HTML file
//
//go:embed index.html
var htmlFile embed.FS

func Connector() (token string, key string) {
	// 임시 파일 생성 (임베드된 HTML 사용)
	tempFile, err := os.CreateTemp("", "index-*.html")
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	// 임베드된 HTML 파일 내용 => 임시 파일에 기록
	htmlContent, err := htmlFile.ReadFile("index.html")
	if err != nil {
		log.Fatalf("Failed to read embedded HTML: %v", err)
	}

	_, err = tempFile.Write(htmlContent)
	if err != nil {
		log.Fatalf("Failed to write to temp file: %v", err)
	}

	defer func() {
		_ = tempFile.Close()
	}()

	// local 경로로 UI 실행
	ui, err := lorca.New("file://"+tempFile.Name(), "", 480, 320, "--remote-allow-origins=*", "--browser=/path/to/chrome")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = ui.Close()
	}()

	_ = ui.Bind("sendTokenAndKey", func(argToken string, argKey string) {
		token = argToken
		key = argKey
		fmt.Printf("Received Token: %s, Key: %s\n", token, key)
		ui.Eval(`document.getElementById("responseMessage").textContent = "Data received successfully!"`)
		_ = ui.Close()
	})

	// 종료 신호 처리
	sigC := make(chan os.Signal)
	signal.Notify(sigC, os.Interrupt)
	select {
	case <-sigC:
	case <-ui.Done():
	}

	return token, key
}
