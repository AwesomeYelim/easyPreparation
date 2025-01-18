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
	"os/exec"
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

// runPnpmBuild 실행 함수
func runPnpmBuild(projectPath string) error {
	fmt.Println("Building UI with pnpm...")

	// pnpm build 명령어 실행
	cmd := exec.Command("pnpm", "build")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run pnpm build: %w", err)
	}

	fmt.Println("UI build completed successfully.")
	return nil
}

// FigmaConnector 함수에서 build/index.html 파일을 로컬 서버로 제공하고 Lorca로 띄우는 방식으로 수정
func FigmaConnector() (target string, figmaInfo *get.Info) {
	execPath, _ := os.Getwd()
	execPath = path.ExecutePath(execPath, "easyPreparation")

	// UI 빌드 실행
	uiBuildPath := filepath.Join(execPath, "ui")
	if err := runPnpmBuild(uiBuildPath); err != nil {
		fmt.Printf("Error running pnpm build: %v\n", err)
		os.Exit(1)
	}

	// 빌드된 React 프로젝트의 경로
	buildFolder := filepath.Join(uiBuildPath, "build")

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
	_ = ui.Bind("sendContentsDate", func(argTarget string, arg []map[string]interface{}) {
		fmt.Printf("Received ContentsDate: %s", arg)
		ui.Eval(`document.getElementById("responseMessage").textContent = "[RECEIVED] : Contents Data received successfully!"`)

		for i, el := range arg {
			title, tIs := el["title"].(string)
			obj, bIs := el["obj"].(string)
			if !tIs || !bIs {
				continue
			}

			if strings.HasSuffix(title, "부름") || strings.HasSuffix(title, "봉독") {
				el["contents"] = quote.GetQuote(obj)
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
