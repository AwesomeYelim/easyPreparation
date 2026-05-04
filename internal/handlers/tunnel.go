package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// tunnelState — cloudflared 터널 전역 상태
var tunnelState = struct {
	mu        sync.RWMutex
	publicURL string      // "https://xxxx.trycloudflare.com" | ""
	running   bool
	cancel    context.CancelFunc
}{
}

// GetTunnelURL — 현재 터널 public URL 반환 (없으면 "")
func GetTunnelURL() string {
	tunnelState.mu.RLock()
	defer tunnelState.mu.RUnlock()
	return tunnelState.publicURL
}

var tunnelURLRe = regexp.MustCompile(`https://[a-zA-Z0-9\-]+\.trycloudflare\.com`)

// StartTunnel — cloudflared quick tunnel 시작. 이미 실행 중이면 무시.
func StartTunnel(port string) {
	tunnelState.mu.Lock()
	if tunnelState.running {
		tunnelState.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	tunnelState.cancel = cancel
	tunnelState.running = true
	tunnelState.mu.Unlock()

	go func() {
		defer func() {
			tunnelState.mu.Lock()
			tunnelState.running = false
			tunnelState.publicURL = ""
			tunnelState.cancel = nil
			tunnelState.mu.Unlock()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			cmd := exec.CommandContext(ctx,
				"/opt/homebrew/bin/cloudflared",
				"tunnel", "--url", "http://localhost:"+port,
			)

			// cloudflared는 URL을 stderr로 출력
			pipe, err := cmd.StderrPipe()
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			if err := cmd.Start(); err != nil {
				time.Sleep(5 * time.Second)
				continue
			}

			// URL 파싱 goroutine
			scanner := bufio.NewScanner(pipe)
			for scanner.Scan() {
				line := scanner.Text()
				if m := tunnelURLRe.FindString(line); m != "" {
					tunnelState.mu.Lock()
					tunnelState.publicURL = m
					tunnelState.mu.Unlock()
					// WS로 알림
					BroadcastMessage("tunnel_url", map[string]interface{}{"url": m})
					break
				}
			}
			// 프로세스 종료 대기
			cmd.Wait()

			select {
			case <-ctx.Done():
				return
			default:
				// 재시작 전 URL 초기화
				tunnelState.mu.Lock()
				tunnelState.publicURL = ""
				tunnelState.mu.Unlock()
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

// StopTunnel — 터널 중지
func StopTunnel() {
	tunnelState.mu.Lock()
	defer tunnelState.mu.Unlock()
	if tunnelState.cancel != nil {
		tunnelState.cancel()
	}
}

// HandleTunnelStatus — GET /api/tunnel/status
func HandleTunnelStatus(w http.ResponseWriter, r *http.Request) {
	url := GetTunnelURL()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running": tunnelState.running,
		"url":     url,
	})
}

// HandleTunnelStart — POST /api/tunnel/start
func HandleTunnelStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Port string `json:"port"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.Port == "" {
		body.Port = "8080"
	}
	body.Port = strings.TrimSpace(body.Port)
	StartTunnel(body.Port)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// HandleTunnelStop — POST /api/tunnel/stop
func HandleTunnelStop(w http.ResponseWriter, r *http.Request) {
	StopTunnel()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}
