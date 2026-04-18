package obs

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type wsConfig struct {
	AlertsEnabled  bool   `json:"alerts_enabled"`
	AuthRequired   bool   `json:"auth_required"`
	FirstLoad      bool   `json:"first_load"`
	ServerEnabled  bool   `json:"server_enabled"`
	ServerPassword string `json:"server_password"`
	ServerPort     int    `json:"server_port"`
}

// obsWebSocketConfigPath — OS별 OBS WebSocket 설정 파일 경로
func obsWebSocketConfigPath() (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
	case "darwin":
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, "Library", "Application Support")
	default:
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	if base == "" {
		return "", fmt.Errorf("설정 디렉토리를 찾을 수 없습니다")
	}
	return filepath.Join(base, "obs-studio", "plugin_config", "obs-websocket", "config.json"), nil
}

// AutoConfigureWebSocket — OBS WebSocket 서버 활성화 + 비밀번호 동기화
// 변경 내용을 적용하려면 OBS를 재시작해야 합니다.
func AutoConfigureWebSocket(password string) error {
	cfgPath, err := obsWebSocketConfigPath()
	if err != nil {
		return err
	}

	var cfg wsConfig
	data, err := os.ReadFile(cfgPath)
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	if cfg.ServerPort == 0 {
		cfg.ServerPort = 4455
	}
	cfg.ServerEnabled = true
	cfg.AuthRequired = password != ""
	if password != "" {
		cfg.ServerPassword = password
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("설정 직렬화 실패: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	if err := os.WriteFile(cfgPath, out, 0644); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}
	return nil
}

// RestartOBS — OBS 프로세스를 종료하고 재시작
func RestartOBS() error {
	switch runtime.GOOS {
	case "windows":
		return restartOBSWindows()
	case "darwin":
		return restartOBSMac()
	default:
		return restartOBSLinux()
	}
}

func restartOBSWindows() error {
	// OBS 종료
	_ = exec.Command("taskkill", "/F", "/IM", "obs64.exe").Run()
	_ = exec.Command("taskkill", "/F", "/IM", "obs32.exe").Run()
	time.Sleep(1500 * time.Millisecond)

	// 일반적인 OBS 설치 경로 탐색
	candidates := []string{
		`C:\Program Files\obs-studio\bin\64bit\obs64.exe`,
		`C:\Program Files (x86)\obs-studio\bin\64bit\obs64.exe`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Programs\obs-studio\bin\64bit\obs64.exe`),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			cmd := exec.Command(p)
			cmd.Dir = filepath.Dir(p)
			return cmd.Start()
		}
	}
	return fmt.Errorf("OBS 실행 파일을 찾을 수 없습니다 (수동으로 재시작해주세요)")
}

func restartOBSMac() error {
	_ = exec.Command("osascript", "-e", `quit app "OBS"`).Run()
	time.Sleep(1500 * time.Millisecond)
	return exec.Command("open", "-a", "OBS").Start()
}

func restartOBSLinux() error {
	_ = exec.Command("pkill", "-f", "obs").Run()
	time.Sleep(1500 * time.Millisecond)
	return exec.Command("obs").Start()
}
