// Package app — 서버/Wails 앱 공통 초기화 로직
package app

import (
	"easyPreparation_1.0/internal/api"
	"easyPreparation_1.0/internal/googleCloud"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/license"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/internal/types"
	"easyPreparation_1.0/internal/youtube"
	"io/fs"
	"log"
	"path/filepath"
)

// Config — Initialize에 전달하는 옵션
type Config struct {
	// FrontendFS — embed된 정적 파일 FS. nil이면 개발 모드(Next.js dev server)
	FrontendFS fs.FS
}

// App — 초기화 결과 및 종료 훅을 담는 구조체
type App struct {
	DataChan chan types.DataEnvelope
	// Shutdown — 서버/앱 종료 시 호출할 정리 함수 목록
	shutdownFns []func()
}

// Initialize — DB, OBS, YouTube, Display, 스케줄러를 초기화하고
// HTTP 서버를 백그라운드 goroutine으로 시작합니다.
// 반환된 *App의 DataChan을 이벤트 루프에서 소비하면 됩니다.
func Initialize(cfg Config) *App {
	execPath := path.ExecutePath("easyPreparation")

	app := &App{
		DataChan: make(chan types.DataEnvelope, 100),
	}

	// DB 연결
	dsn, err := quote.LoadDSN(filepath.Join(execPath, "config", "db.json"))
	if err != nil {
		log.Printf("DB 설정 로드 실패 (성경 조회 비활성): %v", err)
	} else {
		if err := quote.InitDB(dsn); err != nil {
			log.Printf("DB 연결 실패 (성경 조회 비활성): %v", err)
		} else {
			handlers.InitAPIDB(quote.GetDB())
			log.Println("DB 연결 성공")
			app.shutdownFns = append(app.shutdownFns, func() { _ = quote.CloseDB() })
		}
	}

	// 라이선스 초기화 (DB 연결 이후)
	license.Init(quote.GetDB())
	log.Printf("[init] 라이선스 플랜: %s", license.Get().GetPlan())

	// OBS WebSocket 연결
	obs.Init(filepath.Join(execPath, "config", "obs.json"))
	app.shutdownFns = append(app.shutdownFns, func() {
		if m := obs.Get(); m != nil {
			m.Disconnect()
		}
	})

	// YouTube API 초기화
	youtube.Init(youtube.DefaultOAuthPath(), youtube.DefaultTokenPath())

	// Display 상태 복원 (이전 세션)
	handlers.LoadDisplayState()

	// 스케줄러 초기화
	handlers.InitScheduler()
	app.shutdownFns = append(app.shutdownFns, handlers.StopScheduler)

	// Google Drive 진행 콜백 연결
	googleCloud.ProgressFunc = handlers.BroadcastProgress

	// 프론트엔드 정적 파일 서빙 설정
	api.FrontendFS = cfg.FrontendFS
	if api.FrontendFS != nil {
		log.Println("프론트엔드 정적 파일 서빙 활성화 (embedded)")
	} else {
		log.Println("프론트엔드 정적 파일 서빙 비활성 (개발 모드 — Next.js dev server 사용)")
	}

	// HTTP 서버 + keepalive broadcast 시작
	go api.StartServer(app.DataChan)
	go handlers.StartKeepAliveBroadcast()

	return app
}

// Shutdown — 등록된 정리 함수를 역순으로 호출합니다.
func (a *App) Shutdown() {
	for i := len(a.shutdownFns) - 1; i >= 0; i-- {
		a.shutdownFns[i]()
	}
}

// RunEventLoop — DataChan을 소비하는 이벤트 루프 (메인 goroutine에서 호출)
// 채널이 닫히면 반환됩니다.
func (a *App) RunEventLoop() {
	// import cycle 방지를 위해 bulletin/lyrics는 호출자에서 처리
	// 이 함수는 채널 소비 패턴의 편의 래퍼만 제공합니다.
	// 실제 핸들러 분기는 cmd/server/main.go에서 수행합니다.
	<-a.DataChan // 채널 닫힘 대기 (이 함수를 직접 사용하는 경우)
}
