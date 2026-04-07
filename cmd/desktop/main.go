package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"easyPreparation_1.0/internal/api"
	"easyPreparation_1.0/internal/app"
	"easyPreparation_1.0/internal/bulletin"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/internal/types"
	"easyPreparation_1.0/internal/version"
	"easyPreparation_1.0/internal/youtube"
)

// 빌드 시 ldflags로 주입됩니다:
// -X main.Version=v1.0.0 -X main.Commit=abc1234 -X main.BuildTime=2026-01-01T00:00:00Z
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// App — Wails 앱 구조체
type App struct {
	ctx      context.Context
	dataChan chan types.DataEnvelope
}

// startup — Wails WebView가 초기화된 후 호출됨
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	execPath := path.ExecutePath("easyPreparation")

	// embed된 데이터 파일 추출 (첫 실행 시)
	app.ExtractEmbeddedData(getEmbeddedDataFS(), execPath)

	// DB 연결
	dsn, err := quote.LoadDSN(filepath.Join(execPath, "config", "db.json"))
	if err != nil {
		log.Printf("[desktop] DB 설정 로드 실패 (성경 조회 비활성): %v", err)
	} else {
		if err := quote.InitDB(dsn); err != nil {
			log.Printf("[desktop] DB 연결 실패 (성경 조회 비활성): %v", err)
		} else {
			handlers.InitAPIDB(quote.GetDB())
			log.Println("[desktop] DB 연결 성공")
		}
	}

	// OBS WebSocket 연결
	obs.Init(filepath.Join(execPath, "config", "obs.json"))

	// YouTube API 초기화
	youtube.Init(youtube.DefaultOAuthPath(), youtube.DefaultTokenPath())

	// Display 상태 복원 (이전 세션)
	handlers.LoadDisplayState()

	// 스케줄러 초기화
	handlers.InitScheduler()

	// 프론트엔드 정적 파일 서빙 설정 (embed_prod.go / embed_dev.go 분기)
	api.FrontendFS = getFrontendFS()
	if api.FrontendFS != nil {
		log.Println("[desktop] 프론트엔드 정적 파일 서빙 활성화 (embedded)")
	} else {
		log.Println("[desktop] 프론트엔드 정적 파일 서빙 비활성 (Next.js dev server 사용)")
	}

	// HTTP 서버를 goroutine으로 시작
	a.dataChan = make(chan types.DataEnvelope, 100)
	go api.StartServer(a.dataChan)
	go handlers.StartKeepAliveBroadcast()

	// 백그라운드 작업 큐 처리 goroutine
	go a.processDataChan()

	// 서버가 준비될 때까지 대기 후 WebView URL 설정
	go func() {
		waitForServer("http://localhost:8080")
		log.Println("[desktop] 서버 준비 완료 — WebView URL 설정")
		wailsruntime.WindowShow(ctx)
		wailsruntime.BrowserOpenURL(ctx, "http://localhost:8080")
	}()
}

// shutdown — 앱 종료 시 호출됨
func (a *App) shutdown(ctx context.Context) {
	log.Println("[desktop] 앱 종료 중...")

	// HTTP 서버 graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := api.StopServer(shutdownCtx); err != nil {
		log.Printf("[desktop] HTTP 서버 종료 실패: %v", err)
	}

	handlers.StopScheduler()

	if m := obs.Get(); m != nil {
		m.Disconnect()
	}

	if err := quote.CloseDB(); err != nil {
		log.Printf("[desktop] DB 닫기 실패: %v", err)
	}

	if a.dataChan != nil {
		close(a.dataChan)
	}

	log.Println("[desktop] 앱 종료 완료")
}

// processDataChan — bulletin/lyrics 생성 작업을 백그라운드에서 처리
func (a *App) processDataChan() {
	for data := range a.dataChan {
		switch data.Type {
		case "submit":
			go bulletin.CreateBulletin(data.Payload)
		case "submitLyrics":
			go lyrics.CreateLyricsPDF(data.Payload)
		}
	}
}

// waitForServer — HTTP 서버가 응답할 때까지 대기 (최대 5초)
func waitForServer(baseURL string) {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for i := 0; i < 50; i++ {
		resp, err := client.Get(baseURL + "/display/status")
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Println("[desktop] 서버 대기 타임아웃 — 계속 진행")
}

func main() {
	version.Set(Version, Commit, BuildTime)
	log.Printf("easyPreparation %s (commit: %s, built: %s)", Version, Commit, BuildTime)

	app := &App{}

	err := wails.Run(&options.App{
		Title:         "easyPreparation",
		Width:         1400,
		Height:        900,
		MinWidth:      1024,
		MinHeight:     768,
		DisableResize: false,
		Fullscreen:    false,
		StartHidden:   true, // startup에서 서버 준비 후 WindowShow 호출
		OnStartup:     app.startup,
		OnShutdown:    app.shutdown,
		Bind:          []interface{}{app},
		AssetServer: &assetserver.Options{
			// Wails 내장 asset server 대신 Go HTTP 서버(:8080)를 WebView에서 직접 로드
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "http://localhost:8080", http.StatusTemporaryRedirect)
			}),
		},
	})
	if err != nil {
		log.Fatal("[desktop] Wails 실행 실패:", err)
	}
}
