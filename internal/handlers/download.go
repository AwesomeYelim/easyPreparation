package handlers

import (
	"bytes"
	"encoding/json"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// desktopDownloadDir — Desktop 모드에서 파일을 저장할 디렉터리 (비어있으면 비활성)
var desktopDownloadDir string

// SetDesktopMode — cmd/desktop startup에서 호출. Downloads 폴더를 지정합니다.
func SetDesktopMode(downloadDir string) {
	desktopDownloadDir = downloadDir
}

// buildBulletinZip — target에 해당하는 PDF 파일들을 ZIP으로 묶어 반환합니다.
func buildBulletinZip(execPath, target, pdfType string) ([]byte, error) {
	exeTarget := fmt.Sprintf("%s.pdf", target)

	var filePaths []string
	var fileNames []string

	includePresentation := pdfType == "" || pdfType == "presentation" || pdfType == "both"
	includePrint := pdfType == "" || pdfType == "print" || pdfType == "both"

	if includePresentation {
		presPath := filepath.Join(execPath, "output", "bulletin", "presentation", exeTarget)
		if _, err := os.Stat(presPath); err == nil {
			filePaths = append(filePaths, presPath)
			fileNames = append(fileNames, "presentation_"+exeTarget)
		}
	}

	if includePrint {
		printPath := filepath.Join(execPath, "output", "bulletin", "print", exeTarget)
		if _, err := os.Stat(printPath); err == nil {
			filePaths = append(filePaths, printPath)
			fileNames = append(fileNames, "print_"+exeTarget)
		}
	}

	if includePresentation && strings.HasPrefix(target, "sun_") {
		datePart := strings.TrimPrefix(target, "sun_")
		for _, pfx := range []string{"after", "wed"} {
			extraName := fmt.Sprintf("%s_%s.pdf", pfx, datePart)
			extraPath := filepath.Join(execPath, "output", "bulletin", "presentation", extraName)
			if _, err := os.Stat(extraPath); err == nil {
				filePaths = append(filePaths, extraPath)
				fileNames = append(fileNames, pfx+"_presentation_"+extraName)
			}
		}
	}

	if len(filePaths) == 0 {
		return nil, fmt.Errorf("생성된 PDF 파일이 없습니다")
	}
	return utils.CreateZipBufferFromFiles(filePaths, fileNames)
}

func DownloadPDFHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	pdfType := r.URL.Query().Get("type")
	execPath := path.ExecutePath("easyPreparation")

	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	// 단일 타입 요청: ZIP 없이 PDF 직접 서빙
	if pdfType == "print" || pdfType == "presentation" {
		var subDir string
		if pdfType == "print" {
			subDir = "print"
		} else {
			subDir = "presentation"
		}
		pdfPath := filepath.Join(execPath, "output", "bulletin", subDir, fmt.Sprintf("%s.pdf", target))
		if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
			http.Error(w, "PDF 파일이 없습니다", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_%s.pdf\"", pdfType, target))
		http.ServeFile(w, r, pdfPath)
		return
	}

	// both 또는 기본: ZIP으로 묶어서 서빙
	zipBytes, err := buildBulletinZip(execPath, target, pdfType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", target))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipBytes)))

	if _, err = io.Copy(w, bytes.NewReader(zipBytes)); err != nil {
		log.Printf("[download] Failed to write response: %v", err)
	}
}

// SaveToDownloadsHandler — Desktop 모드 전용: PDF ZIP을 ~/Downloads에 저장 후 폴더 열기
// GET /api/save-to-downloads?target=202604_3
func SaveToDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if desktopDownloadDir == "" {
		http.Error(w, "not desktop mode", http.StatusForbidden)
		return
	}
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "target required", http.StatusBadRequest)
		return
	}

	pdfType := r.URL.Query().Get("type")
	execPath := path.ExecutePath("easyPreparation")
	zipBytes, err := buildBulletinZip(execPath, target, pdfType)
	if err != nil {
		log.Printf("[download] SaveToDownloads 빌드 실패: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	savePath := filepath.Join(desktopDownloadDir, target+".zip")
	if err := os.WriteFile(savePath, zipBytes, 0644); err != nil {
		log.Printf("[download] SaveToDownloads 저장 실패: %v", err)
		http.Error(w, "저장 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[download] SaveToDownloads 저장 완료: %s (%d bytes)", savePath, len(zipBytes))

	// 폴더 열기 (OS별)
	openFolder(filepath.Dir(savePath))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// OpenDisplayInBrowserHandler — Desktop 모드: 시스템 브라우저에서 display 페이지 열기
// GET /api/open-display
func OpenDisplayInBrowserHandler(w http.ResponseWriter, r *http.Request) {
	if desktopDownloadDir == "" {
		http.Error(w, "not desktop mode", http.StatusForbidden)
		return
	}
	var cmd *exec.Cmd
	displayURL := "http://localhost:8080/display"
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", displayURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", displayURL)
	default:
		cmd = exec.Command("xdg-open", displayURL)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("[download] 브라우저 열기 실패: %v", err)
		http.Error(w, "브라우저 열기 실패", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func openFolder(dir string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dir)
	case "windows":
		cmd = exec.Command("explorer", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("[download] 폴더 열기 실패: %v", err)
	}
}
