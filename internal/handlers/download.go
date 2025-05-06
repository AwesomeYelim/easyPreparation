package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadPDFHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	// PDF 파일 경로 생성 (예시)
	filePath := fmt.Sprintf("./output/bulletin/presentation/%s", target)

	// 파일 열기
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open PDF file: %s", filePath), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 파일을 HTTP 응답으로 전송
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", target))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", getFileSize(filePath)))

	// 파일 내용 복사
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to send PDF file", http.StatusInternalServerError)
	}
}

func getFileSize(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Error getting file size:", err)
		return 0
	}
	return fileInfo.Size()
}
