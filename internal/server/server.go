package server

import (
	"log"
	"net/http"
)

func StartLocalServer(port, buildFolder string) {
	http.Handle("/", http.FileServer(http.Dir(buildFolder)))
	go func() {
		log.Fatal(http.ListenAndServe(port, nil))
	}()
}
