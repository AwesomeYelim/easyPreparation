package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestMain — 프로젝트 루트로 CWD 변경 + CI용 최소 환경 보장
func TestMain(m *testing.M) {
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	if err := os.Chdir(projectRoot); err != nil {
		panic("프로젝트 루트로 이동 실패: " + err.Error())
	}

	// EASYPREP_DATA_DIR을 CWD로 설정 → ExecutePath()가 프로젝트 루트를 반환하도록 보장
	cwd, _ := os.Getwd()
	os.Setenv("EASYPREP_DATA_DIR", cwd)

	// config/ 디렉토리 없으면 생성 (CI 환경)
	os.MkdirAll("config", 0755)
	os.MkdirAll("data", 0755)

	// main_worship.json 없으면 최소 더미 생성
	configFile := "config/main_worship.json"
	createdDummy := false
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		os.WriteFile(configFile, []byte("[]"), 0644)
		createdDummy = true
	}

	code := m.Run()
	if createdDummy {
		os.Remove(configFile)
	}
	os.Exit(code)
}

// ---------- 1. 헬스체크 ----------

func TestSmokeHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	HealthCheck(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}

	var body struct {
		Status string                 `json:"status"`
		Checks map[string]interface{} `json:"checks"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("JSON decode error: %v", err)
	}

	// config + data가 보장되므로 unhealthy면 실패
	// (Ghostscript 없으면 degraded — 허용)
	if body.Status == "unhealthy" {
		t.Fatalf("health status is unhealthy — checks: %+v", body.Checks)
	}
	t.Logf("health status: %s", body.Status)
}

// ---------- 2. 예배 순서 왕복 (PUT → GET) ----------

func TestSmokeWorshipOrderRoundtrip(t *testing.T) {
	// 원본 파일 백업
	const configFile = "config/main_worship.json"
	original, readErr := os.ReadFile(configFile)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("config 파일 읽기 실패: %v", readErr)
	}
	t.Cleanup(func() {
		if readErr == nil {
			if err := os.WriteFile(configFile, original, 0644); err != nil {
				t.Errorf("원본 복원 실패: %v", err)
			}
		} else {
			os.Remove(configFile)
		}
	})

	// PUT: 테스트 데이터 저장
	putBody := map[string]interface{}{
		"type": "main_worship",
		"items": []map[string]interface{}{
			{"title": "테스트찬송", "info": "c_edit"},
			{"title": "테스트기도", "info": "edit"},
		},
	}
	putJSON, _ := json.Marshal(putBody)

	putReq := httptest.NewRequest(http.MethodPut, "/api/worship-order", bytes.NewReader(putJSON))
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()

	WorshipOrderHandler(putW, putReq)

	if putW.Code != http.StatusOK {
		t.Fatalf("PUT failed: status %d, body: %s", putW.Code, putW.Body.String())
	}

	// GET: 저장된 데이터 확인
	getReq := httptest.NewRequest(http.MethodGet, "/api/worship-order?type=main_worship", nil)
	getW := httptest.NewRecorder()

	WorshipOrderHandler(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("GET failed: status %d, body: %s", getW.Code, getW.Body.String())
	}

	var items []map[string]interface{}
	if err := json.NewDecoder(getW.Body).Decode(&items); err != nil {
		t.Fatalf("GET JSON decode error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0]["title"] != "테스트찬송" {
		t.Errorf("item[0] title mismatch: got %v", items[0]["title"])
	}
	if items[1]["info"] != "edit" {
		t.Errorf("item[1] info mismatch: got %v", items[1]["info"])
	}

	t.Log("worship-order roundtrip OK")
}

// ---------- 3. Display 상태 ----------

func TestSmokeDisplayStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/display/status", nil)
	w := httptest.NewRecorder()

	DisplayStatusHandler(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("JSON decode error: %v", err)
	}

	if _, ok := body["idx"]; !ok {
		t.Fatal("response missing 'idx' key")
	}
	if _, ok := body["count"]; !ok {
		t.Fatal("response missing 'count' key")
	}

	t.Logf("display status: idx=%.0f, count=%.0f", body["idx"], body["count"])
}
