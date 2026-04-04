package api

import (
	"encoding/json"
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/middleware"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/types"
	"easyPreparation_1.0/internal/youtube"
	"fmt"
	"net/http"
	"time"
)

func StartServer(dataChan chan types.DataEnvelope) {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", handlers.WebSocketHandler)
	mux.Handle("/submit", handlers.SubmitHandler(dataChan))
	mux.Handle("/download", middleware.CORS(http.HandlerFunc(handlers.DownloadPDFHandler)))
	mux.Handle("/searchLyrics", handlers.SearchLyrics())
	mux.Handle("/submitLyrics", handlers.SubmitLyricsHandler(dataChan))

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
	mux.Handle("/api/auth/signin", middleware.CORS(http.HandlerFunc(handlers.AuthSignInHandler)))

	// 찬송가 API
	mux.Handle("/api/hymns", middleware.CORS(http.HandlerFunc(handlers.HymnListHandler)))
	mux.Handle("/api/hymns/search", middleware.CORS(http.HandlerFunc(handlers.HymnSearchHandler)))
	mux.Handle("/api/hymns/detail", middleware.CORS(http.HandlerFunc(handlers.HymnDetailHandler)))

	// 설정 + 이력 API
	mux.Handle("/api/settings", middleware.CORS(http.HandlerFunc(handlers.SettingsHandler)))
	mux.Handle("/api/settings/license", middleware.CORS(http.HandlerFunc(handlers.LicenseHandler)))
	mux.Handle("/api/history", middleware.CORS(http.HandlerFunc(handlers.HistoryHandler)))

	// 스케줄러 API
	mux.Handle("/api/schedule", middleware.CORS(http.HandlerFunc(handlers.ScheduleHandler)))
	mux.Handle("/api/schedule/test", middleware.CORS(http.HandlerFunc(handlers.ScheduleTestHandler)))
	mux.Handle("/api/schedule/stream", middleware.CORS(http.HandlerFunc(handlers.StreamControlHandler)))

	// 썸네일 API
	mux.Handle("/api/thumbnail/generate", middleware.CORS(http.HandlerFunc(handlers.ThumbnailGenerateHandler)))
	mux.Handle("/api/thumbnail/preview", middleware.CORS(http.HandlerFunc(handlers.ThumbnailPreviewHandler)))
	mux.Handle("/api/thumbnail/config", middleware.CORS(http.HandlerFunc(handlers.ThumbnailConfigHandler)))
	mux.Handle("/api/thumbnail/upload", middleware.CORS(http.HandlerFunc(handlers.ThumbnailUploadHandler)))
	mux.Handle("/api/thumbnail/image", middleware.CORS(http.HandlerFunc(handlers.ThumbnailImageHandler)))

	// YouTube API
	mux.Handle("/api/youtube/auth", middleware.CORS(http.HandlerFunc(youtube.AuthHandler)))
	mux.Handle("/api/youtube/callback", http.HandlerFunc(youtube.CallbackHandler))
	mux.Handle("/api/youtube/status", middleware.CORS(http.HandlerFunc(youtube.StatusHandler)))

	// YouTube 방송 생성 + 스트림 키 → OBS 자동 세팅
	mux.Handle("/api/youtube/setup-obs", middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})))

	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		panic(err)
	}
}
