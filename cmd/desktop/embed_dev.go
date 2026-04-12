//go:build dev

package main

import "io/fs"

// 개발 모드에서는 embed 없음 — Next.js dev server 사용
func getFrontendFS() fs.FS {
	return nil
}

// getUIBaseURL — dev 모드에서 WebView가 리디렉션할 UI URL
func getUIBaseURL() string {
	return "http://localhost:3000"
}
