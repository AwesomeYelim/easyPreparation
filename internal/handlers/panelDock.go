package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"easyPreparation_1.0/internal/path"
)

// ── 순서 패널 분리/도킹 (OBS 독 방식) ──
//
// 데스크톱(Wails v3)은 SetPanelWindowHooks 로 네이티브 창 제어 함수를 등록한다.
// 훅이 없으면(브라우저 모드) 분리 기능을 쓰지 않는다.
// 상태는 data/panel_dock.json 에 저장해 재시작 시 복원한다.

// PanelDockState — 분리 상태 + 분리 창 위치·크기·핀(항상 위)
type PanelDockState struct {
	Detached bool `json:"detached"`
	Pinned   bool `json:"pinned"`
	X        int  `json:"x"`
	Y        int  `json:"y"`
	W        int  `json:"w"`
	H        int  `json:"h"`
}

// PanelWindowHooks — 데스크톱 앱이 등록하는 네이티브 창 제어 함수
type PanelWindowHooks struct {
	Open      func(st PanelDockState) // 분리 창 표시 (저장된 위치)
	OpenDrag  func(st PanelDockState) // 분리 창을 커서 위치에 표시하고 진행 중인 드래그를 OS 에 넘김
	Focus     func()                  // 분리 창 앞으로 가져오기
	Close     func()                  // 분리 창 닫기
	SetPinned func(pinned bool)       // 항상 위 토글
}

var (
	panelMu    sync.Mutex
	panelState PanelDockState
	panelHooks *PanelWindowHooks
	panelOnce  sync.Once
)

func panelStatePath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "panel_dock.json")
}

func loadPanelState() {
	panelOnce.Do(func() {
		data, err := os.ReadFile(panelStatePath())
		if err != nil {
			return
		}
		if err := json.Unmarshal(data, &panelState); err != nil {
			log.Printf("[panel] 상태 파일 파싱 실패: %v", err)
		}
	})
}

// savePanelStateLocked — panelMu 를 잡은 상태에서 호출
func savePanelStateLocked() {
	data, err := json.MarshalIndent(panelState, "", "  ")
	if err != nil {
		return
	}
	p := panelStatePath()
	_ = os.MkdirAll(filepath.Dir(p), 0755)
	if err := os.WriteFile(p, data, 0644); err != nil {
		log.Printf("[panel] 상태 저장 실패: %v", err)
	}
}

func broadcastPanelDock(st PanelDockState) {
	BroadcastMessage("panel_dock", map[string]interface{}{
		"detached": st.Detached,
		"pinned":   st.Pinned,
	})
}

// SetPanelWindowHooks — 데스크톱 앱이 시작 시 등록
func SetPanelWindowHooks(h *PanelWindowHooks) {
	panelMu.Lock()
	panelHooks = h
	panelMu.Unlock()
}

// GetPanelDockState — 저장된 상태 (데스크톱 시작 시 복원용)
func GetPanelDockState() PanelDockState {
	loadPanelState()
	panelMu.Lock()
	defer panelMu.Unlock()
	return panelState
}

// SavePanelGeometry — 분리 창 위치·크기 기록 (창 이동/크기 변경/닫힘 시 데스크톱이 호출)
func SavePanelGeometry(x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	loadPanelState()
	panelMu.Lock()
	panelState.X, panelState.Y, panelState.W, panelState.H = x, y, w, h
	savePanelStateLocked()
	panelMu.Unlock()
}

// BroadcastPanelDockHint — 분리 창을 메인 창 도킹 영역 위로 끌고 있는지 (메인 창이 붙을 자리를 강조)
func BroadcastPanelDockHint(active bool) {
	BroadcastMessage("panel_dock_hint", map[string]interface{}{"active": active})
}

// OnPanelWindowClosed — 사용자가 분리 창을 닫거나(X) 도킹 영역에 놓음 → 도킹 복귀
func OnPanelWindowClosed() {
	loadPanelState()
	panelMu.Lock()
	if !panelState.Detached {
		panelMu.Unlock()
		return
	}
	panelState.Detached = false
	savePanelStateLocked()
	st := panelState
	panelMu.Unlock()
	broadcastPanelDock(st)
}

type panelStateResponse struct {
	PanelDockState
	Native bool `json:"native"` // true = 데스크톱 네이티브 창, false = 웹(window.open)
}

func writePanelState(w http.ResponseWriter) {
	panelMu.Lock()
	resp := panelStateResponse{PanelDockState: panelState, Native: panelHooks != nil}
	panelMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// PanelStateHandler — GET /api/panel/state
func PanelStateHandler(w http.ResponseWriter, r *http.Request) {
	loadPanelState()
	writePanelState(w)
}

// PanelDetachHandler — POST /api/panel/detach
// 데스크톱 앱에서만 분리한다(네이티브 창 표시). 브라우저 모드에서는 아무것도 하지 않고 현재 상태만 돌려준다.
func PanelDetachHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	loadPanelState()
	panelMu.Lock()
	hooks := panelHooks
	if hooks == nil {
		// 브라우저 모드(데스크톱 창 제어 없음)에서는 분리하지 않는다 — 패널이 접힌 채 사라지는 것 방지
		panelMu.Unlock()
		writePanelState(w)
		return
	}
	panelState.Detached = true
	savePanelStateLocked()
	st := panelState
	panelMu.Unlock()

	// ?drag=1 — 패널 머리를 끌어서 뗀 경우: 창을 커서 위치에 띄우고 드래그를 이어받는다
	if r.URL.Query().Get("drag") == "1" && hooks.OpenDrag != nil {
		hooks.OpenDrag(st)
	} else if hooks.Open != nil {
		hooks.Open(st)
	}
	broadcastPanelDock(st)
	writePanelState(w)
}

// PanelDockHandler — POST /api/panel/dock (붙이기)
func PanelDockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	loadPanelState()
	panelMu.Lock()
	wasDetached := panelState.Detached
	panelState.Detached = false
	savePanelStateLocked()
	st := panelState
	hooks := panelHooks
	panelMu.Unlock()

	// 상태를 먼저 docked 로 바꾼 뒤 창을 닫으므로 OnPanelWindowClosed 는 중복 방송하지 않는다
	if hooks != nil && hooks.Close != nil {
		hooks.Close()
	}
	if wasDetached {
		broadcastPanelDock(st)
	}
	writePanelState(w)
}

// PanelFocusHandler — POST /api/panel/focus (분리 창 앞으로)
func PanelFocusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	panelMu.Lock()
	hooks := panelHooks
	panelMu.Unlock()
	if hooks != nil && hooks.Focus != nil {
		hooks.Focus()
	}
	writePanelState(w)
}

// PanelPinHandler — POST /api/panel/pin {"pinned": bool} (항상 위)
func PanelPinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Pinned bool `json:"pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	loadPanelState()
	panelMu.Lock()
	panelState.Pinned = body.Pinned
	savePanelStateLocked()
	st := panelState
	hooks := panelHooks
	panelMu.Unlock()

	if hooks != nil && hooks.SetPinned != nil {
		hooks.SetPinned(body.Pinned)
	}
	broadcastPanelDock(st)
	writePanelState(w)
}
