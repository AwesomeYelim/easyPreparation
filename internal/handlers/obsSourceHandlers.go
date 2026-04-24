package handlers

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
)

// logoDir — 로고 저장 디렉토리
func logoDir() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "templates", "logo")
}

// logoFilePath — 단일 로고 파일 경로
func logoFilePath() string {
	return filepath.Join(logoDir(), "logo.png")
}

// OBSConnectHandler — POST /api/obs/connect {ip, port, password}
// Feature gate 없이 접근 가능 — 연결 설정 저장 + 재연결
func OBSConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		IP       string `json:"ip"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
		return
	}
	if body.IP == "" {
		body.IP = "localhost"
	}
	if body.Port == 0 {
		body.Port = 4455
	}
	host := fmt.Sprintf("%s:%d", body.IP, body.Port)

	configPath := filepath.Join(path.ExecutePath("easyPreparation"), "config", "obs.json")
	m := obs.Get()
	if m == nil {
		http.Error(w, "OBS 매니저 초기화 안됨", http.StatusInternalServerError)
		return
	}

	m.Connect(host, body.Password, configPath)

	// 연결 시도 후 결과 반환 (최대 2초 대기)
	for i := 0; i < 4; i++ {
		time.Sleep(500 * time.Millisecond)
		if m.GetStatus().Connected {
			break
		}
	}

	status := m.GetStatus()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":        true,
		"connected": status.Connected,
		"host":      host,
	})
}

// OBSAutoConfigureHandler — POST /api/obs/auto-configure
// OBS WebSocket 서버 설정 파일을 자동으로 활성화하고 OBS를 재시작합니다.
func OBSAutoConfigureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if err := obs.AutoConfigureWebSocket(body.Password); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	restartErr := obs.RestartOBS()
	if restartErr != nil {
		// OBS 실행 파일 못 찾아도 config 저장은 성공 — 수동 재시작 안내
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":            true,
			"manualRestart": true,
			"message":       "설정이 저장됐습니다. OBS를 수동으로 재시작해주세요.",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"message": "OBS WebSocket 서버가 활성화됐습니다. 재시작 중...",
	})
}

// OBSStatusHandler — GET /api/obs/status (feature gate 없음)
func OBSStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	m := obs.Get()
	if m == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"connected": false})
		return
	}
	status := m.GetStatus()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected":    status.Connected,
		"currentScene": status.CurrentScene,
	})
}

// OBSScenesHandler — GET /api/obs/scenes
func OBSScenesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	m := obs.Get()
	status := m.GetStatus()

	scenes, err := m.GetScenes()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"connected": status.Connected,
			"scenes":    []string{},
			"error":     err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected":    status.Connected,
		"scenes":       scenes,
		"currentScene": status.CurrentScene,
	})
}

// OBSSourcesHandler — GET /api/obs/sources?scene=xxx
func OBSSourcesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	scene := r.URL.Query().Get("scene")
	if scene == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "scene 파라미터 필요", "items": []obs.SceneItemInfo{}})
		return
	}

	m := obs.Get()
	items, err := m.GetSceneItems(scene)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "items": []obs.SceneItemInfo{}})
		return
	}
	if items == nil {
		items = []obs.SceneItemInfo{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"items": items})
}

// logoIndexPath — 순환 인덱스 파일 경로
func logoIndexPath() string {
	return filepath.Join(logoDir(), "logo_index.txt")
}

// readLogoIndex — 현재 순환 인덱스(1~3)를 읽는다. 파일 없으면 0 반환.
func readLogoIndex() int {
	data, err := os.ReadFile(logoIndexPath())
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || n < 1 || n > 3 {
		return 0
	}
	return n
}

// writeLogoIndex — 순환 인덱스를 파일에 저장한다.
func writeLogoIndex(idx int) error {
	return os.WriteFile(logoIndexPath(), []byte(strconv.Itoa(idx)), 0644)
}

// OBSLogoUploadHandler — POST /api/obs/logo/upload (multipart: image)
func OBSLogoUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	r.ParseMultipartForm(10 << 20) // 10MB
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "이미지 파일 없음", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		http.Error(w, "PNG/JPG만 허용", http.StatusBadRequest)
		return
	}

	dir := logoDir()
	os.MkdirAll(dir, 0755)

	// 기존 logo.png 저장 (apply 핸들러 호환성 유지)
	savePath := logoFilePath()
	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "파일 저장 실패", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		http.Error(w, "파일 쓰기 실패", http.StatusInternalServerError)
		return
	}
	dst.Close()

	// 순환 인덱스 계산 (1 -> 2 -> 3 -> 1 ...)
	prev := readLogoIndex()
	next := (prev % 3) + 1

	// logo_N.png 에 복사
	slotName := fmt.Sprintf("logo_%d.png", next)
	slotPath := filepath.Join(dir, slotName)

	srcFile, err := os.Open(savePath)
	if err != nil {
		http.Error(w, "파일 복사 실패", http.StatusInternalServerError)
		return
	}
	defer srcFile.Close()

	slotDst, err := os.Create(slotPath)
	if err != nil {
		http.Error(w, "순환 저장 실패", http.StatusInternalServerError)
		return
	}
	defer slotDst.Close()

	if _, err := io.Copy(slotDst, srcFile); err != nil {
		http.Error(w, "순환 파일 쓰기 실패", http.StatusInternalServerError)
		return
	}

	// 인덱스 갱신
	_ = writeLogoIndex(next)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":   true,
		"path": savePath,
		"slot": slotName,
	})
}

// OBSLogoHistoryHandler — GET /api/obs/logo/history
// logo_1.png ~ logo_3.png 중 존재하는 파일을 mtime 기준 내림차순으로 반환한다.
func OBSLogoHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	dir := logoDir()
	type entry struct {
		name  string
		mtime int64
	}
	var entries []entry
	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("logo_%d.png", i)
		fi, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		entries = append(entries, entry{name: name, mtime: fi.ModTime().UnixNano()})
	}

	// mtime 내림차순 정렬
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].mtime > entries[j].mtime
	})

	paths := make([]string, 0, len(entries))
	for _, e := range entries {
		paths = append(paths, "/api/obs/logo/image?name="+e.name)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"paths": paths})
}

// OBSLogoImageHandler — GET /api/obs/logo/image?name=logo_N.png
// logo_1.png / logo_2.png / logo_3.png 만 허용. path traversal 방지.
func OBSLogoImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	name := filepath.Base(r.URL.Query().Get("name"))
	allowed := map[string]bool{
		"logo_1.png": true,
		"logo_2.png": true,
		"logo_3.png": true,
	}
	if !allowed[name] {
		http.Error(w, "허용되지 않는 파일명", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(logoDir(), name)
	http.ServeFile(w, r, filePath)
}

// OBSLogoApplyHandler — POST /api/obs/logo/apply {scene, position, scale, x, y}
func OBSLogoApplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Scene    string  `json:"scene"`
		Position string  `json:"position"` // top-left, top-right, bottom-left, bottom-right, custom
		Scale    float64 `json:"scale"`
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
		return
	}
	if body.Scene == "" {
		http.Error(w, "scene 필수", http.StatusBadRequest)
		return
	}
	if body.Scale <= 0 {
		body.Scale = 0.15
	}

	logoPath := logoFilePath()
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		http.Error(w, "로고 파일이 없습니다. 먼저 업로드하세요.", http.StatusBadRequest)
		return
	}

	m := obs.Get()

	// 기존 EP_Logo 입력이 있으면 제거
	inputName := "EP_Logo"
	m.RemoveInput(inputName) // 에러 무시 (없을 수 있음)

	// 이미지 소스 생성
	sceneItemID, err := m.CreateImageSource(body.Scene, inputName, logoPath, true)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	// 실제 이미지 크기 읽기 → OBS 스케일 계산
	// body.Scale = 캔버스 너비 대비 비율 (예: 0.15 → 1920 * 0.15 = 288px)
	imgFile, err2 := os.Open(logoPath)
	imgW, imgH := 300.0, 300.0
	if err2 == nil {
		if cfg, _, decErr := image.DecodeConfig(imgFile); decErr == nil && cfg.Width > 0 {
			imgW = float64(cfg.Width)
			imgH = float64(cfg.Height)
		}
		imgFile.Close()
	}
	targetW := 1920.0 * body.Scale            // 캔버스에서 차지할 목표 너비(px)
	obsScale := targetW / imgW                  // OBS에 적용할 실제 스케일
	logoRenderedW := targetW
	logoRenderedH := imgH * obsScale

	x, y := body.X, body.Y
	margin := 30.0
	switch body.Position {
	case "top-left":
		x, y = margin, margin
	case "top-right":
		x, y = 1920-logoRenderedW-margin, margin
	case "bottom-left":
		x, y = margin, 1080-logoRenderedH-margin
	case "bottom-right":
		x, y = 1920-logoRenderedW-margin, 1080-logoRenderedH-margin
	}

	// 트랜스폼 설정
	if err := m.SetItemTransform(body.Scene, sceneItemID, x, y, obsScale, obsScale); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "sceneItemId": sceneItemID, "warning": "트랜스폼 설정 실패: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":          true,
		"sceneItemId": sceneItemID,
		"inputName":   inputName,
	})
}

// OBSCameraDevicesHandler — GET /api/obs/camera/devices
func OBSCameraDevicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	m := obs.Get()
	devices, err := m.ListCameraDevices()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "devices": []obs.DeviceInfo{}})
		return
	}
	if devices == nil {
		devices = []obs.DeviceInfo{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"devices": devices})
}

// OBSCameraAddHandler — POST /api/obs/camera/add {scene, deviceId, inputName}
func OBSCameraAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Scene     string `json:"scene"`
		DeviceID  string `json:"deviceId"`
		InputName string `json:"inputName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
		return
	}
	if body.Scene == "" || body.DeviceID == "" {
		http.Error(w, "scene, deviceId 필수", http.StatusBadRequest)
		return
	}
	if body.InputName == "" {
		body.InputName = "카메라"
	}

	m := obs.Get()
	sceneItemID, err := m.CreateCameraSource(body.Scene, body.InputName, body.DeviceID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":          true,
		"sceneItemId": sceneItemID,
		"inputName":   body.InputName,
	})
}

