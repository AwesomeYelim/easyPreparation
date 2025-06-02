package handlers

import (
	"archive/zip"
	"easyPreparation_1.0/internal/api/global"
	middleware "easyPreparation_1.0/internal/middlerware"
	"easyPreparation_1.0/internal/path"
	"encoding/json"
	"fmt"
	"io"
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
		tmpZipName := "lyrics_bundle.zip"
		// 4. ZIP 생성
		zipPath := filepath.Join(outputDir, tmpZipName)
		if err := createZipFromFiles(expectedFiles, zipPath); err != nil {
			http.Error(w, "ZIP 생성 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 5. 응답 전송
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", tmpZipName))
		http.ServeFile(w, r, zipPath)

		// 6. 정리 (선택)
		go func() {
			time.Sleep(5 * time.Second)
			_ = os.RemoveAll(zipPath)
		}()
	}))
}

func createZipFromFiles(filePaths []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range filePaths {
		fileToZip, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		info, err := fileToZip.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.Base(file)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			return err
		}
	}

	return nil
}
