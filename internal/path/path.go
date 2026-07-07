package path

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// resolved — 프로세스 생명주기 동안 데이터 루트를 한 번만 해석해 고정한다.
// 실행 중 CWD가 바뀌어도 저장/로드 경로가 흔들리지 않도록 캐시한다.
var (
	resolveOnce sync.Once
	resolvedDir string
)

// ExecutePath — 데이터 루트 디렉터리를 반환한다. 모든 OS에서 동일한 규칙을 적용한다.
//
// 우선순위:
//  1. 환경변수 EASYPREP_DATA_DIR (절대 경로 오버라이드)
//  2. 개발 모드 → 저장소 루트
//     - EASYPREP_DEV=true 이고 CWD 상위에 baseDir 이름의 디렉터리가 있을 때
//     - 또는 CWD 상위의 baseDir 디렉터리에 go.mod가 존재할 때 (go run 직접 실행)
//  3. 포터블/서버 설치 → 실행 파일 상위의 baseDir 디렉터리
//     - baseDir 이름 + config/ 존재 + 쓰기 가능할 때 (예: /home/ubuntu/easyPreparation/bin/server)
//     - Windows Program Files처럼 쓰기 불가면 건너뛰고 사용자 디렉터리로
//  4. 프로덕션 → OS 표준 사용자 데이터 디렉터리 (os.UserConfigDir)
//     - Windows: %APPDATA%\easyPreparation
//     - macOS:   ~/Library/Application Support/easyPreparation
//       (.app 번들 옆에 easyPreparation 디렉터리가 있으면 포터블 설치로 간주하고 그쪽 우선)
//     - Linux:   ~/.config/easyPreparation
//  5. fallback: 실행 파일 디렉터리
func ExecutePath(baseDir string) string {
	resolveOnce.Do(func() {
		resolvedDir = resolve(baseDir)
		log.Printf("[path] 데이터 루트: %s", resolvedDir)
	})
	return resolvedDir
}

func resolve(baseDir string) string {
	// 1. 환경변수 오버라이드
	if override := os.Getenv("EASYPREP_DATA_DIR"); override != "" {
		if abs, err := filepath.Abs(override); err == nil {
			return abs
		}
	}

	// 2. 개발 모드: CWD에서 위로 올라가며 저장소 루트 탐색
	if root, ok := devRepoRoot(baseDir); ok {
		return root
	}

	execPath, execErr := os.Executable()
	if execErr == nil {
		if abs, err := filepath.Abs(execPath); err == nil {
			execPath = abs
		}
	}

	// 3. 포터블/서버 설치: 실행 파일 상위의 baseDir 디렉터리
	// (deploy.yml이 easyPreparation/bin + easyPreparation/config 레이아웃으로 배포)
	if execErr == nil {
		if root, ok := portableRoot(execPath, baseDir); ok {
			return root
		}
	}

	// 4. macOS .app 번들 옆 포터블 디렉터리 (기존 설치 호환)
	if execErr == nil && runtime.GOOS == "darwin" {
		if idx := strings.Index(execPath, ".app/Contents/MacOS"); idx != -1 {
			bundleRoot := execPath[:idx+4] // ".app" 포함
			sideDir := filepath.Join(filepath.Dir(bundleRoot), baseDir)
			if stat, err := os.Stat(sideDir); err == nil && stat.IsDir() {
				return sideDir
			}
		}
	}

	// 5. OS 표준 사용자 데이터 디렉터리
	// (Windows: %APPDATA%, macOS: ~/Library/Application Support, Linux: ~/.config)
	if configDir, err := os.UserConfigDir(); err == nil {
		dataDir := filepath.Join(configDir, baseDir)
		if err := os.MkdirAll(dataDir, 0755); err == nil {
			return dataDir
		}
	}

	// 6. fallback: 실행 파일 디렉터리
	if execErr != nil {
		log.Printf("[path] 실행 파일 경로를 가져오는 중 오류 발생: %v", execErr)
		return ""
	}
	return filepath.Dir(execPath)
}

// devRepoRoot — CWD에서 상위로 올라가며 baseDir 이름의 디렉터리를 찾는다.
// 디렉터리 이름을 세그먼트 단위로 비교하므로 "easyPreparation-backup" 같은
// 유사 이름 경로에는 매칭되지 않는다.
//
// EASYPREP_DEV=true면 이름 일치만으로 인정하고,
// 아니면 go.mod가 있는 실제 저장소일 때만 개발 모드로 판정한다.
// portableRoot — 실행 파일 위치에서 상위로 올라가며 baseDir 이름 + config/ 를 가진
// 쓰기 가능한 디렉터리를 찾는다. Linux 서버 배포(easyPreparation/bin/바이너리)와
// 포터블 설치를 커버하며, Windows Program Files처럼 쓰기 불가한 위치는 제외한다.
func portableRoot(execPath, baseDir string) (string, bool) {
	dir := filepath.Dir(execPath)
	for i := 0; i < 3; i++ {
		if filepath.Base(dir) == baseDir {
			if stat, err := os.Stat(filepath.Join(dir, "config")); err == nil && stat.IsDir() && isWritable(dir) {
				return dir, true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// isWritable — 디렉터리에 파일을 실제로 생성해보고 쓰기 가능 여부를 확인한다.
func isWritable(dir string) bool {
	f, err := os.CreateTemp(dir, ".ep_write_test_*")
	if err != nil {
		return false
	}
	name := f.Name()
	f.Close()
	os.Remove(name)
	return true
}

func devRepoRoot(baseDir string) (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", false
	}
	explicitDev := os.Getenv("EASYPREP_DEV") == "true"
	for cur := abs; ; {
		if filepath.Base(cur) == baseDir {
			if explicitDev {
				return cur, true
			}
			if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
				return cur, true
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", false
		}
		cur = parent
	}
}
