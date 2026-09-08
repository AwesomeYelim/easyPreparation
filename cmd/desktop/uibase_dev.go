//go:build dev

package main

// getUIBaseURL — dev 모드에서 WebView가 리디렉션할 UI URL (Next.js dev server)
func getUIBaseURL() string {
	return "http://localhost:3000"
}
