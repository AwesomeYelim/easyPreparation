package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/ptz"
)

// PTZConfigHandler — GET/POST /api/ptz/config
func PTZConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		cfg, err := ptz.LoadConfig()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(cfg)

	case http.MethodPost:
		var cfg ptz.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := ptz.SaveConfig(&cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// PTZGotoHandler — POST /api/ptz/goto  body: {"preset": 1}
func PTZGotoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var body struct {
		Preset int `json:"preset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := ptz.GotoPreset(body.Preset); err != nil {
		log.Printf("[ptz] GotoPreset(%d) 실패: %v", body.Preset, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// PTZStreamProxyHandler — GET /api/ptz/stream — 카메라 MJPEG 스트림 프록시
func PTZStreamProxyHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := ptz.LoadConfig()
	if err != nil || !cfg.Enabled || cfg.IP == "" {
		http.Error(w, "PTZ 카메라 미설정", http.StatusServiceUnavailable)
		return
	}

	port := cfg.Port
	if port == 0 {
		port = 80
	}
	streamPath := cfg.StreamPath
	if streamPath == "" {
		streamPath = "/stream"
	}

	targetURL := fmt.Sprintf("http://%s:%d%s", cfg.IP, port, streamPath)
	log.Printf("[ptz] stream proxy → %s", targetURL)

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)

	client := &http.Client{} // no timeout — streaming
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// PTZPresetsHandler — GET /api/ptz/presets — 카메라에 저장된 프리셋 목록 조회
func PTZPresetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	presets, err := ptz.GetPresets()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"presets": presets,
	})
}

// PTZPingHandler — GET /api/ptz/ping — 카메라 연결 상태 확인
func PTZPingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	latency, err := ptz.Ping()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":         true,
		"latency_ms": latency,
	})
}

// PTZSourceToggleHandler — POST /api/ptz/source  body: {"mode": "camera"|"slides"}
// OBS 씬을 카메라 씬 또는 모니터(display) 씬으로 수동 전환
func PTZSourceToggleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var body struct {
		Mode string `json:"mode"` // "camera" | "slides"
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	obsM := obs.Get()
	cfg := obsM.GetConfig()
	if body.Mode == "camera" && cfg.CameraScene != "" {
		obsM.SetSceneDirect(cfg.CameraScene)
	} else if body.Mode == "slides" && cfg.DisplayScene != "" {
		obsM.SetSceneDirect(cfg.DisplayScene)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
