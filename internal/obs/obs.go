package obs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/config"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/transitions"
	"github.com/andreykaipov/goobs/api/typedefs"
)

// Config — config/obs.json 구조
type Config struct {
	Host         string            `json:"host"`         // "localhost:4455"
	Password     string            `json:"password"`
	Scenes       map[string]string `json:"scenes"`       // "찬송" → "camera"
	CameraScene  string            `json:"cameraScene"`  // fade-back 복귀 씬 (기본: "camera")
	DisplayScene string            `json:"displayScene"` // fade-back 시 표시할 씬 (기본: "monitor")
	FadeMs       int               `json:"fadeMs"`       // fade 트랜지션 길이 ms (기본: 800)
	FadeDelaySec int               `json:"fadeDelaySec"` // display 표시 후 camera 복귀까지 초 (기본: 3)
}

// Status — OBS 연결 상태
type Status struct {
	Connected    bool   `json:"connected"`
	CurrentScene string `json:"currentScene"`
}

// StreamStatus — OBS 스트리밍 상태
type StreamStatus struct {
	Active       bool    `json:"active"`
	Reconnecting bool    `json:"reconnecting"`
	Timecode     string  `json:"timecode"`
	BytesSent    float64 `json:"bytesSent"`
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
	fadeCancel   chan struct{} // 현재 fadeBack 타이머 취소용
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
		if cfg.CameraScene == "" {
			cfg.CameraScene = "camera"
		}
		if cfg.DisplayScene == "" {
			cfg.DisplayScene = "monitor"
		}
		if cfg.FadeMs <= 0 {
			cfg.FadeMs = 800
		}
		if cfg.FadeDelaySec <= 0 {
			cfg.FadeDelaySec = 3
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

	// 이미 같은 씬이면 스킵
	m.mu.RLock()
	current := m.currentScene
	m.mu.RUnlock()
	if current == sceneName {
		return
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

// CancelFadeBack — 진행 중인 fadeBack 타이머 취소
func (m *Manager) CancelFadeBack() {
	if m == nil {
		return
	}
	m.mu.Lock()
	if m.fadeCancel != nil {
		select {
		case <-m.fadeCancel:
		default:
			close(m.fadeCancel)
		}
		m.fadeCancel = nil
	}
	m.mu.Unlock()
}

// SwitchSceneWithFadeBack — displayScene으로 전환 후 N초 뒤 fade로 camera 복귀
// 찬양/찬송 항목용: display(가사/악보) 보여줬다가 fade out → camera
func (m *Manager) SwitchSceneWithFadeBack(title string) {
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

	// 기존 fadeBack 타이머 취소
	m.CancelFadeBack()

	// displayScene (monitor)으로 즉시 전환
	displayScene := m.config.DisplayScene
	m.mu.RLock()
	current := m.currentScene
	m.mu.RUnlock()

	if current != displayScene {
		params := scenes.NewSetCurrentProgramSceneParams().WithSceneName(displayScene)
		_, err := client.Scenes.SetCurrentProgramScene(params)
		if err != nil {
			log.Printf("[obs] 씬 전환 실패 (%s → %s): %v", title, displayScene, err)
			return
		}
		m.mu.Lock()
		m.currentScene = displayScene
		m.mu.Unlock()
		log.Printf("[obs] 씬 전환: %s → %s (fade-back 예약)", title, displayScene)
	}

	// N초 후 fade로 camera 복귀
	cancel := make(chan struct{})
	m.mu.Lock()
	m.fadeCancel = cancel
	m.mu.Unlock()

	go func() {
		delay := time.Duration(m.config.FadeDelaySec) * time.Second
		select {
		case <-cancel:
			return
		case <-time.After(delay):
		}

		m.mu.RLock()
		c := m.client
		conn := m.connected
		m.mu.RUnlock()
		if !conn || c == nil {
			return
		}

		// Fade 트랜지션 설정
		tParams := transitions.NewSetCurrentSceneTransitionParams().WithTransitionName("Fade")
		_, _ = c.Transitions.SetCurrentSceneTransition(tParams)
		dParams := transitions.NewSetCurrentSceneTransitionDurationParams().WithTransitionDuration(float64(m.config.FadeMs))
		_, _ = c.Transitions.SetCurrentSceneTransitionDuration(dParams)

		// camera 씬으로 전환
		cameraScene := m.config.CameraScene
		sParams := scenes.NewSetCurrentProgramSceneParams().WithSceneName(cameraScene)
		_, err := c.Scenes.SetCurrentProgramScene(sParams)
		if err != nil {
			log.Printf("[obs] fade-back 실패 → %s: %v", cameraScene, err)
			return
		}
		m.mu.Lock()
		m.currentScene = cameraScene
		m.mu.Unlock()
		log.Printf("[obs] fade-back: %s → %s (%dms)", displayScene, cameraScene, m.config.FadeMs)
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

// StartStreaming — OBS 스트리밍 시작 (미연결 시 no-op, 이미 스트리밍 중이면 스킵)
func (m *Manager) StartStreaming() error {
	if m == nil || !m.enabled {
		return nil
	}
	m.mu.RLock()
	client := m.client
	connected := m.connected
	m.mu.RUnlock()
	if !connected || client == nil {
		return nil
	}

	// 이미 스트리밍 중이면 스킵
	status, err := client.Stream.GetStreamStatus()
	if err == nil && status.OutputActive {
		log.Println("[obs] 이미 스트리밍 중 — 스킵")
		return nil
	}

	_, err = client.Stream.StartStream()
	if err != nil {
		log.Printf("[obs] 스트리밍 시작 실패: %v", err)
		return err
	}
	log.Println("[obs] 스트리밍 시작")
	return nil
}

// StopStreaming — OBS 스트리밍 중지 (미연결 시 no-op)
func (m *Manager) StopStreaming() error {
	if m == nil || !m.enabled {
		return nil
	}
	m.mu.RLock()
	client := m.client
	connected := m.connected
	m.mu.RUnlock()
	if !connected || client == nil {
		return nil
	}
	_, err := client.Stream.StopStream()
	if err != nil {
		log.Printf("[obs] 스트리밍 중지 실패: %v", err)
		return err
	}
	log.Println("[obs] 스트리밍 중지")
	return nil
}

// GetStreamStatus — OBS 스트리밍 상태 조회
func (m *Manager) GetStreamStatus() StreamStatus {
	if m == nil || !m.enabled {
		return StreamStatus{}
	}
	m.mu.RLock()
	client := m.client
	connected := m.connected
	m.mu.RUnlock()
	if !connected || client == nil {
		return StreamStatus{}
	}
	resp, err := client.Stream.GetStreamStatus()
	if err != nil {
		log.Printf("[obs] 스트리밍 상태 조회 실패: %v", err)
		return StreamStatus{}
	}
	return StreamStatus{
		Active:       resp.OutputActive,
		Reconnecting: resp.OutputReconnecting,
		Timecode:     resp.OutputTimecode,
		BytesSent:    resp.OutputBytes,
	}
}

// SetStreamSettings — OBS 스트림 서비스를 커스텀 RTMP로 설정
func (m *Manager) SetStreamSettings(server, key string) error {
	if m == nil || !m.enabled {
		return fmt.Errorf("OBS 미연결")
	}
	m.mu.RLock()
	client := m.client
	connected := m.connected
	m.mu.RUnlock()
	if !connected || client == nil {
		return fmt.Errorf("OBS 미연결")
	}

	params := config.NewSetStreamServiceSettingsParams().
		WithStreamServiceType("rtmp_custom").
		WithStreamServiceSettings(&typedefs.StreamServiceSettings{
			Server: server,
			Key:    key,
		})

	_, err := client.Config.SetStreamServiceSettings(params)
	if err != nil {
		return fmt.Errorf("OBS 스트림 설정 실패: %w", err)
	}

	log.Printf("[obs] 스트림 설정 완료: server=%s", server)
	return nil
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

// SceneItemInfo — 씬 내 소스 정보
type SceneItemInfo struct {
	SceneItemID int     `json:"sceneItemId"`
	SourceName  string  `json:"sourceName"`
	InputKind   string  `json:"inputKind"`
	Enabled     bool    `json:"enabled"`
	PositionX   float64 `json:"positionX"`
	PositionY   float64 `json:"positionY"`
	ScaleX      float64 `json:"scaleX"`
	ScaleY      float64 `json:"scaleY"`
}

// DeviceInfo — 카메라 디바이스 정보
type DeviceInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// getClient — 연결 상태 확인 + client 반환 헬퍼
func (m *Manager) getClient() (*goobs.Client, error) {
	if m == nil || !m.enabled {
		return nil, fmt.Errorf("OBS 미연결")
	}
	m.mu.RLock()
	client := m.client
	connected := m.connected
	m.mu.RUnlock()
	if !connected || client == nil {
		return nil, fmt.Errorf("OBS 미연결")
	}
	return client, nil
}

// GetScenes — 씬 목록 조회
func (m *Manager) GetScenes() ([]string, error) {
	client, err := m.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.Scenes.GetSceneList(scenes.NewGetSceneListParams())
	if err != nil {
		return nil, fmt.Errorf("씬 목록 조회 실패: %w", err)
	}
	var names []string
	for _, s := range resp.Scenes {
		names = append(names, s.SceneName)
	}
	return names, nil
}

// GetSceneItems — 씬 내 소스 목록 조회
func (m *Manager) GetSceneItems(sceneName string) ([]SceneItemInfo, error) {
	client, err := m.getClient()
	if err != nil {
		return nil, err
	}
	params := sceneitems.NewGetSceneItemListParams().WithSceneName(sceneName)
	resp, err := client.SceneItems.GetSceneItemList(params)
	if err != nil {
		return nil, fmt.Errorf("소스 목록 조회 실패: %w", err)
	}
	var items []SceneItemInfo
	for _, si := range resp.SceneItems {
		items = append(items, SceneItemInfo{
			SceneItemID: si.SceneItemID,
			SourceName:  si.SourceName,
			InputKind:   si.InputKind,
			Enabled:     si.SceneItemEnabled,
			PositionX:   si.SceneItemTransform.PositionX,
			PositionY:   si.SceneItemTransform.PositionY,
			ScaleX:      si.SceneItemTransform.ScaleX,
			ScaleY:      si.SceneItemTransform.ScaleY,
		})
	}
	return items, nil
}

// CreateImageSource — image_source 생성, sceneItemId 반환
func (m *Manager) CreateImageSource(sceneName, name, filePath string, enabled bool) (int, error) {
	client, err := m.getClient()
	if err != nil {
		return 0, err
	}
	params := inputs.NewCreateInputParams().
		WithInputKind("image_source").
		WithInputName(name).
		WithSceneName(sceneName).
		WithSceneItemEnabled(enabled).
		WithInputSettings(map[string]any{"file": filePath})
	resp, err := client.Inputs.CreateInput(params)
	if err != nil {
		return 0, fmt.Errorf("이미지 소스 생성 실패: %w", err)
	}
	log.Printf("[obs] 이미지 소스 생성: %s (sceneItemId=%d)", name, resp.SceneItemId)
	return resp.SceneItemId, nil
}

// CreateCameraSource — 카메라 소스 생성 (macOS: av_capture_input_v2, Windows: dshow_input)
func (m *Manager) CreateCameraSource(sceneName, name, deviceID string) (int, error) {
	client, err := m.getClient()
	if err != nil {
		return 0, err
	}
	inputKind := "av_capture_input_v2"
	settingsKey := "device"
	if runtime.GOOS == "windows" {
		inputKind = "dshow_input"
		settingsKey = "video_device_id"
	}
	params := inputs.NewCreateInputParams().
		WithInputKind(inputKind).
		WithInputName(name).
		WithSceneName(sceneName).
		WithSceneItemEnabled(true).
		WithInputSettings(map[string]any{settingsKey: deviceID})
	resp, err := client.Inputs.CreateInput(params)
	if err != nil {
		return 0, fmt.Errorf("카메라 소스 생성 실패: %w", err)
	}
	log.Printf("[obs] 카메라 소스 생성: %s (device=%s, sceneItemId=%d)", name, deviceID, resp.SceneItemId)
	return resp.SceneItemId, nil
}

// SetItemTransform — 소스 위치/크기 변경
func (m *Manager) SetItemTransform(sceneName string, itemID int, x, y, scaleX, scaleY float64) error {
	client, err := m.getClient()
	if err != nil {
		return err
	}
	transform := &typedefs.SceneItemTransform{
		PositionX: x,
		PositionY: y,
		ScaleX:    scaleX,
		ScaleY:    scaleY,
	}
	params := sceneitems.NewSetSceneItemTransformParams().
		WithSceneItemId(itemID).
		WithSceneName(sceneName).
		WithSceneItemTransform(transform)
	_, err = client.SceneItems.SetSceneItemTransform(params)
	if err != nil {
		return fmt.Errorf("소스 트랜스폼 설정 실패: %w", err)
	}
	return nil
}

// SetItemEnabled — 소스 표시/숨김
func (m *Manager) SetItemEnabled(sceneName string, itemID int, enabled bool) error {
	client, err := m.getClient()
	if err != nil {
		return err
	}
	params := sceneitems.NewSetSceneItemEnabledParams().
		WithSceneItemId(itemID).
		WithSceneName(sceneName).
		WithSceneItemEnabled(enabled)
	_, err = client.SceneItems.SetSceneItemEnabled(params)
	if err != nil {
		return fmt.Errorf("소스 표시 설정 실패: %w", err)
	}
	return nil
}

// RemoveInput — 소스 제거 (모든 씬에서 자동 제거)
func (m *Manager) RemoveInput(name string) error {
	client, err := m.getClient()
	if err != nil {
		return err
	}
	params := inputs.NewRemoveInputParams().WithInputName(name)
	_, err = client.Inputs.RemoveInput(params)
	if err != nil {
		return fmt.Errorf("소스 제거 실패: %w", err)
	}
	log.Printf("[obs] 소스 제거: %s", name)
	return nil
}

// ListCameraDevices — 임시 입력 생성 → 디바이스 목록 → 정리
func (m *Manager) ListCameraDevices() ([]DeviceInfo, error) {
	client, err := m.getClient()
	if err != nil {
		return nil, err
	}

	inputKind := "av_capture_input_v2"
	propertyName := "device"
	if runtime.GOOS == "windows" {
		inputKind = "dshow_input"
		propertyName = "video_device_id"
	}

	tmpName := "_ep_tmp_cam_probe"
	// 임시 입력 생성
	createParams := inputs.NewCreateInputParams().
		WithInputKind(inputKind).
		WithInputName(tmpName).
		WithSceneName(m.currentScene).
		WithSceneItemEnabled(false)
	_, err = client.Inputs.CreateInput(createParams)
	if err != nil {
		return nil, fmt.Errorf("카메라 탐지용 임시 입력 생성 실패: %w", err)
	}

	// 반드시 정리
	defer func() {
		rmParams := inputs.NewRemoveInputParams().WithInputName(tmpName)
		client.Inputs.RemoveInput(rmParams)
	}()

	// 디바이스 목록 조회
	propParams := inputs.NewGetInputPropertiesListPropertyItemsParams().
		WithInputName(tmpName).
		WithPropertyName(propertyName)
	resp, err := client.Inputs.GetInputPropertiesListPropertyItems(propParams)
	if err != nil {
		return nil, fmt.Errorf("카메라 디바이스 목록 조회 실패: %w", err)
	}

	var devices []DeviceInfo
	for _, item := range resp.PropertyItems {
		val, _ := item.ItemValue.(string)
		if val == "" {
			continue
		}
		devices = append(devices, DeviceInfo{
			Name:  item.ItemName,
			Value: val,
		})
	}
	return devices, nil
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
