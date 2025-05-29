package api

import (
	"easyPreparation_1.0/internal/api/global"
	"easyPreparation_1.0/internal/handlers"
	"fmt"
	"net/http"
)

func StartServer(dataChan chan global.DataEnvelope) {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", handlers.WebSocketHandler)
	mux.Handle("/submit", handlers.SubmitHandler(dataChan))
	mux.HandleFunc("/download", handlers.DownloadPDFHandler)
	mux.Handle("/searchLyrics", handlers.SearchLyrics())
	mux.Handle("/submitLyrics", handlers.SubmitLyricsHandler(dataChan))

	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		panic(err)
	}
}
