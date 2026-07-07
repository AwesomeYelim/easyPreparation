package path

import (
	"os"
	"path/filepath"
	"testing"
)

const testBase = "easyPreparation"

// resolveSymlinks — macOS에서 t.TempDir()가 /var(→/private/var) 심볼릭 링크를 반환하므로
// os.Getwd 결과와 비교하려면 실경로로 정규화 필요
func resolveSymlinks(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

// 저장소 루트(go.mod 존재)에서는 환경변수 없이도 개발 모드로 판정
func TestDevRepoRoot_GoMod(t *testing.T) {
	root := filepath.Join(t.TempDir(), testBase)
	sub := filepath.Join(root, "ui", "app")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x"), 0644)

	chdir(t, sub)
	want := resolveSymlinks(t, root)
	got, ok := devRepoRoot(testBase)
	if !ok || got != want {
		t.Errorf("devRepoRoot = %q, %v; want %q, true", got, ok, want)
	}
}

// go.mod 없는 easyPreparation 디렉터리는 EASYPREP_DEV=true일 때만 인정
func TestDevRepoRoot_ExplicitEnvOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), testBase)
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	chdir(t, root)

	if _, ok := devRepoRoot(testBase); ok {
		t.Error("go.mod 없고 EASYPREP_DEV 미설정인데 개발 모드로 판정됨")
	}

	t.Setenv("EASYPREP_DEV", "true")
	want := resolveSymlinks(t, root)
	got, ok := devRepoRoot(testBase)
	if !ok || got != want {
		t.Errorf("EASYPREP_DEV=true: devRepoRoot = %q, %v; want %q, true", got, ok, want)
	}
}

// "easyPreparation-backup" 같은 유사 이름은 세그먼트 매칭에서 제외 (구버전 substring 버그)
func TestDevRepoRoot_NoSubstringMatch(t *testing.T) {
	root := filepath.Join(t.TempDir(), testBase+"-backup")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x"), 0644)

	chdir(t, root)
	t.Setenv("EASYPREP_DEV", "true")
	if got, ok := devRepoRoot(testBase); ok {
		t.Errorf("유사 이름 디렉터리에 매칭됨: %q", got)
	}
}

// 서버 배포 레이아웃: easyPreparation/bin/서버바이너리 + easyPreparation/config
func TestPortableRoot_ServerLayout(t *testing.T) {
	root := filepath.Join(t.TempDir(), testBase)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "config"), 0755); err != nil {
		t.Fatal(err)
	}

	got, ok := portableRoot(filepath.Join(binDir, "server"), testBase)
	if !ok || got != root {
		t.Errorf("portableRoot = %q, %v; want %q, true", got, ok, root)
	}
}

// config/ 디렉터리가 없으면 포터블 설치로 인정하지 않음
func TestPortableRoot_RequiresConfig(t *testing.T) {
	root := filepath.Join(t.TempDir(), testBase)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}

	if got, ok := portableRoot(filepath.Join(binDir, "server"), testBase); ok {
		t.Errorf("config/ 없는데 포터블로 판정됨: %q", got)
	}
}

// 무관한 경로의 실행 파일은 포터블 매칭 안 됨
func TestPortableRoot_UnrelatedPath(t *testing.T) {
	dir := t.TempDir()
	if got, ok := portableRoot(filepath.Join(dir, "app.exe"), testBase); ok {
		t.Errorf("무관한 경로에 매칭됨: %q", got)
	}
}
