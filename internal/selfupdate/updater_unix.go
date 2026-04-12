//go:build !windows

package selfupdate

import (
	"fmt"
	"log"
	"os"
)

// applyBinary — macOS/Linux에서 실행 중인 바이너리를 새 버전으로 교체합니다.
//
// 순서:
//  1. 현재 실행 파일을 .bak으로 rename (원자적 — 같은 파일시스템이면 inode 유지)
//  2. 새 바이너리를 원래 경로로 rename
//  3. 실행 권한(0755) 설정
//  4. 상태를 restart_required로 변경
//
// 실패 시 .bak에서 원본 복구를 시도합니다.
func (u *Updater) applyBinary(newPath, execPath, version string) error {
	bakPath := execPath + ".bak"

	log.Printf("[updater] 바이너리 교체 시작: %s → %s", newPath, execPath)

	// 1. 현재 실행 파일 → .bak
	if err := os.Rename(execPath, bakPath); err != nil {
		return fmt.Errorf("현재 바이너리 백업 실패: %w", err)
	}

	// 2. 새 바이너리 → 원래 경로
	if err := os.Rename(newPath, execPath); err != nil {
		// 롤백: .bak 복구
		if rollbackErr := os.Rename(bakPath, execPath); rollbackErr != nil {
			log.Printf("[updater] 롤백 실패 (수동 복구 필요): %v", rollbackErr)
		}
		return fmt.Errorf("새 바이너리 이동 실패: %w", err)
	}

	// 3. 실행 권한 설정
	if err := os.Chmod(execPath, 0755); err != nil {
		log.Printf("[updater] 실행 권한 설정 경고 (계속 진행): %v", err)
	}

	log.Printf("[updater] 바이너리 교체 완료 — 재시작 필요")
	u.setState(UpdateStatus{
		State:   StateRestartRequired,
		Version: version,
		Percent: 100,
	})
	return nil
}
