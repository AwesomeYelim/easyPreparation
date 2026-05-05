package handlers

import (
	"encoding/json"
	"easyPreparation_1.0/internal/path"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// desktopDownloadDir — Desktop 모드에서 파일을 저장할 디렉터리 (비어있으면 비활성)
var desktopDownloadDir string

// SetDesktopMode — cmd/desktop startup에서 호출. Downloads 폴더를 지정합니다.
func SetDesktopMode(downloadDir string) {
	desktopDownloadDir = downloadDir
}

// findPresentationPDF — target에 해당하는 프레젠테이션 PDF 경로를 반환합니다.
func findPresentationPDF(execPath, target string) (string, error) {
	pdfPath := filepath.Join(execPath, "output", "bulletin", "presentation", fmt.Sprintf("%s.pdf", target))
	if _, err := os.Stat(pdfPath); err != nil {
		return "", fmt.Errorf("PDF 파일이 없습니다: %s", target)
	}
	return pdfPath, nil
}

// DownloadPDFHandler — PDF 직접 서빙 (ZIP 아님)
func DownloadPDFHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	execPath := path.ExecutePath("easyPreparation")

	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	pdfPath, err := findPresentationPDF(execPath, target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", target))
	http.ServeFile(w, r, pdfPath)
}

// SaveToDownloadsHandler — Desktop 모드 전용: PDF를 ~/Downloads에 저장 후 폴더 열기
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

	execPath := path.ExecutePath("easyPreparation")
	pdfPath, err := findPresentationPDF(execPath, target)
	if err != nil {
		log.Printf("[download] SaveToDownloads PDF 없음: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// PDF 파일 복사
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		http.Error(w, "파일 읽기 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	savePath := filepath.Join(desktopDownloadDir, target+".pdf")
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		log.Printf("[download] SaveToDownloads 저장 실패: %v", err)
		http.Error(w, "저장 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[download] SaveToDownloads 저장 완료: %s (%d bytes)", savePath, len(data))

	openFolder(filepath.Dir(savePath))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// openInBrowser — Desktop 모드 공통 브라우저 열기 헬퍼
func openInBrowser(w http.ResponseWriter, targetURL string) {
	if desktopDownloadDir == "" {
		http.Error(w, "not desktop mode", http.StatusForbidden)
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("[download] 브라우저 열기 실패: %v", err)
		http.Error(w, "브라우저 열기 실패", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// OpenDisplayInBrowserHandler — Desktop 모드: 시스템 브라우저에서 display 페이지 열기
func OpenDisplayInBrowserHandler(w http.ResponseWriter, r *http.Request) {
	openInBrowser(w, "http://localhost:8080/display")
}

// OpenMobileInBrowserHandler — Desktop 모드: 시스템 브라우저에서 모바일 리모컨 열기
func OpenMobileInBrowserHandler(w http.ResponseWriter, r *http.Request) {
	openInBrowser(w, "http://localhost:8080/mobile")
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
