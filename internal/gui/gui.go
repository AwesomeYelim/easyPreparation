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
func startLocalServer(port, buildFolder string) {
	http.Handle("/", http.FileServer(http.Dir(buildFolder)))
	go func() {
		log.Fatal(http.ListenAndServe(port, nil))
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

func uiBuild(execPath string) (buildFolder string) {
	// UI 빌드 실행
	uiBuildPath := filepath.Join(execPath, "ui", "bulletin")

	// 환경 변수 확인 -> dev 모드에서만 UI 빌드 실행
	env := os.Getenv("APP_ENV")
	if env == "dev" {
		if err := runPnpmBuild(uiBuildPath); err != nil {
			fmt.Printf("Error running pnpm build: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Skipping UI build (not in dev mode).")
	}

	// 빌드된 React 프로젝트의 경로
	buildFolder = filepath.Join(uiBuildPath, "build")

	// build 폴더 내의 index.html 경로 설정
	htmlFilePath := filepath.Join(buildFolder, "index.html")

	// 파일이 존재하는지 확인
	if _, err := os.Stat(htmlFilePath); os.IsNotExist(err) {
		log.Fatalf("Failed to find the HTML file at: %v", htmlFilePath)
	}

	return buildFolder
}

// FigmaConnector 함수에서 build/index.html 파일을 로컬 서버로 제공하고 Lorca로 띄우는 방식으로 수정
func FigmaConnector() (target string, figmaInfo *get.Info) {
	execPath, _ := os.Getwd()
	execPath = path.ExecutePath(execPath, "easyPreparation")

	buildFolder := uiBuild(execPath)

	// 로컬 서버로 빌드된 React 파일들을 제공
	port := ":8081"
	startLocalServer(port, buildFolder)

	url := fmt.Sprintf("http://localhost%s", port)
	// 로컬 서버의 URL로 UI 실행
	ui, err := lorca.New(url, "", 600, 600, "--remote-allow-origins=*")
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
