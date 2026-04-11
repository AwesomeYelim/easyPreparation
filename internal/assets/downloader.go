// Package assets — 원격 서버에서 PDF를 다운로드하고 로컬 캐시를 관리합니다.
package assets

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// AssetBaseURL — PDF 에셋 서빙 URL (Oracle Cloud nginx)
var AssetBaseURL = "http://138.2.119.220/assets"

// DownloadPDF — 로컬 캐시 확인 후 없으면 원격에서 PDF를 다운로드합니다.
// category: "hymn" 또는 "responsive_reading"
// filename: "032.pdf" 등
// cacheDir: 로컬 캐시 디렉토리 (예: data/pdf/hymn/)
//
// 우선순위:
//  1. 로컬 캐시 (cacheDir/filename)
//  2. 원격 서버 (AssetBaseURL/category/filename)
func DownloadPDF(category, filename, cacheDir string) error {
	pdfPath := filepath.Join(cacheDir, filename)

	// 1. 로컬 캐시 확인
	if _, err := os.Stat(pdfPath); err == nil {
		return nil // 캐시 히트
	}

	// 2. 원격 다운로드
	url := fmt.Sprintf("%s/%s/%s", AssetBaseURL, category, filename)
	log.Printf("[assets] R2 다운로드: %s", url)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("R2 다운로드 실패 (%s): %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("R2 다운로드 실패: HTTP %d (%s)", resp.StatusCode, url)
	}

	// cacheDir 생성
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("캐시 디렉토리 생성 실패: %w", err)
	}

	// 임시 파일에 쓴 후 rename (원자적 저장)
	tmpFile, err := os.CreateTemp(cacheDir, ".download-*")
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

	if err := os.Rename(tmpPath, pdfPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("파일 이동 실패: %w", err)
	}

	log.Printf("[assets] 다운로드 완료: %s → %s", url, pdfPath)
	return nil
}
