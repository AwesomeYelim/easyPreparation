//go:build windows

package main

// 순서 패널 분리 창 — Windows 전용 동작 (Win32)

import (
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procSetWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procGetWindowRect    = user32.NewProc("GetWindowRect")
	procGetDpiForWindow  = user32.NewProc("GetDpiForWindow")
	procReleaseCapture   = user32.NewProc("ReleaseCapture")
	procPostMessage      = user32.NewProc("PostMessageW")
	procIsIconic         = user32.NewProc("IsIconic")
)

const (
	gwlpHwndParent  = ^uintptr(7) // -8 (GWLP_HWNDPARENT)
	wmNcLButtonDown = 0x00A1
	htCaption       = 2

	// 메인 창 레이아웃 (ProShell grid) — 도킹 영역 = 아이콘바(48) + 순서 패널(280) + 여유
	shellIconBarW = 48
	shellPanelW   = 280
	shellTopBarH  = 44
	dockZoneSlack = 60
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }

func hwndOf(w *application.WebviewWindow) uintptr {
	if w == nil {
		return 0
	}
	return uintptr(w.NativeWindow())
}

// setPanelOwner — 분리 창의 소유자를 메인 창으로 지정.
// 소유된 창은 항상 소유자 위에 있고, 소유자가 최소화되면 같이 숨는다 (OBS 플로팅 독과 같은 성격).
func setPanelOwner(panel, main *application.WebviewWindow) {
	p, m := hwndOf(panel), hwndOf(main)
	if p == 0 || m == 0 {
		return
	}
	procSetWindowLongPtr.Call(p, gwlpHwndParent, m)
}

// cursorPos — 화면 커서 위치 (물리 픽셀)
func cursorPos() (int, int, bool) {
	var pt point
	r, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y), r != 0
}

// beginWindowDrag — 지금 눌려 있는 마우스 버튼의 드래그를 분리 창 제목줄 드래그로 넘긴다.
// 메인 창(WebView)이 잡고 있던 마우스 캡처를 풀고, 분리 창에 "제목줄을 눌렀다"를 보내면
// OS 의 창 이동 루프가 시작돼 창이 커서를 따라온다. InvokeSync(UI 스레드) 안에서 호출할 것.
func beginWindowDrag(w *application.WebviewWindow) {
	h := hwndOf(w)
	if h == 0 {
		return
	}
	x, y, _ := cursorPos()
	lparam := uintptr(uint32(uint16(int16(x))) | uint32(uint16(int16(y)))<<16)
	procReleaseCapture.Call()
	procPostMessage.Call(h, wmNcLButtonDown, htCaption, lparam)
}

// cursorInDockZone — 커서가 메인 창의 순서 패널 자리(왼쪽) 위에 있는지
func cursorInDockZone(main *application.WebviewWindow) bool {
	h := hwndOf(main)
	if h == 0 {
		return false
	}
	if iconic, _, _ := procIsIconic.Call(h); iconic != 0 {
		return false
	}
	var rc rect
	if r, _, _ := procGetWindowRect.Call(h, uintptr(unsafe.Pointer(&rc))); r == 0 {
		return false
	}
	scale := 1.0
	if dpi, _, _ := procGetDpiForWindow.Call(h); dpi > 0 {
		scale = float64(dpi) / 96.0
	}
	x, y, ok := cursorPos()
	if !ok {
		return false
	}
	zoneRight := rc.Left + int32(float64(shellIconBarW+shellPanelW+dockZoneSlack)*scale)
	zoneTop := rc.Top + int32(float64(shellTopBarH)*scale)
	return int32(x) >= rc.Left && int32(x) <= zoneRight && int32(y) >= zoneTop && int32(y) <= rc.Bottom
}
