package app

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// LocalBaseURL — 내장 HTTP 서버 주소 (api.StartServer 가 8080 고정)
const LocalBaseURL = "http://localhost:8080"

// WaitForServer — HTTP 서버가 응답할 때까지 대기 (최대 5초). 타임아웃 시에도 반환한다.
func WaitForServer(baseURL string) {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for i := 0; i < 50; i++ {
		resp, err := client.Get(baseURL + "/display/status")
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Println("[health] 서버 대기 타임아웃 — 계속 진행")
}

// RunHealthCheck — /api/health 를 호출해 "unhealthy" 가 아니면 true
func RunHealthCheck(baseURL string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/api/health")
	if err != nil {
		log.Printf("[health] 헬스체크 요청 실패: %v", err)
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[health] 헬스체크 응답 파싱 실패: %v", err)
		return false
	}
	log.Printf("[health] 헬스체크 결과: %s", result.Status)
	return result.Status != "unhealthy"
}
