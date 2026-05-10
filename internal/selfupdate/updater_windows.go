//go:build windows

package selfupdate

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// applyBinary — Windows에서 새 버전을 적용합니다.
//
// NSIS 설치파일(_setup.exe)인 경우:
//   - 사일런트 모드(/S)로 실행 → 기존 설치 덮어쓰기
//   - 현재 프로세스 종료 → NSIS가 알아서 설치 완료
//
// 일반 바이너리인 경우:
//   - batch script로 프로세스 종료 후 교체
func (u *Updater) applyBinary(newPath, execPath, version string) error {
	if strings.Contains(filepath.Base(newPath), "_setup.exe") || strings.Contains(filepath.Base(newPath), "installer") {
		return u.applyNSIS(newPath, version)
	}
	return u.applyDirectReplace(newPath, execPath, version)
}

// applyNSIS — NSIS 설치파일을 사일런트 실행합니다.
func (u *Updater) applyNSIS(setupPath, version string) error {
	log.Printf("[updater] NSIS 사일런트 설치 시작: %s", setupPath)

	cmd := exec.Command(setupPath, "/S")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("NSIS 실행 실패: %w", err)
	}

	log.Printf("[updater] NSIS 설치 시작됨 — 현재 앱 종료합니다")

	u.setState(UpdateStatus{
		State:   StateRestartRequired,
		Version: version,
		Percent: 100,
	})

	os.Exit(0)
	return nil
}

// applyDirectReplace — batch script로 바이너리 직접 교체합니다.
func (u *Updater) applyDirectReplace(newPath, execPath, version string) error {
	batPath := filepath.Join(filepath.Dir(execPath), "update.bat")

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

	cmd := exec.Command("cmd.exe", "/C", "start", "/b", "", batPath)
	if err := cmd.Start(); err != nil {
		_ = os.Remove(batPath)
		return fmt.Errorf("update.bat 실행 실패: %w", err)
	}

	log.Printf("[updater] 프로세스 종료 후 자동 업데이트 예정 — 재시작합니다")

	u.setState(UpdateStatus{
		State:   StateRestartRequired,
		Version: version,
		Percent: 100,
	})

	os.Exit(0)
	return nil
}
