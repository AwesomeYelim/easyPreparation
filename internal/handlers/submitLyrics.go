package handlers

import (
	"easyPreparation_1.0/internal/api/global"
	middleware "easyPreparation_1.0/internal/middlerware"
	"easyPreparation_1.0/internal/path"
	ziputil "easyPreparation_1.0/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func SubmitLyricsHandler(dataChan chan global.DataEnvelope) http.Handler {
	return middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		execPath := path.ExecutePath("easyPreparation")

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var response map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		BroadcastProgress("Response SearchLyrics", 1, fmt.Sprintf("Response SearchLyrics: %+v", response))

		// 1. 백그라운드로 처리 시작
		dataChan <- global.DataEnvelope{
			Type:    "submitLyrics",
			Payload: response,
		}

		// 2. songs 파싱
		rawSongs, ok := response["songs"].([]interface{})
		if !ok {
			http.Error(w, "songs 형식이 잘못되었습니다", http.StatusBadRequest)
			return
		}

		var expectedFiles []string
		outputDir := filepath.Join(execPath, "output", "lyrics")

		for _, item := range rawSongs {
			if songMap, ok := item.(map[string]interface{}); ok {
				title, _ := songMap["title"].(string)
				expectedFiles = append(expectedFiles, filepath.Join(outputDir, fmt.Sprintf("%s.pdf", title)))
			}
		}
		// 3. 폴링으로 생성 완료 확인
		timeout := time.After(5 * time.Minute)
		ticker := time.Tick(500 * time.Millisecond)

	WAIT_LOOP:
		for {
			select {
			case <-timeout:
				http.Error(w, "PDF 생성 시간 초과", http.StatusGatewayTimeout)
				return
			case <-ticker:
				allExist := true
				for _, path := range expectedFiles {
					if _, err := os.Stat(path); os.IsNotExist(err) {
						allExist = false
						break
					}
				}
				if allExist {
					break WAIT_LOOP
				}
			}
		}

		var fileNames []string
		for _, f := range expectedFiles {
			fileNames = append(fileNames, filepath.Base(f))
		}

		zipBytes, err := ziputil.CreateZipBufferFromFiles(expectedFiles, fileNames)
		if err != nil {
			http.Error(w, "ZIP 생성 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=lyrics_bundle.zip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipBytes)))
		if _, err := w.Write(zipBytes); err != nil {
			fmt.Println("ZIP 전송 오류:", err)
		}

	}))
}
