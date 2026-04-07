// Package app — 서버/Wails 앱 공통 초기화 로직
package app

import (
	"context"
	"easyPreparation_1.0/internal/api"
	"easyPreparation_1.0/internal/googleCloud"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/license"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/internal/selfupdate"
	"easyPreparation_1.0/internal/types"
	"easyPreparation_1.0/internal/youtube"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// Config — Initialize에 전달하는 옵션
type Config struct {
	// FrontendFS — embed된 정적 파일 FS. nil이면 개발 모드(Next.js dev server)
	FrontendFS fs.FS
	// EmbeddedDataFS — embed된 bible.db + 기본 설정 파일 FS. nil이면 개발 모드
	EmbeddedDataFS fs.FS
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

	// embed된 데이터 파일 추출 (첫 실행 시)
	ExtractEmbeddedData(cfg.EmbeddedDataFS, execPath)

	app := &App{
		DataChan: make(chan types.DataEnvelope, 100),
	}

	// 앱 DB 연결 (SQLite)
	dsn, err := quote.LoadDSN(filepath.Join(execPath, "config", "db.json"))
	if err != nil {
		log.Printf("앱 DB 설정 로드 실패: %v", err)
	} else {
		if err := quote.InitDB(dsn); err != nil {
			log.Printf("앱 DB 연결 실패: %v", err)
		} else {
			handlers.InitAPIDB(quote.GetDB())
			log.Println("앱 DB 연결 성공 (SQLite)")
			app.shutdownFns = append(app.shutdownFns, func() { _ = quote.CloseDB() })
		}
	}

	// 성경 DB 연결 (로컬 SQLite)
	biblePath := filepath.Join(execPath, "data", "bible.db")
	if err := quote.InitBibleDB(biblePath); err != nil {
		log.Printf("성경 DB 연결 실패 (성경/찬송 조회 비활성): %v", err)
	} else {
		handlers.InitBibleDB(quote.GetBibleDB())
		log.Println("성경 DB 연결 성공 (SQLite)")
	}

	// 라이선스 초기화 (DB 연결 이후)
	license.Init(quote.GetDB())
	log.Printf("[init] 라이선스 플랜: %s", license.Get().GetPlan())

	// 라이선스 서버 설정 로드 (config/license.json)
	license.LoadServerConfig(filepath.Join(execPath, "config"))

	// 24시간 주기 백그라운드 라이선스 검증 시작
	verifyCtx, verifyCancel := context.WithCancel(context.Background())
	license.StartBackgroundVerification(verifyCtx)
	app.shutdownFns = append(app.shutdownFns, func() { verifyCancel() })

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

	// 업데이트 관련 초기화
	// WS 브로드캐스트 콜백 주입 (순환 참조 방지용 콜백 패턴)
	selfupdate.GetUpdater().SetBroadcast(handlers.BroadcastMessage)
	// 다운로드 디렉토리 설정
	selfupdate.GetUpdater().SetDownloadDir(filepath.Join(execPath, "data", "update"))
	// 이전 업데이트로 남은 .bak 파일 정리
	selfupdate.GetUpdater().CleanupBackup()

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

// ExtractEmbeddedData — embed된 bible.db와 기본 설정 파일을 디스크에 추출합니다.
// 이미 파일이 존재하면 스킵합니다.
func ExtractEmbeddedData(dataFS fs.FS, execPath string) {
	if dataFS == nil {
		return // 개발 모드 — embed 데이터 없음
	}

	// bible.db 추출
	bibleDst := filepath.Join(execPath, "data", "bible.db")
	extractFile(dataFS, "bible.db", bibleDst)

	// 기본 설정 파일 추출
	defaults := []string{
		"bible_info.json",
		"main_worship.json",
		"after_worship.json",
		"wed_worship.json",
		"fri_worship.json",
	}
	for _, name := range defaults {
		dst := filepath.Join(execPath, "config", name)
		extractFile(dataFS, "defaults/"+name, dst)
	}
}

// extractFile — srcFS에서 srcPath를 읽어 dstPath에 저장합니다.
// dstPath가 이미 존재하면 스킵합니다.
func extractFile(srcFS fs.FS, srcPath, dstPath string) {
	if _, err := os.Stat(dstPath); err == nil {
		return // 이미 존재
	}

	data, err := fs.ReadFile(srcFS, srcPath)
	if err != nil {
		log.Printf("[embed] %s 읽기 실패 (스킵): %v", srcPath, err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		log.Printf("[embed] 디렉토리 생성 실패: %v", err)
		return
	}

	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		log.Printf("[embed] %s 쓰기 실패: %v", dstPath, err)
		return
	}

	log.Printf("[embed] 추출 완료: %s (%.1f MB)", filepath.Base(dstPath), float64(len(data))/1024/1024)
}
