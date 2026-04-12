// Package assets — 원격 서버에서 PDF/PNG를 다운로드하고 로컬 캐시를 관리합니다.
package assets

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AssetBaseURL — PDF/PNG 에셋 서빙 URL (Oracle Cloud nginx)
var AssetBaseURL = "http://138.2.119.220/assets"

var httpClient = &http.Client{Timeout: 30 * time.Second}

// downloadFile — url을 localPath에 원자적으로 저장합니다. 404/5xx는 에러 반환.
func downloadFile(url, localPath string) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("다운로드 실패 (%s): %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d (%s)", resp.StatusCode, url)
	}

	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, ".download-*")
	if err != nil {
		return fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("파일 저장 실패: %w", err)
	}
	tmpFile.Close()

	if err := os.Rename(tmpPath, localPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("파일 이동 실패: %w", err)
	}
	return nil
}

// DownloadPDF — 로컬 캐시 확인 후 없으면 원격에서 PDF를 다운로드합니다.
// category: "hymn" 또는 "responsive_reading"
// filename: "032.pdf" 등
// cacheDir: 로컬 캐시 디렉토리 (예: data/pdf/hymn/)
func DownloadPDF(category, filename, cacheDir string) error {
	pdfPath := filepath.Join(cacheDir, filename)
	if _, err := os.Stat(pdfPath); err == nil {
		return nil // 캐시 히트
	}

	url := fmt.Sprintf("%s/%s/%s", AssetBaseURL, category, filename)
	log.Printf("[assets] PDF 다운로드: %s", url)

	if err := downloadFile(url, pdfPath); err != nil {
		return fmt.Errorf("PDF 다운로드 실패: %w", err)
	}

	log.Printf("[assets] 다운로드 완료: %s → %s", url, pdfPath)
	return nil
}

// DownloadPNGPages — Oracle Cloud에서 category_pages/NNN/1.png, 2.png... 를 순서대로 다운로드합니다.
// filename: "032.pdf" 형식, cacheDir: 로컬 PNG 캐시 디렉토리
// 반환값: 로컬에 저장된 PNG 파일 경로 목록 (페이지 순서). 서버에 PNG 없으면 빈 슬라이스.
//
// 서버 URL 규칙:
//
//	AssetBaseURL/hymn_pages/032/1.png
//	AssetBaseURL/responsive_reading_pages/001/1.png
func DownloadPNGPages(category, filename, cacheDir string) []string {
	base := strings.TrimSuffix(filename, ".pdf")
	dirURL := fmt.Sprintf("%s/%s_pages/%s", AssetBaseURL, category, base)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil
	}

	var paths []string
	for page := 1; page <= 30; page++ { // 찬송가 최대 페이지 수 (보통 2~4장)
		localName := fmt.Sprintf("%s_%s_%d.png", category, base, page)
		localPath := filepath.Join(cacheDir, localName)

		// 이미 캐시된 파일이면 추가
		if _, err := os.Stat(localPath); err == nil {
			paths = append(paths, localPath)
			continue
		}

		url := fmt.Sprintf("%s/%d.png", dirURL, page)
		if err := downloadFile(url, localPath); err != nil {
			// 404 = 더 이상 페이지 없음, 다운로드 중단
			break
		}
		log.Printf("[assets] PNG 다운로드: %s", url)
		paths = append(paths, localPath)
	}
	return paths
}
