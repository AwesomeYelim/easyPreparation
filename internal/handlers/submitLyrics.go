package handlers

import (
	"easyPreparation_1.0/internal/types"
	middleware "easyPreparation_1.0/internal/middleware"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/sanitize"
	"easyPreparation_1.0/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func SubmitLyricsHandler(dataChan chan types.DataEnvelope) http.Handler {
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
		dataChan <- types.DataEnvelope{
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
				lyrics, _ := songMap["lyrics"].(string)
				// lyricsPDF.go와 동일 조건: 문자(한글 포함)가 있어야 PDF 생성됨
				if title == "" || len(utils.RemoveEmptyNonLetterLines(lyrics, 25)) == 0 {
					continue
				}
				expectedFiles = append(expectedFiles, filepath.Join(outputDir, sanitize.FileName(title)+".pdf"))
			}
		}
		if len(expectedFiles) == 0 {
			http.Error(w, "생성 가능한 가사 데이터가 없습니다", http.StatusBadRequest)
			return
		}

		// 3. 폴링으로 생성 완료 확인 (최대 2분, 타임아웃 시 생성된 파일만 전송)
		deadline := time.After(2 * time.Minute)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

	WAIT_LOOP:
		for {
			select {
			case <-deadline:
				break WAIT_LOOP // 타임아웃: 생성된 파일만 ZIP 전송
			case <-ticker.C:
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

		// 실제로 존재하는 파일만 ZIP에 포함
		var existFiles, fileNames []string
		for _, f := range expectedFiles {
			if _, err := os.Stat(f); err == nil {
				existFiles = append(existFiles, f)
				fileNames = append(fileNames, filepath.Base(f))
			}
		}
		if len(existFiles) == 0 {
			http.Error(w, "PDF 생성 실패", http.StatusInternalServerError)
			return
		}

		zipBytes, err := utils.CreateZipBufferFromFiles(existFiles, fileNames)
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

		// 생성 이력 기록 (songs 포함 — 재활용 가능하게)
		if email, ok := response["email"].(string); ok && email != "" {
			RecordGeneration(email, "lyrics_ppt", "lyrics_bundle.zip", "", "success", response["songs"])
		}

	}))
}
