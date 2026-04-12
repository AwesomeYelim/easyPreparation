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

	// 서버 에러 감지 → 다이얼로그 표시 후 앱 종료
	go func() {
		select {
		case err := <-api.ServerError:
			log.Printf("[desktop] 서버 시작 실패: %v", err)
			wailsruntime.WindowShow(ctx)
			wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
				Type:    wailsruntime.ErrorDialog,
				Title:   "easyPreparation 시작 실패",
				Message: err.Error(),
			})
			wailsruntime.Quit(ctx)
			return
		case <-time.After(10 * time.Second):
			// 타임아웃 — 서버가 조용히 실패했을 수 있음
		}
	}()

	// 서버가 준비될 때까지 대기 후 윈도우 표시
	go func() {
		waitForServer("http://localhost:8080")
		log.Println("[desktop] 서버 준비 완료 — 윈도우 표시")
		wailsruntime.WindowShow(ctx)
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
			// 서버 준비될 때까지 로딩 화면 표시 후 자동 전환
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(`<!DOCTYPE html>
<html><head><style>
  body{margin:0;height:100vh;display:flex;align-items:center;justify-content:center;
       background:#0F172A;font-family:Inter,system-ui,sans-serif;color:#fff}
  .wrap{text-align:center}
  .logo{font-size:32px;font-weight:800;letter-spacing:-0.5px;margin-bottom:16px}
  .dot{display:inline-block;width:8px;height:8px;border-radius:50%;background:#3B82F6;
       margin:0 4px;animation:pulse 1.2s ease-in-out infinite}
  .dot:nth-child(2){animation-delay:.2s} .dot:nth-child(3){animation-delay:.4s}
  @keyframes pulse{0%,80%,100%{opacity:.3;transform:scale(.8)}40%{opacity:1;transform:scale(1)}}
  .msg{margin-top:12px;font-size:13px;color:#64748B}
</style></head><body><div class="wrap">
  <div class="logo">easyPreparation</div>
  <div><span class="dot"></span><span class="dot"></span><span class="dot"></span></div>
  <div class="msg">서버를 시작하는 중...</div>
</div><script>
(function check(){
  fetch("http://localhost:8080/display/status",{mode:'no-cors'})
    .then(function(){window.location.replace("http://localhost:8080")})
    .catch(function(){setTimeout(check,500)});
})();
</script></body></html>`))
			}),
		},
	})
	if err != nil {
		log.Fatal("[desktop] Wails 실행 실패:", err)
	}
}
