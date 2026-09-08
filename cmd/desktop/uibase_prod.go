//go:build !dev

package main

// getUIBaseURL — prod 모드에서 WebView가 리디렉션할 UI URL (embed된 프론트를 Go 서버가 서빙)
func getUIBaseURL() string {
	return "http://localhost:8080"
}
