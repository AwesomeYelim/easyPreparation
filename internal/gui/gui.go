package main

import (
	"github.com/zserge/lorca"
	"log"
	"path/filepath"
)

func main() {
	// HTML 파일 경로
	htmlFilePath, _ := filepath.Abs("./gui/index.html")

	// 해당 gui 이슈 https://github.com/zserge/lorca/issues/183 참조
	ui, err := lorca.New("file://"+htmlFilePath, "", 480, 320, "--remote-allow-origins=*", "--browser=/path/to/chrome")
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	<-ui.Done()
}
