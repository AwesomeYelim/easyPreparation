package main

import (
	"github.com/zserge/lorca"
	"log"
	"net/url"
)

func main() {
	// 해당 gui 이슈 https://github.com/zserge/lorca/issues/183 참조
	ui, err := lorca.New("data:text/html,"+url.PathEscape(`
	<html>
		<head><title>Hello</title></head>
		<body><h1>Hello, world!</h1></body>
	</html>
	`), "", 480, 320, "--remote-allow-origins=*")
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()
	// Wait until UI window is closed
	<-ui.Done()
}
