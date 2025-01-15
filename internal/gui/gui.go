package gui

import (
	"easyPreparation_1.0/internal/figma"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"github.com/zserge/lorca"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 로컬 웹 서버로 빌드된 React 파일들을 제공하는 함수
func startLocalServer(buildFolder string) {
	http.Handle("/", http.FileServer(http.Dir(buildFolder)))
	go func() {
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()
}

// FigmaConnector 함수에서 build/index.html 파일을 로컬 서버로 제공하고 Lorca로 띄우는 방식으로 수정
func FigmaConnector() (target string, figmaInfo *get.Info) {
	// 빌드된 React 프로젝트의 경로
	execPath, _ := os.Getwd()
	execPath = path.ExecutePath(execPath, "easyPreparation")
	buildFolder := filepath.Join(execPath, "ui", "build")

	// build 폴더 내의 index.html 경로 설정
	htmlFilePath := filepath.Join(buildFolder, "index.html")

	// 파일이 존재하는지 확인
	if _, err := os.Stat(htmlFilePath); os.IsNotExist(err) {
		log.Fatalf("Failed to find the HTML file at: %v", htmlFilePath)
	}

	// 로컬 서버로 빌드된 React 파일들을 제공
	startLocalServer(buildFolder)

	// 로컬 서버의 URL로 UI 실행
	ui, err := lorca.New("http://localhost:8081", "", 600, 600, "--remote-allow-origins=*")
	if err != nil {
		log.Fatal(err)
	}

	// 데이터 수신 신호를 위한 채널
	dataReceived := make(chan struct{})

	// sendTokenAndKey 함수 바인딩
	_ = ui.Bind("sendTokenAndKey", func(arg map[string]string) {
		token := arg["token"]
		key := arg["key"]
		fmt.Printf("Received Token: %s, Key: %s\n", token, key)

		figmaInfo = figma.New(&token, &key, execPath)
		err = figmaInfo.GetNodes()
		if err != nil {
			ui.Eval(fmt.Sprintf(`document.getElementById("responseMessage").textContent = "[ERROR] : %s"`, err.Error()))
		} else {
			ui.Eval(`document.getElementById("responseMessage").textContent = "[PASS] : The token and key have been verified !"`)
		}
	})

	// sendContentsDate 함수 바인딩
	_ = ui.Bind("sendContentsDate", func(argTarget string, arg []map[string]string) {
		fmt.Printf("Received ContentsDate: %s", arg)
		ui.Eval(`document.getElementById("responseMessage").textContent = "[RECEIVED] : Contents Data received successfully!"`)

		for _, el := range arg {
			title := el["title"]
			content := el["content"]
			if strings.HasSuffix(title, "부름") || strings.HasSuffix(title, "봉독") {
				el["obj"] = quote.GetQuote(content)
			}
		}
		target = argTarget
		sample, _ := json.MarshalIndent(arg, "", "  ")
		_ = pkg.CheckDirIs(filepath.Join(execPath, "config"))
		_ = os.WriteFile(filepath.Join(execPath, "config", target+".json"), sample, 0644)

		dataReceived <- struct{}{}
	})

	// FIXME: 기존 ui 블로킹 버그 채널 통신 ui 창 닫지 않고 리턴되도록..
	select {
	case <-dataReceived:
		return target, figmaInfo
	case <-ui.Done():
		return target, figmaInfo
	}
}
