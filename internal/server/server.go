package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//func StartLocalServer(port, buildFolder string) {
//	http.Handle("/", http.FileServer(http.Dir(buildFolder)))
//	go func() {
//		log.Fatal(http.ListenAndServe(port, nil))
//	}()
//}

func StartLocalServer(port, buildFolder string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 요청된 파일 경로 설정
		filePath := filepath.Join(buildFolder, r.URL.Path)

		// 파일이 존재하는 경우 그대로 서빙
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}
		http.ServeFile(w, r, filepath.Join(buildFolder, "index.html"))
	})

	log.Printf("서버 실행 중: http://localhost%s\n", port)
	go func() {
		log.Fatal(http.ListenAndServe(port, nil))
	}()
}
