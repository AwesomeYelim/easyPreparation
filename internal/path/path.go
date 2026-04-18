package path

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ExecutePath — 실행 파일/작업 디렉터리 기준으로 프로젝트 루트를 반환합니다.
//
// 우선순위:
//  1. 환경변수 EASYPREP_DATA_DIR (절대 경로 오버라이드)
//  2. 현재 작업 디렉터리에서 baseDir 세그먼트 탐색
//  3. Windows: %APPDATA%\easyPreparation (C:\Program Files\ 쓰기 권한 문제 방지)
//  4. macOS .app 번들 감지 (Contents/MacOS 포함 시)
//     → 번들 옆 디렉터리(easyPreparation) 또는 ~/Library/Application Support/easyPreparation
//  5. 실행 파일 경로 반환 (fallback)
func ExecutePath(baseDir string) string {
	// 1. 환경변수 오버라이드
	if override := os.Getenv("EASYPREP_DATA_DIR"); override != "" {
		abs, err := filepath.Abs(override)
		if err == nil {
			return abs
		}
	}

	// 2. Windows: %APPDATA%\easyPreparation 최우선
	// C:\Program Files\ 는 쓰기 권한이 없고, cwd 탐색에서
	// "easyPreparation" 세그먼트가 걸려 Program Files 경로를 반환하는 버그 방지
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			appDataDir := filepath.Join(appData, baseDir)
			if err := os.MkdirAll(appDataDir, 0755); err == nil {
				return appDataDir
			}
		}
	}

	// 3. 작업 디렉터리에서 baseDir 탐색 (개발 모드 — Mac/Linux)
	fullPath, err := os.Getwd()
	if err != nil {
		log.Printf("작업 디렉터리를 가져오는 중 오류 발생: %v", err)
	} else {
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			log.Printf("절대 경로 변환 오류: %v", err)
		} else if index := strings.Index(absPath, baseDir); index != -1 {
			return absPath[:index+len(baseDir)]
		}
	}

	// 4. 실행 파일 경로 확인
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("실행 파일 경로를 가져오는 중 오류 발생: %v", err)
		return ""
	}

	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		log.Printf("실행 파일 절대 경로 변환 오류: %v", err)
		return ""
	}

	// 5. macOS .app 번들 감지: .../Foo.app/Contents/MacOS/binary
	if runtime.GOOS == "darwin" && strings.Contains(absExecPath, ".app/Contents/MacOS") {
		// 번들 루트: .app 디렉터리
		appIdx := strings.Index(absExecPath, ".app/Contents/MacOS")
		bundleRoot := absExecPath[:appIdx+4] // ".app" 포함

		// 방법 A: 번들 옆에 easyPreparation 디렉터리
		sideDir := filepath.Join(filepath.Dir(bundleRoot), baseDir)
		if stat, err := os.Stat(sideDir); err == nil && stat.IsDir() {
			return sideDir
		}

		// 방법 B: ~/Library/Application Support/easyPreparation
		if home, err := os.UserHomeDir(); err == nil {
			supportDir := filepath.Join(home, "Library", "Application Support", baseDir)
			// 없으면 생성
			if err := os.MkdirAll(supportDir, 0755); err == nil {
				return supportDir
			}
		}

		// fallback: 번들 디렉터리 반환
		return filepath.Dir(bundleRoot)
	}

	return filepath.Dir(absExecPath)
}
