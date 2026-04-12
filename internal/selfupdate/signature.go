package selfupdate

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// VerifyChecksum — checksumURL에서 checksums.txt를 다운로드하여
// filePath의 SHA256 해시를 검증합니다.
//
// checksums.txt 형식 (각 줄):
//
//	sha256hexhash  filename
func VerifyChecksum(filePath string, checksumURL string) error {
	// 파일명 추출 (경로 마지막 구성 요소)
	fileName := filePath
	if idx := strings.LastIndexAny(filePath, "/\\"); idx != -1 {
		fileName = filePath[idx+1:]
	}

	// checksums.txt 다운로드
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("checksums.txt 다운로드 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksums.txt 다운로드 오류: HTTP %d", resp.StatusCode)
	}

	// checksums.txt에서 해당 파일명의 해시 찾기
	expectedHash := ""
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// 형식: "sha256hash  filename" 또는 "sha256hash filename"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		name := parts[len(parts)-1]
		// 경로가 포함된 경우 마지막 구성 요소만 비교
		if idx := strings.LastIndexAny(name, "/\\"); idx != -1 {
			name = name[idx+1:]
		}
		if name == fileName {
			expectedHash = strings.ToLower(hash)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("checksums.txt 읽기 오류: %w", err)
	}

	if expectedHash == "" {
		return fmt.Errorf("checksums.txt에서 '%s' 항목을 찾을 수 없습니다", fileName)
	}

	// 실제 파일의 SHA256 계산
	actualHash, err := computeSHA256(filePath)
	if err != nil {
		return fmt.Errorf("SHA256 계산 실패: %w", err)
	}

	if actualHash != expectedHash {
		return fmt.Errorf("체크섬 불일치: 기대값=%s, 실제값=%s", expectedHash, actualHash)
	}

	return nil
}

// computeSHA256 — 파일의 SHA256 해시를 16진수 문자열로 반환합니다.
func computeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("해시 계산 중 읽기 실패: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
