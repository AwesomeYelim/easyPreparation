package handlers

import (
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/utils"
	"fmt"
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
	exeTarget := fmt.Sprintf("%s.pdf", target)

	var filePaths []string
	var fileNames []string

	// presentation PDF (항상 포함)
	presPath := filepath.Join(execPath, "output", "bulletin", "presentation", exeTarget)
	if _, err := os.Stat(presPath); err == nil {
		filePaths = append(filePaths, presPath)
		fileNames = append(fileNames, "presentation_"+exeTarget)
	}

	// print PDF (있으면 포함 — 주일예배만 생성됨)
	printPath := filepath.Join(execPath, "output", "bulletin", "print", exeTarget)
	if _, err := os.Stat(printPath); err == nil {
		filePaths = append(filePaths, printPath)
		fileNames = append(fileNames, "print_"+exeTarget)
	}

	// 주일예배: 오후/수요 프레젠테이션 추가 포함
	if target == "main_worship" {
		for _, et := range []string{"after_worship", "wed_worship"} {
			extraPath := filepath.Join(execPath, "output", "bulletin", "presentation", et+"_"+exeTarget)
			if _, err := os.Stat(extraPath); err == nil {
				filePaths = append(filePaths, extraPath)
				fileNames = append(fileNames, et+"_presentation_"+exeTarget)
			}
		}
	}

	if len(filePaths) == 0 {
		http.Error(w, "생성된 PDF 파일이 없습니다", http.StatusNotFound)
		return
	}

	zipBytes, err := utils.CreateZipBufferFromFiles(filePaths, fileNames)
	if err != nil {
		http.Error(w, "ZIP 생성 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", target))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipBytes)))

	_, err = w.Write(zipBytes)
	if err != nil {
		BroadcastProgress("Failed to write", -1, fmt.Sprintf("Failed to write response: %v", err))
	}
}
