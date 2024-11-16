package gui

import (
	"fmt"
	"github.com/zserge/lorca"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

func Connector() (token string, key string) {
	// HTML 파일 경로
	htmlFilePath, _ := filepath.Abs("./gui/index.html")

	// 해당 gui 이슈 https://github.com/zserge/lorca/issues/183 참조
	ui, err := lorca.New("file://"+htmlFilePath, "", 480, 320, "--remote-allow-origins=*", "--browser=/path/to/chrome")
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	_ = ui.Bind("sendTokenAndKey", func(argToken string, argKey string) {

		token = argToken
		key = argKey

		fmt.Printf("Received Token: %s, Key: %s\n", token, key)

		// Go에서 js 로 메시지 전달
		ui.Eval(`document.getElementById("responseMessage").textContent = "Data received successfully!"`)

		_ = ui.Close()
	})

	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	return token, key
}
