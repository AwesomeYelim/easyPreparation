//go:build !windows

package main

// 순서 패널 분리 창 — Windows 외 OS. 버튼으로 떼기·붙이기만 동작하고
// 끌어서 떼기·도킹 영역 판정·메인 창 소유는 하지 않는다 (macOS 는 추후 별도 구현).

import "github.com/wailsapp/wails/v3/pkg/application"

func setPanelOwner(panel, main *application.WebviewWindow) {}

func cursorPos() (int, int, bool) { return 0, 0, false }

func beginWindowDrag(w *application.WebviewWindow) {}

func cursorInDockZone(main *application.WebviewWindow) bool { return false }
