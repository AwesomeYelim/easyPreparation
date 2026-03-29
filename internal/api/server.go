package api

import (
	"easyPreparation_1.0/internal/handlers"
	"easyPreparation_1.0/internal/middleware"
	"easyPreparation_1.0/internal/types"
	"fmt"
	"net/http"
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
	mux.Handle("/display/order", middleware.CORS(http.HandlerFunc(handlers.DisplayOrderHandler)))
	mux.Handle("/display/append", middleware.CORS(http.HandlerFunc(handlers.DisplayAppendHandler)))
	mux.Handle("/display/remove", middleware.CORS(http.HandlerFunc(handlers.DisplayRemoveHandler)))
	mux.Handle("/display/navigate", middleware.CORS(http.HandlerFunc(handlers.DisplayNavigateHandler)))
	mux.Handle("/display/push", middleware.CORS(http.HandlerFunc(handlers.DisplayPushHandler)))
	mux.Handle("/display/jump", middleware.CORS(http.HandlerFunc(handlers.DisplayJumpHandler)))
	mux.Handle("/display/status", middleware.CORS(http.HandlerFunc(handlers.DisplayStatusHandler)))
	mux.Handle("/display/timer", middleware.CORS(http.HandlerFunc(handlers.DisplayTimerHandler)))
	mux.Handle("/display/lyrics-order", middleware.CORS(http.HandlerFunc(handlers.DisplayLyricsOrderHandler)))

	// 통합 API (프론트 DB 직접 연결 제거)
	mux.Handle("/api/bible/books", middleware.CORS(http.HandlerFunc(handlers.BibleBooksHandler)))
	mux.Handle("/api/bible/versions", middleware.CORS(http.HandlerFunc(handlers.BibleVersionsHandler)))
	mux.Handle("/api/bible/search", middleware.CORS(http.HandlerFunc(handlers.BibleSearchHandler)))
	mux.Handle("/api/bible/verses", middleware.CORS(http.HandlerFunc(handlers.BibleVersesHandler)))
	mux.Handle("/api/user", middleware.CORS(http.HandlerFunc(handlers.UserHandler)))
	mux.Handle("/api/auth/signin", middleware.CORS(http.HandlerFunc(handlers.AuthSignInHandler)))

	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		panic(err)
	}
}
