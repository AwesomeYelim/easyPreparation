// Package httpx — HTTP 응답 헬퍼. 모든 API 응답(성공·에러)을 JSON 한 형식으로 통일한다.
// 에러는 항상 {"error": "<message>"} 이므로 프론트는 상태코드와 무관하게 body.error 를 읽으면 된다.
package httpx

import (
	"encoding/json"
	"net/http"
)

// JSON — status 코드와 함께 v 를 JSON 으로 쓴다.
func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error — {"error": msg} 를 status 코드와 함께 쓴다. http.Error 의 JSON 대체.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}
