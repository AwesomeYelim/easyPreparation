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
	mux.Handle("/display/push", middleware.CORS(http.HandlerFunc(handlers.DisplayPushHandler)))

	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		panic(err)
	}
}
