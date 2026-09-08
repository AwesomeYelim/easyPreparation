package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"easyPreparation_1.0/internal/httpx"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/version"
)

// checkResult — 개별 체크 항목 결과
type checkResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// healthResponse — /api/health 응답
type healthResponse struct {
	Status  string                 `json:"status"`
	Checks  map[string]checkResult `json:"checks"`
	Version string                 `json:"version"`
}

// HealthCheck — GET /api/health
// config 디렉토리, data 쓰기 가능, Ghostscript 실행 가능 여부를 체크한다.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	checks := make(map[string]checkResult)

	// 1. config 디렉토리 존재 + main_worship.json 읽기 가능
	checks["config"] = checkConfig()

	// 2. data 디렉토리 쓰기 가능
	checks["data_writable"] = checkDataWritable()

	// 3. Ghostscript 실행 가능
	checks["ghostscript"] = checkGhostscript()

	// 전체 상태 결정
	status := "healthy"
	if checks["config"].Status == "fail" {
		status = "unhealthy"
	} else {
		for _, c := range checks {
			if c.Status == "fail" {
				status = "degraded"
				break
			}
		}
	}

	resp := healthResponse{
		Status:  status,
		Checks:  checks,
		Version: version.Get().Version,
	}

	w.Header().Set("Content-Type", "application/json")
	if status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// checkConfig — config/ 디렉토리 존재 + main_worship.json 읽기 가능 여부
func checkConfig() checkResult {
	execPath := path.ExecutePath("easyPreparation")
	configDir := filepath.Join(execPath, "config")

	info, err := os.Stat(configDir)
	if err != nil {
		return checkResult{Status: "fail", Message: "config 디렉토리 없음"}
	}
	if !info.IsDir() {
		return checkResult{Status: "fail", Message: "config가 디렉토리가 아님"}
	}

	f, err := os.Open(filepath.Join(configDir, "main_worship.json"))
	if err != nil {
		return checkResult{Status: "fail", Message: fmt.Sprintf("main_worship.json 읽기 실패: %v", err)}
	}
	f.Close()

	return checkResult{Status: "pass", Message: ""}
}

// checkDataWritable — data/ 디렉토리 쓰기 가능 여부
func checkDataWritable() checkResult {
	execPath := path.ExecutePath("easyPreparation")
	dataDir := filepath.Join(execPath, "data")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return checkResult{Status: "fail", Message: fmt.Sprintf("data 디렉토리 생성 실패: %v", err)}
	}

	tmp := filepath.Join(dataDir, ".health_check_tmp")
	if err := os.WriteFile(tmp, []byte("ok"), 0644); err != nil {
		return checkResult{Status: "fail", Message: fmt.Sprintf("data 디렉토리 쓰기 불가: %v", err)}
	}
	os.Remove(tmp)

	return checkResult{Status: "pass", Message: ""}
}

// checkGhostscript — Ghostscript 실행 가능 여부 + 버전 반환
func checkGhostscript() checkResult {
	gsPath := resolveGhostscriptPath()

	out, err := exec.Command(gsPath, "--version").Output()
	if err != nil {
		return checkResult{Status: "fail", Message: fmt.Sprintf("gs 실행 불가 (%s): %v", gsPath, err)}
	}

	ver := strings.TrimSpace(string(out))
	return checkResult{Status: "pass", Message: fmt.Sprintf("gs %s", ver)}
}

// resolveGhostscriptPath — 플랫폼별 Ghostscript 경로 탐색
// (pdfrender 패키지의 resolveGhostscript와 동일 로직)
func resolveGhostscriptPath() string {
	switch runtime.GOOS {
	case "windows":
		for _, name := range []string{"gswin64c", "gswin32c", "gs"} {
			if _, err := exec.LookPath(name); err == nil {
				return name
			}
		}
		return "gswin64c"
	default:
		for _, p := range []string{
			"/opt/homebrew/bin/gs",
			"/usr/local/bin/gs",
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		if path, err := exec.LookPath("gs"); err == nil {
			return path
		}
		return "gs"
	}
}
