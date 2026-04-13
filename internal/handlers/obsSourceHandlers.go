package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	savePath := logoFilePath()
	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "파일 저장 실패", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "파일 쓰기 실패", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":   true,
		"path": savePath,
	})
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

	// 위치 계산 (1920x1080 기준, 로고 크기 추정 200x200 * scale)
	x, y := body.X, body.Y
	logoSize := 200.0 * body.Scale
	margin := 30.0
	switch body.Position {
	case "top-left":
		x, y = margin, margin
	case "top-right":
		x, y = 1920-logoSize-margin, margin
	case "bottom-left":
		x, y = margin, 1080-logoSize-margin
	case "bottom-right":
		x, y = 1920-logoSize-margin, 1080-logoSize-margin
	}

	// 트랜스폼 설정
	if err := m.SetItemTransform(body.Scene, sceneItemID, x, y, body.Scale, body.Scale); err != nil {
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
