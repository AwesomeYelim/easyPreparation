package handlers

import (
	"easyPreparation_1.0/internal/path"
	ziputil "easyPreparation_1.0/internal/utils"
	"fmt"
	"net/http"
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

	filePaths := []string{filepath.Join(execPath, "output", "bulletin", "presentation", exeTarget), filepath.Join(execPath, "output", "bulletin", "print", exeTarget)}
	fileNames := []string{"presentation_" + exeTarget, "print_" + exeTarget}

	zipBytes, err := ziputil.CreateZipBufferFromFiles(filePaths, fileNames)
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
