package api

import (
	"easyPreparation_1.0/internal/handlers"
	"fmt"
	"net/http"
)

func StartServer(dataChan chan map[string]interface{}) {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", handlers.WebSocketHandler)
	mux.Handle("/submit", handlers.SubmitHandler(dataChan))
	mux.HandleFunc("/download", handlers.DownloadPDFHandler)
	mux.Handle("/searchLyrics", handlers.SearchLyrics())

	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		panic(err)
	}
}
