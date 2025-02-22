package gui

import (
	"easyPreparation_1.0/internal/build"
	"easyPreparation_1.0/internal/figma"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/internal/server"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"github.com/zserge/lorca"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func SetLyricsGui(execPath string) (target map[string]string, figmaInfo *get.Info) {
	var buildFolder string
	buildFolder = filepath.Join(execPath, "build")

	if _, err := os.Stat(buildFolder); os.IsNotExist(err) {
		uiBuildPath := filepath.Join(execPath, "ui", "lyrics")
		buildFolder = build.UiBuild(uiBuildPath, buildFolder)
	}
	port := ":8080"
	server.StartLocalServer(port, buildFolder)
	url := setUrl(port)
	ui, err := setWindowSize(url, 600, 600)
	dataReceived := make(chan struct{})

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

	_ = ui.Bind("sendLyrics", func(arg map[string]string) {
		fmt.Printf("Received Song Title: %s", arg)
		target = arg
		dataReceived <- struct{}{}
	})

	select {
	case <-dataReceived:
		return target, figmaInfo
	case <-ui.Done():
		return target, figmaInfo
	}
}

// build/index.html 파일을 로컬 서버로 제공하고 Lorca로 띄우는 방식
func SetBulletinGui(execPath string) (target string, figmaInfo *get.Info) {
	var buildFolder string
	buildFolder = filepath.Join(execPath, "build")

	if _, err := os.Stat(buildFolder); os.IsNotExist(err) {
		uiBuildPath := filepath.Join(execPath, "ui", "bulletin")
		buildFolder = build.UiBuild(uiBuildPath, buildFolder)
	}

	port := ":8081"
	server.StartLocalServer(port, buildFolder)
	url := setUrl(port)
	ui, err := setWindowSize(url, 600, 600)
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
	_ = ui.Bind("sendContentsDate", func(argTarget string, arg []map[string]interface{}) {
		fmt.Printf("Received ContentsDate: %s", arg)
		ui.Eval(`document.getElementById("responseMessage").textContent = "[RECEIVED] : Contents Data received successfully!"`)

		for i, el := range arg {
			title, tIs := el["title"].(string)
			info, iIs := el["info"].(string)
			obj, bIs := el["obj"].(string)
			if !tIs || !bIs {
				continue
			}
			// 성경 구절 처리
			if iIs && strings.HasPrefix(info, "b_") {
				kor := strings.Split(obj, "_")[0]
				forUrl := strings.Split(obj, "_")[1]

				el["contents"] = quote.GetQuote(forUrl)
				el["obj"] = fmt.Sprintf("%s %s", kor, strings.Split(forUrl, "/")[1])
			}
			if strings.HasSuffix(title, "말씀내용") {
				arg[i]["contents"] = arg[i-1]["contents"]
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

func setUrl(port string) (url string) {
	url = fmt.Sprintf("http://localhost%s", port)
	return url
}

func setWindowSize(url string, wSize, HSize int) (ui lorca.UI, err error) {
	ui, err = lorca.New(url, "", wSize, HSize, "--remote-allow-origins=*")
	if err != nil {
		log.Fatal(err)
	}

	return ui, nil
}
