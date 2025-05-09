package handlers

import (
	"easyPreparation_1.0/internal/path"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadPDFHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	execPath := path.ExecutePath("easyPreparation")

	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	// PDF 파일 경로 생성 (예시)
	filePath := filepath.Join(execPath, "output", "bulletin", "presentation", target)

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
