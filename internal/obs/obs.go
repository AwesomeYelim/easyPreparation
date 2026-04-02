package obs

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
)

// Config — config/obs.json 구조
type Config struct {
	Host     string            `json:"host"`     // "localhost:4455"
	Password string            `json:"password"`
	Scenes   map[string]string `json:"scenes"`   // "찬송" → "camera"
}

// Status — OBS 연결 상태
type Status struct {
	Connected    bool   `json:"connected"`
	CurrentScene string `json:"currentScene"`
}

// Manager — OBS WebSocket 매니저 (싱글턴)
type Manager struct {
	mu           sync.RWMutex
	client       *goobs.Client
	config       Config
	enabled      bool
	connected    bool
	currentScene string
	stopCh       chan struct{}
}

var (
	instance *Manager
	once     sync.Once
)

// Init — config 파일 로드 + 연결 시작
func Init(configPath string) {
	once.Do(func() {
		instance = &Manager{stopCh: make(chan struct{})}

		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Printf("[obs] config 파일 없음 (%s) — OBS 연동 비활성", configPath)
			return
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Printf("[obs] config 파싱 실패: %v — OBS 연동 비활성", err)
			return
		}
		if cfg.Host == "" {
			cfg.Host = "localhost:4455"
		}

		instance.config = cfg
		instance.enabled = true
		log.Printf("[obs] OBS 연동 활성 (host=%s)", cfg.Host)

		go instance.connectLoop()
	})
}

// Get — 싱글턴 접근
func Get() *Manager {
	return instance
}

// SwitchScene — title → OBS 씬 매핑 → 씬 전환 (비차단, 에러 로그만)
func (m *Manager) SwitchScene(title string) {
	if m == nil || !m.enabled {
		return
	}

	m.mu.RLock()
	client := m.client
	connected := m.connected
	m.mu.RUnlock()

	if !connected || client == nil {
		return
	}

	sceneName, ok := m.resolveScene(title)
	if !ok {
		return // 매핑 없는 항목은 씬 전환 안 함
	}

	go func() {
		params := scenes.NewSetCurrentProgramSceneParams().WithSceneName(sceneName)
		_, err := client.Scenes.SetCurrentProgramScene(params)
		if err != nil {
			log.Printf("[obs] 씬 전환 실패 (%s → %s): %v", title, sceneName, err)
			return
		}
		m.mu.Lock()
		m.currentScene = sceneName
		m.mu.Unlock()
		log.Printf("[obs] 씬 전환: %s → %s", title, sceneName)
	}()
}

// GetStatus — 현재 OBS 연결 상태
func (m *Manager) GetStatus() Status {
	if m == nil || !m.enabled {
		return Status{}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Status{
		Connected:    m.connected,
		CurrentScene: m.currentScene,
	}
}

// Disconnect — 종료 시 정리
func (m *Manager) Disconnect() {
	if m == nil || !m.enabled {
		return
	}
	close(m.stopCh)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil {
		_ = m.client.Disconnect()
		m.client = nil
	}
	m.connected = false
	log.Println("[obs] 연결 해제")
}

// resolveScene — title → OBS 씬 이름 매핑 (매핑 없으면 false)
func (m *Manager) resolveScene(title string) (string, bool) {
	if scene, ok := m.config.Scenes[title]; ok {
		return scene, true
	}
	return "", false
}

// connectLoop — 연결 유지 루프 (5초마다 재연결 시도)
func (m *Manager) connectLoop() {
	for {
		select {
		case <-m.stopCh:
			return
		default:
		}

		m.tryConnect()

		// 연결 상태면 끊길 때까지 대기
		m.mu.RLock()
		connected := m.connected
		m.mu.RUnlock()

		if connected {
			// 연결 유지 확인 루프
			for {
				select {
				case <-m.stopCh:
					return
				case <-time.After(5 * time.Second):
				}

				m.mu.RLock()
				c := m.client
				m.mu.RUnlock()
				if c == nil {
					break
				}

				// 연결 상태 확인 (간단한 요청)
				_, err := c.Scenes.GetCurrentProgramScene()
				if err != nil {
					log.Printf("[obs] 연결 끊김: %v — 재연결 시도", err)
					m.mu.Lock()
					m.connected = false
					m.client = nil
					m.mu.Unlock()
					break
				}
			}
		}

		select {
		case <-m.stopCh:
			return
		case <-time.After(5 * time.Second):
		}
	}
}

// tryConnect — OBS WebSocket 연결 시도
func (m *Manager) tryConnect() {
	var opts []goobs.Option
	if m.config.Password != "" {
		opts = append(opts, goobs.WithPassword(m.config.Password))
	}

	client, err := goobs.New(m.config.Host, opts...)
	if err != nil {
		log.Printf("[obs] 연결 실패 (%s): %v", m.config.Host, err)
		return
	}

	// 현재 씬 조회
	resp, err := client.Scenes.GetCurrentProgramScene()
	if err != nil {
		log.Printf("[obs] 씬 조회 실패: %v", err)
		_ = client.Disconnect()
		return
	}

	m.mu.Lock()
	m.client = client
	m.connected = true
	m.currentScene = resp.CurrentProgramSceneName
	m.mu.Unlock()

	log.Printf("[obs] OBS 연결 성공 (현재 씬: %s)", resp.CurrentProgramSceneName)
}
