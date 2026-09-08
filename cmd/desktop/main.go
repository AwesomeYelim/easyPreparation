package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"easyPreparation_1.0/internal/app"
	"easyPreparation_1.0/internal/bulletin"
	"easyPreparation_1.0/internal/embedded"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/lyrics"
	"easyPreparation_1.0/internal/selfupdate"
	"easyPreparation_1.0/internal/version"
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
	ctx  context.Context
	core *app.App
}

// startup — Wails WebView가 초기화된 후 호출됨
// 서버/DB/라이선스/OBS/스케줄러 초기화는 internal/app.Initialize 가 담당하고,
// 여기서는 데스크톱 고유 동작(다운로드 경로, 시작 실패 다이얼로그, 헬스체크·롤백, 창 표시)만 처리한다.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Desktop 모드 활성화 — ~/Downloads에 파일 저장 (서버 시작 전에 설정)
	if homeDir, err := os.UserHomeDir(); err == nil {
		handlers.SetDesktopMode(filepath.Join(homeDir, "Downloads"))
	}

	a.core = app.Initialize(app.Config{
		FrontendFS:     embedded.FrontendFS(),
		EmbeddedDataFS: embedded.DataFS(),
		OnServerError: func(err error) {
			log.Printf("[desktop] 서버 시작 실패: %v", err)
			wailsruntime.WindowShow(ctx)
			wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
				Type:    wailsruntime.ErrorDialog,
				Title:   "easyPreparation 시작 실패",
				Message: err.Error(),
			})
			wailsruntime.Quit(ctx)
		},
	})

	// 백그라운드 작업 큐 처리 goroutine
	go a.processDataChan()

	// 서버가 준비될 때까지 대기 → 헬스체크 → 윈도우 표시
	go func() {
		app.WaitForServer(app.LocalBaseURL)
		log.Println("[desktop] 서버 준비 완료")

		healthy := app.RunHealthCheck(app.LocalBaseURL)
		if !healthy && selfupdate.GetUpdater().HasBackup() {
			// 이전 버전 백업이 있고 헬스체크 실패 → 롤백 제안
			wailsruntime.WindowShow(ctx)
			result, _ := wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
				Type:          wailsruntime.QuestionDialog,
				Title:         "문제 감지됨",
				Message:       "앱 상태 검사에서 문제가 발견되었습니다.\n이전 버전으로 되돌리시겠습니까?\n\n(아니오를 선택하면 현재 버전으로 계속합니다)",
				DefaultButton: "No",
				Buttons:       []string{"Yes", "No"},
			})
			if result == "Yes" {
				if err := selfupdate.GetUpdater().Rollback(); err != nil {
					log.Printf("[desktop] 롤백 실패: %v", err)
					wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
						Type:    wailsruntime.ErrorDialog,
						Title:   "롤백 실패",
						Message: "이전 버전으로 복구하지 못했습니다: " + err.Error(),
					})
				} else {
					wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
						Type:    wailsruntime.InfoDialog,
						Title:   "롤백 완료",
						Message: "이전 버전으로 복구되었습니다. 앱을 다시 시작해주세요.",
					})
					wailsruntime.Quit(ctx)
					return
				}
			}
		} else if healthy {
			// 헬스체크 통과 → 이전 백업 정리
			selfupdate.GetUpdater().CleanupBackup()
		}

		wailsruntime.WindowShow(ctx)
	}()
}

// SaveZip — 서버에서 ZIP을 받아 ~/Downloads에 저장 후 Finder로 표시
// 반환값: 에러 메시지 (빈 문자열 = 성공)
func (a *App) SaveZip(target string) string {
	log.Printf("[desktop] SaveZip 호출: %s", target)

	// ZIP 다운로드
	resp, err := http.Get(fmt.Sprintf("%s/download?target=%s", app.LocalBaseURL, target))
	if err != nil {
		log.Printf("[desktop] SaveZip 다운로드 실패: %v", err)
		return "다운로드 실패: " + err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[desktop] SaveZip 서버 오류: %d", resp.StatusCode)
		return fmt.Sprintf("PDF를 찾을 수 없습니다 (%d). 먼저 주보를 생성해주세요.", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "파일 읽기 실패: " + err.Error()
	}
	log.Printf("[desktop] SaveZip ZIP 로드 완료: %d bytes", len(data))

	// ~/Downloads에 저장
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "홈 디렉토리 조회 실패: " + err.Error()
	}
	savePath := filepath.Join(homeDir, "Downloads", target+".zip")

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return "저장 실패: " + err.Error()
	}
	log.Printf("[desktop] SaveZip 저장 완료: %s", savePath)

	// 저장 폴더 열기. Windows 경로(C:\...)는 file:///C:/... 형식이어야 함.
	dir := filepath.Dir(savePath)
	var fileURL string
	if runtime.GOOS == "windows" {
		fileURL = "file:///" + strings.ReplaceAll(dir, "\\", "/")
	} else {
		fileURL = "file://" + dir
	}
	wailsruntime.BrowserOpenURL(a.ctx, fileURL)
	return ""
}

// OpenURL — 시스템 브라우저에서 URL 열기 (파일 다운로드 등 WebView2가 처리 못하는 경우 사용)
func (a *App) OpenURL(url string) {
	wailsruntime.BrowserOpenURL(a.ctx, url)
}

// shutdown — 앱 종료 시 호출됨. 서버 정지·스케줄러·OBS·DB·로그 정리는 app.Shutdown 이 역순으로 수행.
func (a *App) shutdown(ctx context.Context) {
	log.Println("[desktop] 앱 종료 중...")
	if a.core != nil {
		a.core.Shutdown()
		close(a.core.DataChan)
	}
	log.Println("[desktop] 앱 종료 완료")
}

// processDataChan — bulletin/lyrics 생성 작업을 백그라운드에서 처리
func (a *App) processDataChan() {
	for data := range a.core.DataChan {
		switch data.Type {
		case "submit":
			go bulletin.CreateBulletin(data.Payload)
		case "submitLyrics":
			go lyrics.CreateLyricsPDF(data.Payload)
		}
	}
}

func main() {
	version.Set(Version, Commit, BuildTime)
	log.Printf("easyPreparation %s (commit: %s, built: %s)", Version, Commit, BuildTime)

	desktop := &App{}

	err := wails.Run(&options.App{
		Title:         "easyPreparation",
		Width:         1400,
		Height:        900,
		MinWidth:      1400,
		MaxWidth:      1400,
		MinHeight:     768,
		DisableResize: false,
		Fullscreen:    false,
		StartHidden:   true, // startup에서 서버 준비 후 WindowShow 호출
		OnStartup:     desktop.startup,
		OnShutdown:    desktop.shutdown,
		Bind:          []interface{}{desktop},
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
  fetch("` + app.LocalBaseURL + `/display/status",{mode:'no-cors'})
    .then(function(){window.location.replace("` + getUIBaseURL() + `")})
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
