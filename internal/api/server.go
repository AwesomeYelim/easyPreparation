package api

import (
	"context"
	"encoding/json"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/license"
	"easyPreparation_1.0/internal/middleware"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/types"
	"easyPreparation_1.0/internal/version"
	"easyPreparation_1.0/internal/youtube"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"
)

// FrontendFS — main.go에서 embed.FS 서브 디렉토리를 설정
// nil이면 정적 파일 서빙 비활성 (개발 모드에서 Next.js dev server 사용)
var FrontendFS fs.FS

// srv — 서버 인스턴스 (StopServer에서 사용)
var srv *http.Server

// ServerError — 서버 시작 에러를 외부로 전달하기 위한 채널
var ServerError = make(chan error, 1)

// StartServer — HTTP 서버를 시작합니다.
// readyCh가 nil이 아니면 리슨 준비 완료 시 닫힙니다.
func StartServer(dataChan chan types.DataEnvelope, readyCh ...chan struct{}) {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", handlers.WebSocketHandler)
	mux.Handle("/submit", handlers.SubmitHandler(dataChan))
	mux.Handle("/download", middleware.CORS(http.HandlerFunc(handlers.DownloadPDFHandler)))
	mux.Handle("/api/save-to-downloads", middleware.CORS(http.HandlerFunc(handlers.SaveToDownloadsHandler)))
	mux.Handle("/api/open-display", middleware.CORS(http.HandlerFunc(handlers.OpenDisplayInBrowserHandler)))
	mux.Handle("/searchLyrics", handlers.SearchLyrics())
	mux.Handle("/submitLyrics", handlers.SubmitLyricsHandler(dataChan))

	// 모바일 PWA 리모컨
	mux.HandleFunc("/mobile", handlers.MobileRemoteHandler)
	mux.HandleFunc("/mobile/manifest.json", handlers.MobileManifestHandler)
	mux.HandleFunc("/mobile/sw.js", handlers.MobileServiceWorkerHandler)
	mux.HandleFunc("/mobile/icon-192.svg", handlers.MobileIconHandler)
	mux.Handle("/mobile/qr.png", middleware.CORS(http.HandlerFunc(handlers.MobileQRHandler)))

	// OBS Browser Source
	mux.Handle("/display", middleware.CORS(http.HandlerFunc(handlers.DisplayHandler)))
	mux.Handle("/display/bg", middleware.CORS(http.HandlerFunc(handlers.DisplayBgHandler)))
	mux.Handle("/display/assets/", middleware.CORS(http.HandlerFunc(handlers.DisplayAssetsHandler)))
	mux.Handle("/display/tmp/", middleware.CORS(http.HandlerFunc(handlers.DisplayTmpHandler)))
	mux.Handle("/display/font/", middleware.CORS(http.HandlerFunc(handlers.DisplayFontHandler)))
	mux.Handle("/display/order", middleware.CORS(http.HandlerFunc(handlers.DisplayOrderHandler)))
	mux.Handle("/display/append", middleware.CORS(http.HandlerFunc(handlers.DisplayAppendHandler)))
	mux.Handle("/display/remove", middleware.CORS(http.HandlerFunc(handlers.DisplayRemoveHandler)))
	mux.Handle("/display/reorder", middleware.CORS(http.HandlerFunc(handlers.DisplayReorderHandler)))
	mux.Handle("/display/church-name", middleware.CORS(http.HandlerFunc(handlers.DisplayChurchNameHandler)))
	mux.Handle("/display/navigate", middleware.CORS(http.HandlerFunc(handlers.DisplayNavigateHandler)))
	mux.Handle("/display/push", middleware.CORS(http.HandlerFunc(handlers.DisplayPushHandler)))
	mux.Handle("/display/jump", middleware.CORS(http.HandlerFunc(handlers.DisplayJumpHandler)))
	mux.Handle("/display/status", middleware.CORS(http.HandlerFunc(handlers.DisplayStatusHandler)))
	mux.Handle("/display/timer", middleware.CORS(http.HandlerFunc(handlers.DisplayTimerHandler)))
	mux.Handle("/display/lyrics-order", middleware.CORS(http.HandlerFunc(handlers.DisplayLyricsOrderHandler)))
	mux.Handle("/display/overlay", middleware.CORS(http.HandlerFunc(handlers.DisplayOverlayHandler)))

	// 통합 API (프론트 DB 직접 연결 제거)
	mux.Handle("/api/bible/books", middleware.CORS(http.HandlerFunc(handlers.BibleBooksHandler)))
	mux.Handle("/api/bible/versions", middleware.CORS(http.HandlerFunc(handlers.BibleVersionsHandler)))
	mux.Handle("/api/bible/search", middleware.CORS(http.HandlerFunc(handlers.BibleSearchHandler)))
	mux.Handle("/api/bible/verses", middleware.CORS(http.HandlerFunc(handlers.BibleVersesHandler)))
	mux.Handle("/api/user", middleware.CORS(http.HandlerFunc(handlers.UserHandler)))

	// 초기 설정 API (Desktop 앱용)
	mux.Handle("/api/setup/status", middleware.CORS(http.HandlerFunc(handlers.SetupStatusHandler)))
	mux.Handle("/api/setup", middleware.CORS(http.HandlerFunc(handlers.SetupHandler)))

	// 찬송가 API
	mux.Handle("/api/hymns", middleware.CORS(http.HandlerFunc(handlers.HymnListHandler)))
	mux.Handle("/api/hymns/search", middleware.CORS(http.HandlerFunc(handlers.HymnSearchHandler)))
	mux.Handle("/api/hymns/detail", middleware.CORS(http.HandlerFunc(handlers.HymnDetailHandler)))

	// 설정 + 이력 API
	mux.Handle("/api/settings", middleware.CORS(http.HandlerFunc(handlers.SettingsHandler)))
	mux.Handle("/api/settings/license", middleware.CORS(http.HandlerFunc(handlers.LicenseHandler)))
	mux.Handle("/api/history", middleware.CORS(http.HandlerFunc(handlers.HistoryHandler)))

	// 예배 순서 API
	mux.Handle("/api/worship-order", middleware.CORS(http.HandlerFunc(handlers.WorshipOrderHandler)))
	mux.Handle("/api/worship-order/list", middleware.CORS(http.HandlerFunc(handlers.WorshipOrderListHandler)))

	// 라이선스 API
	mux.Handle("/api/license", middleware.CORS(http.HandlerFunc(handlers.LicenseStatusHandler)))
	mux.Handle("/api/license/activate", middleware.CORS(http.HandlerFunc(handlers.LicenseActivateHandler)))
	mux.Handle("/api/license/deactivate", middleware.CORS(http.HandlerFunc(handlers.LicenseDeactivateHandler)))
	mux.Handle("/api/license/verify", middleware.CORS(http.HandlerFunc(handlers.LicenseVerifyHandler)))

	// 결제 연동 API (CF Workers + Stripe)
	mux.Handle("/api/license/checkout", middleware.CORS(http.HandlerFunc(handlers.LicenseCheckoutHandler)))
	mux.Handle("/api/license/callback", middleware.CORS(http.HandlerFunc(handlers.LicenseCallbackHandler)))
	mux.Handle("/api/license/portal", middleware.CORS(http.HandlerFunc(handlers.LicensePortalHandler)))
	// 개발모드 전용: 플랜 즉시 변경
	mux.Handle("/api/license/set-plan", middleware.CORS(http.HandlerFunc(handlers.LicenseSetPlanHandler)))

	// 스케줄러 API (Pro)
	mux.Handle("/api/schedule", middleware.FeatureGate(license.FeatureAutoScheduler, handlers.ScheduleHandler))
	mux.Handle("/api/schedule/test", middleware.FeatureGate(license.FeatureAutoScheduler, handlers.ScheduleTestHandler))
	mux.Handle("/api/schedule/stream", middleware.FeatureGate(license.FeatureAutoScheduler, handlers.StreamControlHandler))

	// OBS 연결 설정 + 상태 (Feature gate 없음)
	mux.Handle("/api/obs/connect", middleware.CORS(http.HandlerFunc(handlers.OBSConnectHandler)))
	mux.Handle("/api/obs/status", middleware.CORS(http.HandlerFunc(handlers.OBSStatusHandler)))
	mux.Handle("/api/obs/auto-configure", middleware.CORS(http.HandlerFunc(handlers.OBSAutoConfigureHandler)))

	// OBS 소스 관리 API (Pro)
	mux.Handle("/api/obs/scenes", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSScenesHandler))
	mux.Handle("/api/obs/sources", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSSourcesHandler))
	mux.Handle("/api/obs/logo/upload", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSLogoUploadHandler))
	mux.Handle("/api/obs/logo/apply", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSLogoApplyHandler))
	mux.Handle("/api/obs/camera/devices", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSCameraDevicesHandler))
	mux.Handle("/api/obs/camera/add", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSCameraAddHandler))
	mux.Handle("/api/obs/sources/toggle", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSSourceToggleHandler))
	mux.Handle("/api/obs/sources/remove", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSSourceRemoveHandler))
	mux.Handle("/api/obs/setup-display", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSSetupDisplayHandler))
	mux.Handle("/api/obs/setup-initial", middleware.FeatureGate(license.FeatureOBSControl, handlers.OBSSetupInitialHandler))

	// 버전 + 업데이트 체크 API
	mux.Handle("/api/version", middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		info := version.Get()
		json.NewEncoder(w).Encode(map[string]string{
			"version":   info.Version,
			"commit":    info.Commit,
			"buildTime": info.BuildTime,
		})
	})))
	mux.Handle("/api/update/check", middleware.CORS(http.HandlerFunc(handlers.UpdateCheckHandler)))
	mux.Handle("/api/update/status", middleware.CORS(http.HandlerFunc(handlers.UpdateStatusHandler)))
	mux.Handle("/api/update/download", middleware.CORS(http.HandlerFunc(handlers.UpdateDownloadHandler)))
	mux.Handle("/api/update/apply", middleware.CORS(http.HandlerFunc(handlers.UpdateApplyHandler)))
	mux.Handle("/api/update/cancel", middleware.CORS(http.HandlerFunc(handlers.UpdateCancelHandler)))

	// PDF 에셋 서빙 (138 서버에서 R2 역할 — data/pdf/ 디렉토리 서빙)
	mux.Handle("/api/assets/", middleware.CORS(http.HandlerFunc(handlers.AssetServeHandler)))

	// 배경 템플릿 관리 API
	mux.Handle("/api/templates", middleware.CORS(http.HandlerFunc(handlers.TemplateHandler)))
	mux.Handle("/api/templates/", middleware.CORS(http.HandlerFunc(handlers.TemplateHandler)))

	// 썸네일 API (generate/upload = Pro, 나머지 = 무료)
	mux.Handle("/api/thumbnail/generate", middleware.FeatureGate(license.FeatureThumbnail, handlers.ThumbnailGenerateHandler))
	mux.Handle("/api/thumbnail/preview", middleware.CORS(http.HandlerFunc(handlers.ThumbnailPreviewHandler)))
	mux.Handle("/api/thumbnail/config", middleware.CORS(http.HandlerFunc(handlers.ThumbnailConfigHandler)))
	mux.Handle("/api/thumbnail/upload", middleware.FeatureGate(license.FeatureThumbnail, handlers.ThumbnailUploadHandler))
	mux.Handle("/api/thumbnail/image", middleware.CORS(http.HandlerFunc(handlers.ThumbnailImageHandler)))

	// YouTube API (auth/setup-obs = Pro, callback/status = 무료)
	mux.Handle("/api/youtube/auth", middleware.FeatureGate(license.FeatureYouTube, youtube.AuthHandler))
	mux.Handle("/api/youtube/open-auth", middleware.FeatureGate(license.FeatureYouTube, youtube.OpenAuthHandler))
	mux.Handle("/api/youtube/callback", http.HandlerFunc(youtube.CallbackHandler))
	mux.Handle("/api/youtube/status", middleware.CORS(http.HandlerFunc(youtube.StatusHandler)))

	// YouTube 방송 생성 + 스트림 키 → OBS 자동 세팅 (Pro)
	mux.Handle("/api/youtube/setup-obs", middleware.FeatureGate(license.FeatureYouTube, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		// 제목 + 예배유형 결정
		title := "라이브 예배"
		worshipType := "main_worship"
		var body struct {
			Title       string `json:"title"`
			WorshipType string `json:"worshipType"`
		}
		if json.NewDecoder(r.Body).Decode(&body) == nil {
			if body.Title != "" {
				title = body.Title
			}
			if body.WorshipType != "" {
				worshipType = body.WorshipType
			}
		}

		// YouTube 방송 생성 + 스트림 바인딩
		server, key, broadcastID, err := youtube.CreateBroadcastAndBind(title)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}

		// 방송이 upcoming 상태인 지금 썸네일 생성 + 업로드 (active 되면 YouTube가 덮어씌움)
		go func() {
			handlers.GenerateAndUploadThumbnailTo(worshipType, broadcastID)
		}()

		// OBS 스트리밍 중이면 먼저 중지
		obsM := obs.Get()
		streamStatus := obsM.GetStreamStatus()
		if streamStatus.Active {
			obsM.StopStreaming()
			// 중지 대기
			for i := 0; i < 10; i++ {
				s := obsM.GetStreamStatus()
				if !s.Active {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
		}

		// OBS 스트림 설정
		if err := obsM.SetStreamSettings(server, key); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "방송 생성됨, OBS 설정 실패: " + err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"message":     "YouTube 방송 생성 + OBS 스트림 설정 + 썸네일 업로드 완료",
			"broadcastId": broadcastID,
		})
	}))

	// 정적 파일 서빙 (프로덕션 모드에서 embed된 Next.js static export)
	if FrontendFS != nil {
		mux.Handle("/", spaHandler(FrontendFS))
	}

	// net.Listen으로 포트 바인딩 — 준비 완료 시점 감지 가능
	ln, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		ServerError <- fmt.Errorf("포트 8080이 이미 사용 중입니다.\n다른 easyPreparation 또는 프로그램이 실행 중인지 확인해주세요.\n\n상세: %v", err)
		return
	}

	srv = &http.Server{Handler: mux}

	fmt.Println("Server running on http://localhost:8080")

	// 준비 완료 신호 전송
	if len(readyCh) > 0 && readyCh[0] != nil {
		close(readyCh[0])
	}

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		ServerError <- fmt.Errorf("서버 오류: %v", err)
	}
}

// StopServer — graceful shutdown (Wails 앱 종료 시 호출)
func StopServer(ctx context.Context) error {
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

// spaHandler — SPA fallback handler
// 파일이 존재하면 서빙, 없으면 index.html로 fallback (클라이언트 사이드 라우팅 지원)
func spaHandler(frontFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(frontFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// API나 ws는 여기 안 옴 (이미 위에서 매칭됨)

		// 파일 존재 확인
		tryPath := strings.TrimPrefix(path, "/")
		if tryPath == "" {
			tryPath = "index.html"
		}
		f, err := frontFS.Open(tryPath)
		if err != nil {
			// 파일 없으면 index.html (SPA 라우팅)
			r.URL.Path = "/index.html"
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	})
}