// OBSSourceToggleHandler — POST /api/obs/sources/toggle {scene, sceneItemId, enabled}
func OBSSourceToggleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Scene       string `json:"scene"`
		SceneItemID int    `json:"sceneItemId"`
		Enabled     bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
		return
	}
	if body.Scene == "" || body.SceneItemID == 0 {
		http.Error(w, "scene, sceneItemId 필수", http.StatusBadRequest)
		return
	}

	m := obs.Get()
	if err := m.SetItemEnabled(body.Scene, body.SceneItemID, body.Enabled); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// OBSSetupDisplayHandler — POST /api/obs/setup-display
// displayScene에 EP_Display 브라우저 소스를 자동 생성합니다.
// 기존 EP_Display가 있으면 먼저 제거 후 재생성합니다.
func OBSSetupDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Scene string `json:"scene"` // 비어있으면 config.displayScene 사용
		URL   string `json:"url"`   // 비어있으면 http://localhost:8080/display
	}
	json.NewDecoder(r.Body).Decode(&body)

	m := obs.Get()
	cfg := m.GetConfig()

	sceneName := body.Scene
	if sceneName == "" {
		sceneName = cfg.DisplayScene
	}
	if sceneName == "" {
		http.Error(w, "씬 이름 필수 (config/obs.json displayScene 또는 body.scene)", http.StatusBadRequest)
		return
	}

	displayURL := body.URL
	if displayURL == "" {
		displayURL = "http://localhost:8080/display"
	}

	inputName := "EP_Display"
	if strings.Contains(displayURL, "/display/pdf") {
		inputName = "EP_PDF"
	}
	// 기존 소스 제거 (없어도 무시)
	m.RemoveInput(inputName)

	sceneItemID, err := m.CreateBrowserSource(sceneName, inputName, displayURL, 1920, 1080)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":          true,
		"sceneItemId": sceneItemID,
		"inputName":   inputName,
		"scene":       sceneName,
		"url":         displayURL,
	})
}

