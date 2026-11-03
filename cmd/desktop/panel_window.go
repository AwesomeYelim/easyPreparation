package main

// ── 순서 패널 분리 창 (OBS 독 방식) ──
//
// 웹 UI 의 DOM 은 다른 OS 창으로 옮길 수 없으므로, 분리 창을 앱 시작 때 숨긴 채 미리 만들어 둔다.
// 떼기 = 위치 지정 + 표시, 붙이기 = 숨김. 창을 다시 만들지 않으니 떼는 순간 이미 그려진 화면이 뜬다.
// 분리 창도 서버 WS 클라이언트라 숨겨져 있는 동안에도 상태가 최신이다.
//
// Windows 전용 동작(메인 창 소유, 드래그 이어받기, 도킹 영역 판정)은 panel_window_windows.go,
// 그 외 OS 는 panel_window_other.go 의 no-op 이다.

import (
	"log"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"easyPreparation_1.0/internal/handlers"
)

type panelDock struct {
	win      *application.WebviewWindow
	dragging atomic.Bool  // 분리 창을 끄는 중 (도킹 영역 강조 판정용)
	hint     atomic.Bool  // 마지막으로 방송한 도킹 영역 강조 상태
	startW   atomic.Int64 // StartMove 시점 크기 — EndMove 때 바뀌었으면 이동이 아니라 크기 조절
	startH   atomic.Int64
}

const (
	panelDefaultW = 340
	panelDefaultH = 860
)

// initPanelWindow — 메인 창 표시 후 1회 호출. 숨긴 분리 창을 만들고, 지난 실행에서 분리돼 있었으면 표시.
func (a *App) initPanelWindow() {
	st := handlers.GetPanelDockState()
	application.InvokeSync(func() {
		w, h := st.W, st.H
		if w <= 0 || h <= 0 {
			w, h = panelDefaultW, panelDefaultH
		}
		win := a.wapp.Window.NewWithOptions(application.WebviewWindowOptions{
			Name:        "panel",
			Title:       "예배 순서",
			Width:       w,
			Height:      h,
			MinWidth:    280,
			MinHeight:   420,
			AlwaysOnTop: st.Pinned,
			Hidden:      true,
			URL:         getUIBaseURL() + "/remote/",
			Windows: application.WindowsWindow{
				HiddenOnTaskbar: true, // WS_EX_TOOLWINDOW — 작업 표시줄에 안 뜨고 얇은 제목줄
			},
		})
		a.panel.win = win
		if st.W > 0 && st.H > 0 {
			win.SetPosition(st.X, st.Y)
		}
		// 메인 창 소유 → 항상 메인 위, 메인 최소화 시 같이 숨김 (Windows)
		setPanelOwner(win, a.mainWin)

		// X 버튼 = 붙이기. 창은 없애지 않고 숨긴다. 앱 종료 중일 때만 실제로 닫힌다.
		win.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
			if a.quitting.Load() {
				return
			}
			e.Cancel()
			a.savePanelGeometry()
			win.Hide()
			handlers.OnPanelWindowClosed()
		})

		// 끄는 중: 메인 창 도킹 영역 위면 붙을 자리 강조 / 놓으면 도킹
		// (프로그램이 SetPosition 으로 옮길 때는 StartMove 가 없으므로 강조하지 않는다)
		win.OnWindowEvent(events.Windows.WindowStartMove, func(e *application.WindowEvent) {
			w, h := win.Size()
			a.panel.startW.Store(int64(w))
			a.panel.startH.Store(int64(h))
			a.panel.dragging.Store(true)
		})
		win.OnWindowEvent(events.Common.WindowDidMove, func(e *application.WindowEvent) {
			if !a.panel.dragging.Load() {
				return
			}
			a.updateDockHint(cursorInDockZone(a.mainWin))
		})
		win.OnWindowEvent(events.Windows.WindowEndMove, func(e *application.WindowEvent) {
			wasDragging := a.panel.dragging.Swap(false)
			// 크기 조절(WM_EXITSIZEMOVE 는 크기 조절에도 온다)로 끝난 경우는 도킹하지 않는다
			w, h := win.Size()
			resized := int64(w) != a.panel.startW.Load() || int64(h) != a.panel.startH.Load()
			inZone := wasDragging && !resized && cursorInDockZone(a.mainWin)
			a.updateDockHint(false)
			if inZone {
				win.Hide()
				handlers.OnPanelWindowClosed()
				return
			}
			a.savePanelGeometry()
		})
	})

	if st.Detached {
		a.showPanelWindow(st)
	}
	log.Println("[desktop] 순서 패널 분리 창 준비 (숨김)")
}

func (a *App) updateDockHint(active bool) {
	if a.panel.hint.Swap(active) != active {
		handlers.BroadcastPanelDockHint(active)
	}
}

// savePanelGeometry — 분리 창의 현재 위치·크기를 기록 (숨김·최소화 상태는 건너뜀)
func (a *App) savePanelGeometry() {
	win := a.panel.win
	if win == nil || !win.IsVisible() || win.IsMinimised() {
		return
	}
	x, y := win.Position()
	w, h := win.Size()
	handlers.SavePanelGeometry(x, y, w, h)
}

// showPanelWindow — 저장된 위치에 표시 (버튼 분리 / 재시작 복원)
func (a *App) showPanelWindow(st handlers.PanelDockState) {
	win := a.panel.win
	if win == nil {
		return
	}
	application.InvokeSync(func() {
		if st.W > 0 && st.H > 0 {
			win.SetPosition(st.X, st.Y)
		}
		win.SetAlwaysOnTop(st.Pinned)
		win.Show()
		win.Focus()
	})
}

// showPanelWindowAtCursor — 패널 머리를 끌어서 뗀 경우: 커서 아래에 표시하고 드래그를 OS 에 넘긴다
func (a *App) showPanelWindowAtCursor(st handlers.PanelDockState) {
	win := a.panel.win
	if win == nil {
		return
	}
	application.InvokeSync(func() {
		if x, y, ok := cursorPos(); ok {
			// 제목줄 안쪽을 잡은 것처럼 보이도록 약간 왼쪽·위로
			win.SetPosition(x-60, y-12)
		}
		win.SetAlwaysOnTop(st.Pinned)
		win.Show()
		win.Focus()
		beginWindowDrag(win)
	})
}

func (a *App) focusPanelWindow() {
	win := a.panel.win
	if win == nil {
		return
	}
	application.InvokeSync(func() {
		if win.IsMinimised() {
			win.UnMinimise()
		}
		win.Show()
		win.Focus()
	})
}

func (a *App) hidePanelWindow() {
	win := a.panel.win
	if win == nil {
		return
	}
	a.savePanelGeometry()
	application.InvokeSync(func() { win.Hide() })
}

func (a *App) setPanelPinned(pinned bool) {
	if win := a.panel.win; win != nil {
		win.SetAlwaysOnTop(pinned)
	}
}
