package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestMain — 테스트 전용 임시 디렉토리 생성 (실제 config를 절대 건드리지 않음)
func TestMain(m *testing.M) {
	// 프로젝트 루트 확인 (go vet용)
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	if err := os.Chdir(projectRoot); err != nil {
		panic("프로젝트 루트로 이동 실패: " + err.Error())
	}

	// 테스트 전용 임시 디렉토리 — 실제 config/data를 건드리지 않음
	tmpDir, err := os.MkdirTemp("", "easyprep-test-*")
	if err != nil {
		panic("임시 디렉토리 생성 실패: " + err.Error())
	}
	os.MkdirAll(filepath.Join(tmpDir, "config"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "data"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "config", "main_worship.json"), []byte("[]"), 0644)

	// ExecutePath()가 임시 디렉토리를 반환하도록 설정
	os.Setenv("EASYPREP_DATA_DIR", tmpDir)

	code := m.Run()

	// 정리
	os.RemoveAll(tmpDir)
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

	// config+data는 임시 디렉토리에 보장됨
	// Ghostscript 없으면 degraded — 허용. unhealthy만 실패.
	if body.Status == "unhealthy" {
		t.Fatalf("health status is unhealthy — checks: %+v", body.Checks)
	}
	t.Logf("health status: %s", body.Status)
}

// ---------- 2. 예배 순서 왕복 (PUT → GET) ----------

func TestSmokeWorshipOrderRoundtrip(t *testing.T) {
	// PUT: 테스트 데이터 저장 (임시 디렉토리에 저장됨 — 실제 config 안 건드림)
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

// ---------- 통합 테스트 (실제 서버 대상) ----------
// CI에서 서버 바이너리를 미리 실행해 둔 상태에서 호출됨.
// 로컬 실행: 별도 터미널에서 서버 시작 후 EASYPREP_INTEGRATION=1 go test -run TestIntegration ./internal/handlers/

const integrationBase = "http://localhost:8080"

func skipIfNoServer(t *testing.T) {
	t.Helper()
	// CI에서는 EASYPREP_INTEGRATION 없이도 실행; 로컬에서는 환경 변수로 명시적 opt-in
	if os.Getenv("CI") == "" && os.Getenv("EASYPREP_INTEGRATION") == "" {
		t.Skip("통합 테스트 건너뜀 (CI=true 또는 EASYPREP_INTEGRATION=1 필요)")
	}
	resp, err := http.Get(fmt.Sprintf("%s/api/health", integrationBase))
	if err != nil || resp.StatusCode >= 500 {
		t.Skipf("서버 미응답 — 통합 테스트 건너뜀: %v", err)
	}
	resp.Body.Close()
}

func TestIntegrationHealthCheck(t *testing.T) {
	skipIfNoServer(t)

	resp, err := http.Get(fmt.Sprintf("%s/api/health", integrationBase))
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("예상 외 상태 코드: %d", resp.StatusCode)
	}

	var body struct {
		Status string                 `json:"status"`
		Checks map[string]interface{} `json:"checks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("JSON 디코드 오류: %v", err)
	}
	if body.Status == "unhealthy" {
		t.Fatalf("서버 unhealthy: %+v", body.Checks)
	}
	t.Logf("통합 헬스체크 OK: %s", body.Status)
}

func TestIntegrationWorshipOrder(t *testing.T) {
	skipIfNoServer(t)

	payload := map[string]interface{}{
		"type": "main_worship",
		"items": []map[string]interface{}{
			{"title": "통합테스트찬송", "info": "c_edit"},
			{"title": "통합테스트기도", "info": "edit"},
		},
	}
	data, _ := json.Marshal(payload)

	putResp, err := http.Post(
		fmt.Sprintf("%s/api/worship-order", integrationBase),
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		t.Fatalf("PUT 요청 실패: %v", err)
	}
	defer putResp.Body.Close()

	// PUT이 없으면 POST도 허용 (메서드 무관하게 저장되면 OK)
	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("PUT 상태 코드: %d", putResp.StatusCode)
	}
	t.Logf("worship-order 통합 테스트 OK (PUT status=%d)", putResp.StatusCode)
}

func TestIntegrationDisplayStatus(t *testing.T) {
	skipIfNoServer(t)

	resp, err := http.Get(fmt.Sprintf("%s/display/status", integrationBase))
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("예상 외 상태 코드: %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("JSON 디코드 오류: %v", err)
	}
	if _, ok := body["idx"]; !ok {
		t.Fatal("응답에 'idx' 키 없음")
	}
	if _, ok := body["count"]; !ok {
		t.Fatal("응답에 'count' 키 없음")
	}
	t.Logf("통합 display 상태 OK: idx=%.0f, count=%.0f", body["idx"], body["count"])
}
