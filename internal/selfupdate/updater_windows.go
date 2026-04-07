//go:build windows

package selfupdate

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// applyBinary — Windows에서 실행 중인 바이너리를 새 버전으로 교체합니다.
//
// Windows는 실행 중인 .exe를 직접 교체할 수 없으므로 batch script를 생성하여
// 프로세스 종료 후 교체합니다.
//
// 순서:
//  1. update.bat 생성 (2초 대기 → 새 바이너리로 교체 → 재시작 → 배치 자기 삭제)
//  2. cmd.exe로 배치 파일 백그라운드 실행
//  3. os.Exit(0)으로 현재 프로세스 종료
func (u *Updater) applyBinary(newPath, execPath, version string) error {
	batPath := filepath.Join(filepath.Dir(execPath), "update.bat")

	// batch script 내용
	// - timeout으로 현재 프로세스 종료 대기
	// - move로 새 바이너리로 교체
	// - 교체된 실행 파일 재시작
	// - batch 파일 자기 삭제
	batContent := fmt.Sprintf(`@echo off
timeout /t 2 /nobreak >nul
move /Y "%s" "%s"
if errorlevel 1 (
    echo 바이너리 교체 실패
    pause
    goto :eof
)
start "" "%s"
del "%%~f0"
`, newPath, execPath, execPath)

	if err := os.WriteFile(batPath, []byte(batContent), 0644); err != nil {
		return fmt.Errorf("update.bat 생성 실패: %w", err)
	}

	log.Printf("[updater] update.bat 생성 완료: %s", batPath)

	// 배치 파일을 백그라운드로 실행 (새 콘솔 창 없이)
	cmd := exec.Command("cmd.exe", "/C", "start", "/b", "", batPath)
	if err := cmd.Start(); err != nil {
		_ = os.Remove(batPath)
		return fmt.Errorf("update.bat 실행 실패: %w", err)
	}

	log.Printf("[updater] 프로세스 종료 후 자동 업데이트 예정 — 재시작합니다")

	// 상태를 restart_required로 표시한 뒤 종료
	u.setState(UpdateStatus{
		State:   StateRestartRequired,
		Version: version,
		Percent: 100,
	})

	// 잠시 대기하여 WS 브로드캐스트가 전송되도록 함
	// 실제 종료는 배치 파일이 대신 처리
	os.Exit(0)
	return nil
}
