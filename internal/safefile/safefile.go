package safefile

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// WriteJSON — JSON 파일을 안전하게 저장합니다.
//
// 순서:
//  1. 임시 파일에 먼저 쓰기 (깨진 데이터가 원본을 덮지 않도록)
//  2. JSON 유효성 검증 (빈 배열/빈 객체도 허용, 빈 문자열은 거부)
//  3. 현재 파일을 .backup으로 복사 (롤백용)
//  4. 임시 파일을 원본으로 rename (같은 파일시스템이면 원자적)
func WriteJSON(filePath string, data interface{}) error {
	// 1. JSON 직렬화
	marshaled, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 직렬화 실패: %w", err)
	}

	// 2. 빈 데이터 방지
	if len(marshaled) < 2 {
		return fmt.Errorf("빈 데이터 저장 거부: %s", filePath)
	}

	// 3. 임시 파일에 쓰기
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // 실패 시 정리

	if _, err := tmpFile.Write(marshaled); err != nil {
		tmpFile.Close()
		return fmt.Errorf("임시 파일 쓰기 실패: %w", err)
	}
	tmpFile.Close()

	// 4. 쓴 데이터 재검증 (디스크에서 다시 읽어서 JSON 파싱)
	check, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("임시 파일 재읽기 실패: %w", err)
	}
	var verify json.RawMessage
	if err := json.Unmarshal(check, &verify); err != nil {
		return fmt.Errorf("저장된 JSON 검증 실패: %w", err)
	}

	// 5. 현재 파일 → .backup (이미 있으면 덮어쓰기)
	if _, err := os.Stat(filePath); err == nil {
		backupPath := filePath + ".backup"
		if copyErr := copyFile(filePath, backupPath); copyErr != nil {
			log.Printf("[safefile] 백업 생성 경고 (계속 진행): %v", copyErr)
		}
	}

	// 6. 임시 파일 → 원본 (원자적 rename)
	if err := os.Rename(tmpPath, filePath); err != nil {
		// Windows: 대상 파일이 열려있으면 rename 실패 → WriteFile로 직접 덮어쓰기 (원본 삭제 없이)
		if data, readErr := os.ReadFile(tmpPath); readErr == nil {
			if writeErr := os.WriteFile(filePath, data, 0644); writeErr != nil {
				return fmt.Errorf("파일 교체 실패: %w", err)
			}
		} else {
			return fmt.Errorf("파일 교체 실패: %w", err)
		}
	}

	return nil
}

// ReadJSONWithRecovery — JSON 파일을 읽되, 실패하면 .backup에서 자동 복구합니다.
func ReadJSONWithRecovery(filePath string) ([]byte, error) {
	data, err := readAndValidate(filePath)
	if err == nil {
		return data, nil
	}
	// 파일 자체가 없으면 복구 대상이 아님 (첫 실행·선택적 설정 파일) — 호출자가 기본값으로 처리
	if os.IsNotExist(err) {
		return nil, err
	}

	// 원본 손상 → .backup 시도
	backupPath := filePath + ".backup"
	log.Printf("[safefile] 원본 읽기 실패 (%v) — 백업에서 복구 시도: %s", err, backupPath)

	backupData, backupErr := readAndValidate(backupPath)
	if backupErr != nil {
		return nil, fmt.Errorf("원본(%v)과 백업(%v) 모두 읽기 실패", err, backupErr)
	}

	// 백업에서 원본 복구
	if writeErr := os.WriteFile(filePath, backupData, 0644); writeErr != nil {
		log.Printf("[safefile] 백업에서 원본 복구 실패: %v", writeErr)
	} else {
		log.Printf("[safefile] 백업에서 원본 복구 완료: %s", filePath)
	}

	return backupData, nil
}

func readAndValidate(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 2 {
		return nil, fmt.Errorf("파일이 비어있음: %s", path)
	}
	var check json.RawMessage
	if err := json.Unmarshal(data, &check); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
	}
	return data, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