// OBSSourceRemoveHandler — POST /api/obs/sources/remove {inputName}
func OBSSourceRemoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		InputName string `json:"inputName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
		return
	}
	if body.InputName == "" {
		http.Error(w, "inputName 필수", http.StatusBadRequest)
		return
	}

	m := obs.Get()
	if err := m.RemoveInput(body.InputName); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// OBSSetupInitialHandler — POST /api/obs/setup-initial
// camera/monitor 씬 자동 생성 + 카메라·모니터캡처·EP_Display 소스 자동 추가
func OBSSetupInitialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		CameraDeviceID string `json:"cameraDeviceId"`
	}
	json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck

	m := obs.Get()
	status := m.GetStatus()
	if !status.Connected {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(&obs.InitialSetupResult{ //nolint:errcheck
			Success:        false,
			ScenesCreated:  []string{},
			SourcesCreated: []string{},
			Warnings:       []string{"OBS가 연결되지 않았습니다"},
		})
		return
	}

	result, err := m.SetupInitial(body.CameraDeviceID, logoFilePath())
	if err != nil {
		json.NewEncoder(w).Encode(&obs.InitialSetupResult{ //nolint:errcheck
			Success:        false,
			ScenesCreated:  []string{},
			SourcesCreated: []string{},
			Warnings:       []string{err.Error()},
		})
		return
	}

	json.NewEncoder(w).Encode(result) //nolint:errcheck
}
