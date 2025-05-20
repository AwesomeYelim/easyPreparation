package handlers

import (
	"archive/zip"
	"bytes"
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
	exeTarget := fmt.Sprintf("%s.pdf", target)

	// PDF file paths to include in the ZIP
	filePaths := []struct {
		Path string
		Name string
	}{
		{
			Path: filepath.Join(execPath, "output", "bulletin", "presentation", exeTarget),
			Name: "presentation_" + exeTarget,
		},
		{
			Path: filepath.Join(execPath, "output", "bulletin", "print", exeTarget),
			Name: "print_" + exeTarget,
		},
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, file := range filePaths {
		BroadcastProgress("Processing file", 1, fmt.Sprintf("Processing file: %s", file.Path))

		f, err := os.Open(file.Path)
		if err != nil {
			BroadcastProgress("Failed to open file", -1, fmt.Sprintf("Failed to open file: %s", file.Path))
			continue
		}
		defer f.Close()

		fw, err := zipWriter.Create(file.Name)
		if err != nil {
			BroadcastProgress("Failed to create ZIP file", -1, fmt.Sprintf("Failed to create ZIP entry: %s", file.Name))
			continue
		}

		_, err = io.Copy(fw, f)
		if err != nil {
			BroadcastProgress("Failed to copy file to ZIP", -1, fmt.Sprintf("Failed to copy file to ZIP: %s", file.Name))
			continue
		}

		BroadcastProgress("File added to ZIP", 1, fmt.Sprintf("File added to ZIP: %s", file.Name))

	}

	err := zipWriter.Close()
	if err != nil {
		http.Error(w, "Failed to finalize ZIP archive", http.StatusInternalServerError)
		return
	}

	// Send ZIP response
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", target))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

	_, err = w.Write(buf.Bytes())
	if err != nil {
		BroadcastProgress("Failed to write", -1, fmt.Sprintf("Failed to write response: %v", err))
	}
}
