package handlers

import (
	"easyPreparation_1.0/internal/assets"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/internal/safefile"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// 현재 예배 순서 메모리 저장
var (
	orderMu              sync.RWMutex
	currentOrder         []map[string]interface{}
	currentIdx           int
	currentSubPageIdx    int
	displayChurchName    string
	currentWorshipType   string // 현재 로드된 예배 타입 (config 변경 시 자동 갱신 판단용)
)

// ── Display 상태 파일 영속화 ──

func displayStatePath() string {
	execPath := path.ExecutePath("easyPreparation")
	return filepath.Join(execPath, "data", "display_state.json")
}

// lastFullOrderCount — 마지막으로 저장된 "정상" 순서의 항목 수 (급격한 감소 감지용)
var lastFullOrderCount int

// saveDisplayState — 현재 order+idx를 파일에 저장 (orderMu 잠긴 상태에서 호출하지 말 것)
// 기존 순서보다 항목이 급격히 줄어들면 별도 full_backup을 유지하여 복원 가능
func saveDisplayState() {
	orderMu.RLock()
	snapshot := deepCopyOrder(currentOrder)
	idx := currentIdx
	subPage := currentSubPageIdx
	cn := displayChurchName
	orderMu.RUnlock()

	if snapshot == nil {
		snapshot = []map[string]interface{}{}
	}

	newCount := len(snapshot)

	// 항목이 급격히 줄어들면 (5개 이상 → 3개 이하) 기존 상태를 full_backup으로 보존
	if lastFullOrderCount >= 5 && newCount < lastFullOrderCount/2 {
		execPath := path.ExecutePath("easyPreparation")
		backupPath := filepath.Join(execPath, "data", "display_state_full.json")
		// 현재 파일을 full_backup으로 복사 (이미 있으면 덮어쓰지 않음)
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			if src, readErr := os.ReadFile(displayStatePath()); readErr == nil {
				os.WriteFile(backupPath, src, 0644)
				log.Printf("[display] 순서 급감 감지 (%d→%d) — full_backup 보존", lastFullOrderCount, newCount)
			}
		}
	}

	// 정상 크기면 카운트 갱신
	if newCount >= 5 {
		lastFullOrderCount = newCount
	}

	state := map[string]interface{}{
		"items":      snapshot,
		"idx":        idx,
		"subPageIdx": subPage,
		"churchName": cn,
	}

	if err := safefile.WriteJSON(displayStatePath(), state); err != nil {
		log.Printf("[display] 상태 저장 실패: %v", err)
	}
}

// deepCopyOrder — slice of maps를 완전한 깊은 복사 (JSON round-trip)
// orderMu 잠긴 상태에서 호출할 것 (잠금 없이 내부적으로 안전)
func deepCopyOrder(src []map[string]interface{}) []map[string]interface{} {
	if len(src) == 0 {
		return nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		// fallback: 얕은 복사
		cp := make([]map[string]interface{}, len(src))
		copy(cp, src)
		return cp
	}
	var dst []map[string]interface{}
	if err := json.Unmarshal(data, &dst); err != nil {
		cp := make([]map[string]interface{}, len(src))
		copy(cp, src)
		return cp
	}
	return dst
}

// getOrderSnapshotLocked — orderMu 잠긴 상태에서 호출. 깊은 복사 + idx + churchName 반환.
func getOrderSnapshotLocked() ([]map[string]interface{}, int, string) {
	return deepCopyOrder(currentOrder), currentIdx, displayChurchName
}

// LoadDisplayState — 서버 시작 시 파일에서 복원
// 항목이 비정상적으로 적으면 (3개 이하) full_backup에서 복원 시도
func LoadDisplayState() {
	data, err := safefile.ReadJSONWithRecovery(displayStatePath())
	if err != nil {
		return
	}
	var state struct {
		Items      []map[string]interface{} `json:"items"`
		Idx        int                      `json:"idx"`
		SubPageIdx int                      `json:"subPageIdx"`
		ChurchName string                   `json:"churchName"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("[display] 상태 복원 실패: %v", err)
		return
	}

	// 항목이 비정상적으로 적으면 full_backup에서 복원 시도
	if len(state.Items) <= 3 {
		execPath := path.ExecutePath("easyPreparation")
		backupPath := filepath.Join(execPath, "data", "display_state_full.json")
		if backupData, readErr := os.ReadFile(backupPath); readErr == nil {
			var backupState struct {
				Items      []map[string]interface{} `json:"items"`
				Idx        int                      `json:"idx"`
				SubPageIdx int                      `json:"subPageIdx"`
				ChurchName string                   `json:"churchName"`
			}
			if json.Unmarshal(backupData, &backupState) == nil && len(backupState.Items) > len(state.Items) {
				log.Printf("[display] 항목 부족 (%d개) — full_backup에서 복원 (%d개)", len(state.Items), len(backupState.Items))
				state = backupState
				// 복원된 상태를 메인 파일에도 저장
				safefile.WriteJSON(displayStatePath(), map[string]interface{}{
					"items": backupState.Items, "idx": backupState.Idx, "subPageIdx": backupState.SubPageIdx, "churchName": backupState.ChurchName,
				})
			}
		}
	}

	if len(state.Items) == 0 {
		return
	}
	lastFullOrderCount = len(state.Items)
	// lyricsMap/sections 재생성 — 알고리즘 변경 시 캐시된 값이 남아있지 않도록
	processed := make([]map[string]interface{}, 0, len(state.Items))
	for _, item := range state.Items {
		delete(item, "lyricsMap") // 강제 재계산
		delete(item, "sections")  // startPage 공식 변경 시 구형 값 제거
		if info, _ := item["info"].(string); info == "lyrics_display" {
			// 가사 항목은 preprocessItem이 sections를 만들지 못해 리모컨에 "N페이지"로 표시됨
			// — append 때와 동일하게 가사 텍스트 기준으로 pages/sections 재생성
			lyr, _ := item["contents"].(string)
			rebuilt := preprocessLyricsItem(map[string]interface{}{
				"title":  item["title"],
				"lyrics": lyr,
				"bpm":    item["bpm"],
			})
			item["pages"] = rebuilt["pages"]
			item["sections"] = rebuilt["sections"]
			processed = append(processed, item)
			continue
		}
		processed = append(processed, preprocessItem(item))
	}
	orderMu.Lock()
	currentOrder = processed
	currentIdx = state.Idx
	currentSubPageIdx = state.SubPageIdx
	displayChurchName = state.ChurchName
	orderMu.Unlock()
	log.Printf("[display] 상태 복원: %d개 항목, idx=%d, subPage=%d", len(processed), state.Idx, state.SubPageIdx)
}

// ── 서버 사이드 자동 넘김 타이머 ──
var (
	timerMu          sync.Mutex
	timerEnabled     bool
	timerSpeedFactor float64 = 1.0
	timerCancel      chan struct{}
	timerCountdown   int
	timerCurIdx      int
	timerCurSubPage  int
)

const lordsPrayer = `하늘에 계신 우리 아버지,
아버지의 이름을 거룩하게 하시며, 아버지의 나라가 오게 하시며,
아버지의 뜻이 하늘에서와 같이 땅에서도 이루어지게 하소서.
오늘 우리에게 일용할 양식을 주시고,
우리가 우리에게 잘못한 사람을 용서하여 준 것같이 우리 죄를 용서하여 주시고,
우리를 시험에 빠지지 않게 하시고, 악에서 구하소서.
나라와 권능과 영광이 영원히 아버지의 것입니다. 아멘.`

const apostlesCreed = `나는 전능하신 아버지 하나님, 천지의 창조주를 믿습니다.
나는 그의 유일하신 아들, 우리 주 예수 그리스도를 믿습니다.
그는 성령으로 잉태되어 동정녀 마리아에게서 나시고,
본디오 빌라도에게 고난을 받아 십자가에 못 박혀 죽으시고,
장사된 지 사흘만에 죽은 자 가운데서 다시 살아나셨으며,
하늘에 오르시어 전능하신 아버지 하나님 우편에 앉아 계시다가,
거기로부터 살아있는 자와 죽은 자를 심판하러 오십니다.
나는 성령을 믿으며,
거룩한 공교회와 성도의 교제와 죄를 용서 받는 것과
몸의 부활과 영생을 믿습니다. 아멘.`

//go:embed html/display.html
var displayHTML string

// UpdateDisplayIdx — display HTML이 WS로 보고한 현재 위치 업데이트
func UpdateDisplayIdx(newIdx int, newSubPageIdx int) {
	orderMu.Lock()
	defer orderMu.Unlock()
	if newIdx >= 0 && newIdx < len(currentOrder) {
		currentIdx = newIdx
		currentSubPageIdx = newSubPageIdx
		go saveDisplayState()
	}
}

// GetCurrentTitle — OBS 씬 전환용 현재 항목 title 조회
func GetCurrentTitle() string {
	orderMu.RLock()
	defer orderMu.RUnlock()
	if currentIdx >= 0 && currentIdx < len(currentOrder) {
		t, _ := currentOrder[currentIdx]["title"].(string)
		return t
	}
	return ""
}

// GetCurrentInfo — 현재 항목의 info 필드 조회
func GetCurrentInfo() string {
	orderMu.RLock()
	defer orderMu.RUnlock()
	if currentIdx >= 0 && currentIdx < len(currentOrder) {
		info, _ := currentOrder[currentIdx]["info"].(string)
		return info
	}
	return ""
}

// GetCurrentSource — 현재 항목의 source 필드 조회 ("bible"/"lyrics" — 애드혹 추가 항목 표시)
func GetCurrentSource() string {
	orderMu.RLock()
	defer orderMu.RUnlock()
	if currentIdx >= 0 && currentIdx < len(currentOrder) {
		source, _ := currentOrder[currentIdx]["source"].(string)
		return source
	}
	return ""
}

// ── 서버 사이드 타이머 함수 ──

// calcSlideDelayLocked — 현재 항목/서브페이지에 대한 딜레이(초) 계산
// 호출 시점에 timerMu가 이미 잠겨 있어야 하고, orderMu는 잠기지 않은 상태여야 한다.
// timerCurIdx, timerCurSubPage를 인자로 받아 lock re-entry를 방지.
func calcSlideDelayLocked(curIdx, curSubPage int) int {
	orderMu.RLock()
	defer orderMu.RUnlock()

	if curIdx < 0 || curIdx >= len(currentOrder) {
		return 0
	}
	item := currentOrder[curIdx]
	title, _ := item["title"].(string)
	info, _ := item["info"].(string)

	bpm := 0
	switch v := item["bpm"].(type) {
	case float64:
		bpm = int(v)
	case int:
		bpm = v
	}

	// 전주/후주: 60초
	if title == "전주" || title == "후주" {
		return 60
	}

	// 성시교독: 표지 없음, 이미지만 15초
	if title == "성시교독" {
		return 15
	}

	// 찬송/헌금봉헌 이미지: 커버 5초, 나머지 15초
	if title == "찬송" || title == "헌금봉헌" {
		hasImages := false
		if imgs, ok := item["images"].([]string); ok && len(imgs) > 0 {
			hasImages = true
		} else if imgs, ok := item["images"].([]interface{}); ok && len(imgs) > 0 {
			hasImages = true
		}
		if hasImages {
			if curSubPage == 0 {
				return 5
			}
			return 15
		}
	}

	// 가사 (lyrics_display): 글자 수 비례 + BPM 보정
	if info == "lyrics_display" {
		baseTime := 8.0
		if bpm > 0 {
			baseTime = float64(16) / float64(bpm) * 60
		}

		// pages에서 현재 슬라이드와 전체 평균 글자 수 계산
		pages := extractPages(item)
		if len(pages) > 0 && curSubPage >= 0 && curSubPage < len(pages) {
			avgChars := 0.0
			for _, p := range pages {
				avgChars += float64(countKoreanChars(p))
			}
			avgChars /= float64(len(pages))
			if avgChars > 0 {
				curChars := float64(countKoreanChars(pages[curSubPage]))
				ratio := curChars / avgChars
				// 비율 범위 제한 (0.5 ~ 2.0)
				if ratio < 0.5 {
					ratio = 0.5
				} else if ratio > 2.0 {
					ratio = 2.0
				}
				return int(math.Round(baseTime * ratio))
			}
		}
		return int(math.Round(baseTime))
	}

	// 그 외 (대표기도, 말씀, 교회소식 등): 수동
	return 0
}

// extractPages — item에서 pages 배열 추출
func extractPages(item map[string]interface{}) []string {
	switch v := item["pages"].(type) {
	case []string:
		return v
	case []interface{}:
		pages := make([]string, 0, len(v))
		for _, p := range v {
			if s, ok := p.(string); ok {
				pages = append(pages, s)
			}
		}
		return pages
	}
	return nil
}

// countKoreanChars — 공백/줄바꿈 제외한 글자 수 (한글+영문+숫자)
func countKoreanChars(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			count++
		}
	}
	return count
}

// restartServerTimer — 타이머 재시작 (timerMu 외부에서 호출)
func restartServerTimer() {
	timerMu.Lock()

	// 기존 타이머 취소
	if timerCancel != nil {
		close(timerCancel)
	}
	cancel := make(chan struct{})
	timerCancel = cancel

	if !timerEnabled {
		timerCountdown = 0
		timerMu.Unlock()
		broadcastTimerState()
		return
	}

	// timerMu를 잠근 상태에서 timer 변수를 로컬로 복사한 뒤 Unlock → calcSlideDelayLocked 호출
	// (calcSlideDelayLocked 내부에서 orderMu.RLock을 사용하므로 timerMu를 먼저 풀어야 교착 방지)
	curIdx := timerCurIdx
	curSubPage := timerCurSubPage
	speedFactor := timerSpeedFactor
	timerMu.Unlock()

	delay := calcSlideDelayLocked(curIdx, curSubPage)
	if delay <= 0 {
		timerMu.Lock()
		timerCountdown = 0
		timerMu.Unlock()
		broadcastTimerState()
		return
	}

	// 속도 적용
	delay = int(math.Round(float64(delay) / speedFactor))
	if delay < 1 {
		delay = 1
	}

	timerMu.Lock()
	timerCountdown = delay

	// goroutine 시작 후 Unlock — 빠른 호출 시 다중 goroutine 누수 방지
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-cancel:
				return
			case <-ticker.C:
				timerMu.Lock()
				timerCountdown--
				if timerCountdown <= 0 {
					timerCountdown = 0
					curI := timerCurIdx
					curS := timerCurSubPage
					timerMu.Unlock()
					broadcastTimerState()
					log.Printf("[timer] auto-navigate next (idx=%d subPage=%d)", curI, curS)
					BroadcastMessage("navigate", map[string]interface{}{"direction": "next"})
					return
				}
				timerMu.Unlock()
				broadcastTimerState()
			}
		}
	}()

	timerMu.Unlock()
	broadcastTimerState()
}

// stopServerTimer — 타이머 정지
func stopServerTimer() {
	timerMu.Lock()
	timerEnabled = false
	if timerCancel != nil {
		close(timerCancel)
		timerCancel = nil
	}
	timerCountdown = 0
	timerMu.Unlock()
	broadcastTimerState()
}

// broadcastTimerState — 제어판에 타이머 상태 전송
func broadcastTimerState() {
	timerMu.Lock()
	state := map[string]interface{}{
		"enabled":     timerEnabled,
		"countdown":   timerCountdown,
		"idx":         timerCurIdx,
		"subPageIdx":  timerCurSubPage,
		"speedFactor": timerSpeedFactor,
	}
	timerMu.Unlock()
	BroadcastMessage("timer_state", state)
}

// OnPositionUpdate — Display에서 위치 보고 시 호출 (websocket.go에서 호출)
func OnPositionUpdate(newIdx, newSubPage int) {
	timerMu.Lock()
	changed := newIdx != timerCurIdx || newSubPage != timerCurSubPage
	timerCurIdx = newIdx
	timerCurSubPage = newSubPage
	enabled := timerEnabled
	timerMu.Unlock()

	if enabled && changed {
		restartServerTimer()
	}
}

// ── /display/overlay — OBS 방송용 텍스트 오버레이 ──

//go:embed html/display-overlay.html
var displayOverlayHTML string

// DisplayOverlayHandler — GET /display/overlay
func DisplayOverlayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(displayOverlayHTML))
}

// DisplayHandler — GET /display
func DisplayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(displayHTML))
}

// DisplayAssetsHandler — GET /display/assets/{name}.png
// 배경 이미지 서빙 (캐시 방지 — Studio에서 교체 즉시 반영)
func DisplayAssetsHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Path)
	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "data", "templates", "display", name)
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, imgPath)
}

// DisplayBgHandler — GET /display/bg
// 기본 배경 이미지 서빙 (data/default_bg.png, 없으면 204)
func DisplayBgHandler(w http.ResponseWriter, r *http.Request) {
	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "data", "default_bg.png")
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.ServeFile(w, r, imgPath)
}

// DisplayFontHandler — GET /display/font/{name}
func DisplayFontHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Path)
	execPath := path.ExecutePath("easyPreparation")
	fontPath := filepath.Join(execPath, "public", "font", name)
	http.ServeFile(w, r, fontPath)
}

// DisplayTmpHandler — GET /display/tmp/{category}/{num}/{page}.png
// 찬송/교독 PNG 서빙 (예: /display/tmp/hymn/635/1.png)
func DisplayTmpHandler(w http.ResponseWriter, r *http.Request) {
	rel := strings.TrimPrefix(r.URL.Path, "/display/tmp/")
	rel = filepath.FromSlash(rel)
	execPath := path.ExecutePath("easyPreparation")
	cacheRoot := filepath.Join(execPath, "data", "cache")
	imgPath := filepath.Clean(filepath.Join(cacheRoot, rel))
	if !strings.HasPrefix(imgPath, cacheRoot+string(filepath.Separator)) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, imgPath)
}

// DisplayOrderHandler — POST /display/order
// 전체 예배 순서 수신 → 성경 본문 자동 조회 → WebSocket broadcast
func DisplayOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var order []map[string]interface{}
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// wrapper format: {"items": [...], "churchName": "..."} 또는 plain array [...]
	var wrapper struct {
		Items        []map[string]interface{} `json:"items"`
		ChurchName   string                   `json:"churchName"`
		Email        string                   `json:"email"`
		Preprocessed bool                     `json:"preprocessed"`
	}
	var displayEmail string
	var skipPreprocess bool
	var newChurchName string
	// wrapper format 판별: JSON이 '{' 로 시작하면 wrapper, '[' 로 시작하면 plain array
	rawStr := strings.TrimSpace(string(raw))
	if len(rawStr) > 0 && rawStr[0] == '{' {
		if err := json.Unmarshal(raw, &wrapper); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		order = wrapper.Items
		newChurchName = wrapper.ChurchName
		displayEmail = wrapper.Email
		skipPreprocess = wrapper.Preprocessed
	} else {
		if err := json.Unmarshal(raw, &order); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "예배 화면 준비 중...",
		"total":   len(order),
	})

	// 항목별 전처리 (이미 전처리된 이력 데이터는 스킵)
	if !skipPreprocess {
		for i, item := range order {
			title, _ := item["title"].(string)
			BroadcastMessage("display_loading", map[string]interface{}{
				"message": fmt.Sprintf("%s 처리 중... (%d/%d)", title, i+1, len(order)),
				"current": i + 1,
				"total":   len(order),
			})
			delete(item, "sections")  // stale sections 강제 제거 (loadCurrentOrder와 동일)
			delete(item, "lyricsMap") // stale lyricsMap 강제 제거
			info, _ := item["info"].(string)
			if info == "lyrics_display" {
				order[i] = preprocessLyricsItem(item)
			} else {
				order[i] = preprocessItem(item)
			}
		}
	}

	// 말씀 항목에 직전 성경봉독의 구절 참조를 bibleRef로 주입
	var lastBibleRef string
	for i := range order {
		title, _ := order[i]["title"].(string)
		info, _ := order[i]["info"].(string)
		if strings.HasPrefix(info, "b_") && title != "말씀" {
			if ref, ok := order[i]["obj"].(string); ok && ref != "" && ref != "-" {
				lastBibleRef = ref
			}
		}
		if title == "말씀" && lastBibleRef != "" {
			order[i]["bibleRef"] = lastBibleRef
		}
	}

	// 기존 order에서 lyrics/bible source 항목 보존 (깊은 복사로 보존)
	orderMu.RLock()
	var preserved []map[string]interface{}
	for _, item := range currentOrder {
		if src, ok := item["source"].(string); ok && (src == "lyrics" || src == "bible") {
			preserved = append(preserved, item)
		}
	}
	orderMu.RUnlock()

	if len(preserved) > 0 {
		order = append(order, preserved...)
	}

	// 새 순서 로드 시 타이머 초기화
	stopServerTimer()

	orderMu.Lock()
	currentOrder = order
	currentIdx = 0
	displayChurchName = newChurchName
	cn := displayChurchName
	orderMu.Unlock()

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "준비 완료!",
		"done":    true,
	})

	BroadcastMessage("order", map[string]interface{}{"items": order, "churchName": cn})
	go saveDisplayState()

	// OBS: 첫 항목 씬 전환
	if len(order) > 0 {
		if title, ok := order[0]["title"].(string); ok {
			go obs.Get().SwitchScene(title)
		}
	}

	// 생성 이력 기록 (order 포함) — email 없으면 RecordGeneration 내부에서 "local@localhost" 사용
	go RecordGeneration(displayEmail, "display", time.Now().Format("060102"), "", "success", order)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayNavigateHandler — POST /display/navigate
// 운영자 UI에서 원격 슬라이드 이동 (서브페이지 포함)
// 서버는 currentIdx를 업데이트하지 않음 — display HTML이 position 메시지로 보고
func DisplayNavigateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var payload struct {
		Direction   string `json:"direction"`   // "next" | "prev" | "jump_sub"
		SubPageIdx  int    `json:"subPageIdx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	msg := map[string]interface{}{"direction": payload.Direction}
	if payload.Direction == "jump_sub" {
		msg["subPageIdx"] = payload.SubPageIdx
	}

	log.Printf("[navigate] direction=%s broadcast to clients", payload.Direction)
	BroadcastMessage("navigate", msg)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// DisplayPushHandler — POST /display/push (단독 항목 push, 기존 호환)
func DisplayPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	info, _ := payload["info"].(string)
	contents, _ := payload["contents"].(string)
	obj, _ := payload["obj"].(string)

	if strings.HasPrefix(info, "b_") && strings.TrimSpace(contents) == "" && obj != "" {
		versionID := 1
		if vid, ok := payload["versionId"].(float64); ok && vid > 0 {
			versionID = int(vid)
		}
		text, humanRef := fetchBibleTextWithVersion(obj, versionID)
		if text != "" {
			payload["contents"] = text
		}
		if humanRef != "" {
			payload["obj"] = humanRef
		}
	}

	BroadcastMessage("display", payload)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// DisplayJumpHandler — POST /display/jump
// 순서 목록에서 특정 항목으로 점프
func DisplayJumpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var payload struct {
		Index      int `json:"index"`
		SubPageIdx int `json:"subPageIdx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	if payload.Index < 0 || payload.Index >= len(currentOrder) {
		orderMu.Unlock()
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}
	currentIdx = payload.Index
	currentSubPageIdx = payload.SubPageIdx
	item := currentOrder[currentIdx]
	title, _ := item["title"].(string)
	info, _ := item["info"].(string)
	orderMu.Unlock()

	navPayload := map[string]interface{}{
		"direction":  "jump",
		"idx":        payload.Index,
		"subPageIdx": payload.SubPageIdx,
		"title":      title,
		"info":       info,
	}

	// OBS 씬 전환 + PTZ 프리셋은 여기서 직접 호출하지 않는다.
	// display 클라이언트가 이 navigate 브로드캐스트를 받아 자기 위치를 position으로 재보고하면
	// websocket.go의 디바운스 로직이 동일한 전환을 실행한다 (next/prev와 동일 경로).
	// 과거엔 여기서도 직접 호출해서 두 경로가 겹쳐 PTZ GotoPreset이 300ms 간격으로 중복 발사되며
	// 카메라가 씬 전환 직후 한 번 더 깜박이는 버그가 있었다.

	log.Printf("[jump] idx=%d title=%s broadcast to clients", payload.Index, title)
	BroadcastMessage("navigate", navPayload)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "idx": payload.Index})
}

// DisplayLyricsOrderHandler — POST /display/lyrics-order
// 가사 텍스트 → Display용 슬라이드 목록으로 변환하여 전송
func DisplayLyricsOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Songs []struct {
			Title  string `json:"title"`
			Lyrics string `json:"lyrics"`
			BPM    int    `json:"bpm"`
		} `json:"songs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "찬양 가사 준비 중...",
		"total":   len(payload.Songs),
	})

	var order []map[string]interface{}
	for _, song := range payload.Songs {
		songMap := map[string]interface{}{
			"title":  song.Title,
			"lyrics": song.Lyrics,
			"bpm":    song.BPM,
		}
		order = append(order, preprocessLyricsItem(songMap))
	}

	// 새 순서 로드 시 타이머 초기화
	stopServerTimer()

	orderMu.Lock()
	currentOrder = order
	currentIdx = 0
	cn := displayChurchName
	orderMu.Unlock()

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "준비 완료!",
		"done":    true,
	})

	BroadcastMessage("order", map[string]interface{}{"items": order, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayAppendHandler — POST /display/append
// 기존 순서에 항목을 추가 (교체가 아닌 append)
func DisplayAppendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Items    []map[string]interface{} `json:"items"`
		Source   string                   `json:"source"`
		AfterIdx *int                     `json:"afterIdx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if len(payload.Items) == 0 {
		http.Error(w, "No items", http.StatusBadRequest)
		return
	}

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "항목 추가 준비 중...",
		"total":   len(payload.Items),
	})

	// 항목별 전처리 + source 태깅
	var processed []map[string]interface{}
	for i, item := range payload.Items {
		info, _ := item["info"].(string)
		title, _ := item["title"].(string)

		BroadcastMessage("display_loading", map[string]interface{}{
			"message": fmt.Sprintf("%s 처리 중... (%d/%d)", title, i+1, len(payload.Items)),
			"current": i + 1,
			"total":   len(payload.Items),
		})

		var result map[string]interface{}
		if info == "lyrics_display" {
			result = preprocessLyricsItem(item)
		} else {
			result = preprocessItem(item)
		}
		// source 태깅 (lyrics/bible)
		if payload.Source != "" {
			result["source"] = payload.Source
		}
		processed = append(processed, result)
	}

	orderMu.Lock()
	// afterIdx가 지정되면 해당 위치 뒤에 삽입, 아니면 끝에 추가
	if payload.AfterIdx != nil && *payload.AfterIdx >= 0 && *payload.AfterIdx < len(currentOrder) {
		insertAt := *payload.AfterIdx + 1
		tail := make([]map[string]interface{}, len(currentOrder[insertAt:]))
		copy(tail, currentOrder[insertAt:])
		currentOrder = append(currentOrder[:insertAt], processed...)
		currentOrder = append(currentOrder, tail...)
	} else {
		currentOrder = append(currentOrder, processed...)
	}
	order, idx, cn := getOrderSnapshotLocked()
	subPage := currentSubPageIdx
	orderMu.Unlock()

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "준비 완료!",
		"done":    true,
	})

	// append는 현재 재생 위치를 바꾸지 않으므로 subPageIdx를 함께 보내 클라이언트 리셋 방지
	BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "subPageIdx": subPage, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayRemoveHandler — POST /display/remove
// 순서 목록에서 특정 인덱스의 항목 제거
func DisplayRemoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Index int `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	if payload.Index < 0 || payload.Index >= len(currentOrder) {
		orderMu.Unlock()
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}
	removedCurrent := payload.Index == currentIdx
	currentOrder = append(currentOrder[:payload.Index], currentOrder[payload.Index+1:]...)
	// currentIdx 보정
	if payload.Index < currentIdx {
		currentIdx--
	} else if payload.Index == currentIdx && currentIdx >= len(currentOrder) && len(currentOrder) > 0 {
		currentIdx = len(currentOrder) - 1
	}
	if currentIdx < 0 {
		currentIdx = 0
	}
	// 재생 중이던 항목이 삭제된 경우에만 subPage 리셋
	if removedCurrent {
		currentSubPageIdx = 0
	}
	order, idx, cn := getOrderSnapshotLocked()
	subPage := currentSubPageIdx
	orderMu.Unlock()

	BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "subPageIdx": subPage, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayItemPatchHandler — POST /display/item-patch
// 개별 항목의 cameraView, ptzPreset 필드만 조용히 업데이트 (WS broadcast 없음)
func DisplayItemPatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Index     int  `json:"index"`
		PtzPreset int  `json:"ptzPreset"`  // 1~9 | 0 (삭제)
		HasPtzPreset bool `json:"hasPtzPreset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	if payload.Index < 0 || payload.Index >= len(currentOrder) {
		orderMu.Unlock()
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}
	item := currentOrder[payload.Index]
	if payload.HasPtzPreset {
		if payload.PtzPreset <= 0 {
			delete(item, "ptzPreset")
		} else {
			item["ptzPreset"] = float64(payload.PtzPreset)
		}
	}
	orderMu.Unlock()

	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// DisplayTimerHandler — POST /display/timer
// 제어판에서 서버 타이머 직접 제어
func DisplayTimerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var payload struct {
		Action string  `json:"action"` // enable/disable/toggle/repeat/restart/speed
		Factor float64 `json:"factor"` // speed 조절 배율
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[timer] action=%s factor=%.1f", payload.Action, payload.Factor)

	switch payload.Action {
	case "enable":
		timerMu.Lock()
		timerEnabled = true
		timerMu.Unlock()
		restartServerTimer()
	case "disable":
		stopServerTimer()
	case "toggle":
		timerMu.Lock()
		timerEnabled = !timerEnabled
		wasEnabled := timerEnabled
		timerMu.Unlock()
		if wasEnabled {
			restartServerTimer()
		} else {
			stopServerTimer()
		}
	case "repeat":
		// 현재 항목의 서브페이지 0으로 돌아가기
		timerMu.Lock()
		timerCurSubPage = -1 // force change detection
		timerMu.Unlock()
		BroadcastMessage("navigate", map[string]interface{}{
			"direction":  "jump_sub",
			"subPageIdx": 0,
		})
	case "restart":
		// 현재 항목의 처음으로
		orderMu.RLock()
		idx := currentIdx
		orderMu.RUnlock()
		timerMu.Lock()
		timerCurIdx = -1
		timerCurSubPage = -1
		timerMu.Unlock()
		BroadcastMessage("navigate", map[string]interface{}{
			"direction": "jump",
			"idx":       idx,
		})
	case "speed":
		if payload.Factor > 0 {
			timerMu.Lock()
			timerSpeedFactor = payload.Factor
			timerMu.Unlock()
			broadcastTimerState()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// DisplayReorderHandler — POST /display/reorder
// 순서 목록에서 항목 위치 이동 (드래그 앤 드롭)
func DisplayReorderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		From int `json:"from"`
		To   int `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	n := len(currentOrder)
	if payload.From < 0 || payload.From >= n || payload.To < 0 || payload.To >= n {
		orderMu.Unlock()
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	// 새 슬라이스에 복사 (원본 배열 변경 방지)
	newOrder := make([]map[string]interface{}, 0, n)
	item := currentOrder[payload.From]
	for i, v := range currentOrder {
		if i == payload.From {
			continue
		}
		if len(newOrder) == payload.To {
			newOrder = append(newOrder, item)
		}
		newOrder = append(newOrder, v)
	}
	if len(newOrder) == payload.To {
		newOrder = append(newOrder, item)
	}
	currentOrder = newOrder

	// currentIdx 보정
	if currentIdx == payload.From {
		currentIdx = payload.To
	} else {
		if payload.From < currentIdx {
			currentIdx--
		}
		if payload.To <= currentIdx {
			currentIdx++
		}
	}
	if currentIdx < 0 {
		currentIdx = 0
	} else if currentIdx >= n {
		currentIdx = n - 1
	}

	order, idx, cn := getOrderSnapshotLocked()
	orderMu.Unlock()

	BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayChurchNameHandler — POST /display/church-name
// 교회명 변경 시 Display에 즉시 반영
func DisplayChurchNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		ChurchName string `json:"churchName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	displayChurchName = body.ChurchName
	order, idx, cn := getOrderSnapshotLocked()
	orderMu.Unlock()

	BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// DisplayStatusHandler — GET /display/status
// 현재 상태 반환 (idx, 항목 목록, OBS 상태)
func DisplayStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	orderMu.RLock()
	items, idx, _ := getOrderSnapshotLocked()
	subPageIdx := currentSubPageIdx
	count := len(items)

	var title string
	if idx >= 0 && idx < count {
		title, _ = items[idx]["title"].(string)
	}
	orderMu.RUnlock()

	obsStatus := obs.Get().GetStatus()

	streamStatus := obs.Get().GetStreamStatus()

	timerMu.Lock()
	tEnabled := timerEnabled
	timerMu.Unlock()

	// 스케줄러 활성 여부
	scheduleMu.RLock()
	schedActive := false
	for _, e := range scheduleConf.Entries {
		if e.Enabled {
			schedActive = true
			break
		}
	}
	scheduleMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"idx":            idx,
		"subPageIdx":     subPageIdx,
		"count":          count,
		"title":          title,
		"items":          items,
		"obs":            obsStatus,
		"stream":         streamStatus,
		"timerEnabled":   tEnabled,
		"scheduleActive": schedActive,
	})
}

// refreshDisplayIfNeeded — config 저장 시 현재 display 순서의 변경 항목을 자동 재전처리
// config의 각 항목 obj 값이 display의 같은 title 항목과 다르면 해당 항목만 재전처리
func refreshDisplayIfNeeded(configItems []map[string]interface{}) {
	orderMu.RLock()
	if len(currentOrder) == 0 {
		orderMu.RUnlock()
		return
	}
	// 현재 display 항목의 key→index 맵 (title은 "찬송"처럼 같은 항목이 여러 개 있을 수 있어 매칭 키로 부적합)
	displayMap := make(map[string]int)
	for i, item := range currentOrder {
		key, _ := item["key"].(string)
		if key == "" {
			continue
		}
		if _, exists := displayMap[key]; !exists {
			displayMap[key] = i
		}
	}
	orderMu.RUnlock()

	var changed bool
	for _, cfgItem := range configItems {
		key, _ := cfgItem["key"].(string)
		title, _ := cfgItem["title"].(string)
		cfgObj, _ := cfgItem["obj"].(string)
		if key == "" {
			continue
		}
		displayIdx, exists := displayMap[key]
		if !exists {
			continue
		}
		orderMu.RLock()
		displayObj, _ := currentOrder[displayIdx]["obj"].(string)
		orderMu.RUnlock()

		if cfgObj != displayObj {
			log.Printf("[display-refresh] %s 변경 감지: %q → %q, 재전처리", title, displayObj, cfgObj)
			processed := preprocessItem(cfgItem)
			processed = buildSections(processed)

			orderMu.Lock()
			if displayIdx < len(currentOrder) {
				currentOrder[displayIdx] = processed
			}
			orderMu.Unlock()
			changed = true
		}
	}

	if changed {
		orderMu.RLock()
		order, idx, cn := deepCopyOrder(currentOrder), currentIdx, displayChurchName
		orderMu.RUnlock()
		BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "churchName": cn})
		go saveDisplayState()
		log.Printf("[display-refresh] 변경 항목 재전처리 완료 → WS broadcast")
	}
}

// preprocessItem — 단일 항목 전처리 (성경, 신앙고백, 주기도문, 교회소식, 찬송/교독 이미지)
func preprocessItem(item map[string]interface{}) map[string]interface{} {
	info, _ := item["info"].(string)
	obj, _ := item["obj"].(string)
	title, _ := item["title"].(string)

	// b_edit: 성경 본문 자동 조회
	if strings.HasPrefix(info, "b_") && obj != "" {
		versionID := 1
		if vid, ok := item["versionId"].(float64); ok && vid > 0 {
			versionID = int(vid)
		}
		text, humanRef := fetchBibleTextWithVersion(obj, versionID)
		if text != "" {
			item["contents"] = text
			delete(item, "sections") // contents 변경 시 stale sections 강제 제거
		}
		if humanRef != "" {
			item["obj"] = humanRef
		}
	}

	// 신앙고백 (사도신경): 본문 자동 삽입
	if title == "신앙고백" {
		item["contents"] = apostlesCreed
	}

	// 주기도문: 본문 자동 삽입
	if title == "주기도문" {
		item["contents"] = lordsPrayer
	}

	// 항목별 배경 이미지 — data/templates/display/{title}.png/.jpg 자동 매핑
	// 매칭 이미지 없으면 배경 없음 (비디오 배경 또는 검정 배경)
	{
		execPath := path.ExecutePath("easyPreparation")
		displayDir := filepath.Join(execPath, "data", "templates", "display")
		for _, ext := range []string{".png", ".jpg", ".jpeg"} {
			bgPath := filepath.Join(displayDir, title+ext)
			if info, err := os.Stat(bgPath); err == nil {
				modTime := strconv.FormatInt(info.ModTime().Unix(), 10)
				item["bgImage"] = "/display/assets/" + url.PathEscape(title+ext) + "?v=" + modTime
				break
			}
		}
	}

	// 교회소식: children → contents 계층 텍스트 전처리
	if info == "notice" || strings.Contains(title, "교회소식") {
		if rawChildren, ok := item["children"]; ok {
			childrenJSON, _ := json.Marshal(rawChildren)
			var children []map[string]interface{}
			if json.Unmarshal(childrenJSON, &children) == nil && len(children) > 0 {
				item["contents"] = formatChurchNews(children, 1)
			}
		}
	}

	// 찬송/헌금봉헌/성시교독: Google Drive PDF → PNG 변환
	if title == "찬송" || title == "헌금봉헌" || title == "성시교독" {
		if images := fetchDisplayImages(title, obj); len(images) > 0 {
			item["images"] = images

			// 찬송: 가사↔페이지 자동 매핑
			if title == "찬송" || title == "헌금봉헌" {
				var hymnNum int
				for _, r := range obj {
					if r >= '0' && r <= '9' {
						hymnNum = hymnNum*10 + int(r-'0')
					}
				}
				if hymnNum > 0 {
					if lyrics := fetchHymnLyrics(hymnNum); lyrics != "" {
						verses := splitVerses(lyrics)
						if lm := mapLyricsToPages(verses, len(images)); len(lm) > 0 {
							item["lyricsMap"] = lm
						}
					}
				}
			}
		}
	}

	// ── sections 생성 (제어판 토글용) ──
	// Display HTML의 서브페이지 분할 로직과 동일하게 맞춤
	item = buildSections(item)

	return item
}

// buildSections — 항목의 서브페이지 구조에 맞춰 sections 배열 생성
func buildSections(item map[string]interface{}) map[string]interface{} {
	info, _ := item["info"].(string)
	title, _ := item["title"].(string)
	contents, _ := item["contents"].(string)

	// 이미 sections가 있으면 스킵 (lyrics 등)
	if _, ok := item["sections"]; ok {
		return item
	}

	// 1. 성경 본문 → 3줄 단위 페이징 (성경봉독은 타이틀 페이지 포함)
	if strings.HasPrefix(info, "b_") && contents != "" {
		pages := paginateText(contents, 3)
		if title == "성경봉독" {
			// Display와 동일하게 __title__ 페이지 추가
			titleSection := map[string]interface{}{"label": "제목", "startPage": 0, "text": ""}
			textSections := buildTextSections(pages)
			// startPage를 1씩 밀기
			for i := range textSections {
				if sp, ok := textSections[i]["startPage"].(int); ok {
					textSections[i]["startPage"] = sp + 1
				}
			}
			combined := []map[string]interface{}{titleSection}
			combined = append(combined, textSections...)
			item["sections"] = combined
		} else if len(pages) > 1 {
			item["sections"] = buildTextSections(pages)
		}
		return item
	}

	// 2. 신앙고백 → 10줄 단위 페이징 (주기도문은 짧아서 분할 불필요)
	if title == "신앙고백" && contents != "" {
		pages := paginateText(contents, 10)
		if len(pages) > 1 {
			item["sections"] = buildTextSections(pages)
		}
		return item
	}

	// 3. 찬송/헌금봉헌/성시교독 → 이미지 페이지 (성시교독은 표지 없음)
	if title == "찬송" || title == "헌금봉헌" || title == "성시교독" {
		var images []string
		switch v := item["images"].(type) {
		case []string:
			images = v
		case []interface{}:
			for _, img := range v {
				if s, ok := img.(string); ok {
					images = append(images, s)
				}
			}
		}
		if len(images) > 0 {
			obj, _ := item["obj"].(string)
			// lyricsMap이 있으면 각 이미지 페이지에 가사 미리보기 삽입
			var lyricsMap []string
			if lm, ok := item["lyricsMap"].([]string); ok {
				lyricsMap = lm
			}
			var sections []map[string]interface{}
			if title == "성시교독" {
				// 성시교독: 표지 없이 이미지만
				for i := range images {
					sections = append(sections, map[string]interface{}{
						"label":     fmt.Sprintf("%d", i+1),
						"startPage": i,
						"text":      "",
					})
				}
			} else {
				// 찬송/헌금봉헌: 표지 + 이미지
				// startPage는 expandHymnSubPages 고아-병합 로직과 동기화:
				//   lineCount <= 2: subSteps=1 / lineCount>=3: subSteps=lineCount/2 (정수 나눗셈)
				sections = []map[string]interface{}{
					{"label": "표지", "startPage": 0, "text": obj},
				}
				subPageOffset := 0
				for i := range images {
					startPage := subPageOffset + 1
					subSteps := 1
					if i < len(lyricsMap) && lyricsMap[i] != "" {
						lineCount := 0
						for _, line := range strings.Split(lyricsMap[i], "\n") {
							if strings.TrimSpace(line) != "" {
								lineCount++
							}
						}
						if lineCount >= 3 {
							subSteps = lineCount / 2 // JS orphan-merge와 동일 결과
						}
					}
					preview := ""
					if i < len(lyricsMap) && lyricsMap[i] != "" {
						preview = lyricsMap[i]
						runes := []rune(preview)
						if len(runes) > 60 {
							preview = string(runes[:60]) + "..."
						}
						preview = strings.ReplaceAll(preview, "\n", " ")
					}
					sections = append(sections, map[string]interface{}{
						"label":     fmt.Sprintf("%d", i+1),
						"startPage": startPage,
						"text":      preview,
					})
					subPageOffset += subSteps
				}
			}
			item["sections"] = sections
		}
		return item
	}

	return item
}

// paginateText — 빈 줄 제거 후 N줄 단위로 페이지 분할 (Display HTML의 paginate 함수와 동일)
func paginateText(text string, linesPerPage int) []string {
	allLines := strings.Split(text, "\n")
	var lines []string
	for _, l := range allLines {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return []string{text}
	}
	var pages []string
	for i := 0; i < len(lines); i += linesPerPage {
		end := i + linesPerPage
		if end > len(lines) {
			end = len(lines)
		}
		pages = append(pages, strings.Join(lines[i:end], "\n"))
	}
	return pages
}

// buildTextSections — 텍스트 페이지 배열 → sections 배열
func buildTextSections(pages []string) []map[string]interface{} {
	sections := make([]map[string]interface{}, len(pages))
	for i, page := range pages {
		// 미리보기: 첫 줄만 표시
		preview := page
		if idx := strings.Index(page, "\n"); idx > 0 {
			preview = page[:idx] + " ..."
		}
		sections[i] = map[string]interface{}{
			"label":     fmt.Sprintf("%d", i+1),
			"startPage": i,
			"text":      preview,
		}
	}
	return sections
}

// preprocessLyricsItem — 가사 곡 하나를 Display 항목으로 변환
func preprocessLyricsItem(song map[string]interface{}) map[string]interface{} {
	title, _ := song["title"].(string)
	lyrics, _ := song["lyrics"].(string)
	if lyrics == "" {
		// 이미 전처리된 항목(재전처리 시 "lyrics" 없이 "contents"만 존재) 대응 — idempotent 처리
		lyrics, _ = song["contents"].(string)
	}
	bpm := 0
	switch v := song["bpm"].(type) {
	case float64:
		bpm = int(v)
	case int:
		bpm = v
	}

	rawLines := strings.Split(strings.TrimSpace(lyrics), "\n")
	var lines []string
	for _, l := range rawLines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}

	// (xN) 표기는 독립 페이지가 아닌 이전 페이지에 붙임
	isRepeatMark := func(s string) bool {
		t := strings.TrimSpace(s)
		return len(t) >= 3 && t[0] == '(' && t[1] == 'x' && t[len(t)-1] == ')'
	}

	var allPages []string
	var sectionMarkers []map[string]interface{}
	for i := 0; i < len(lines); i += 2 {
		end := i + 2
		if end > len(lines) {
			end = len(lines)
		}
		pageText := strings.Join(lines[i:end], "\n")

		// 다음 줄이 (xN)이면 현재 페이지에 포함
		if end < len(lines) && isRepeatMark(lines[end]) {
			pageText += "\n" + lines[end]
			i++ // 다음 루프에서 (xN)줄 건너뜀
		}

		allPages = append(allPages, pageText)
		sectionMarkers = append(sectionMarkers, map[string]interface{}{
			"label":     fmt.Sprintf("%d", len(sectionMarkers)+1),
			"startPage": len(sectionMarkers),
			"text":      pageText,
		})
	}

	return map[string]interface{}{
		"title":    title,
		"info":     "lyrics_display",
		"obj":      "-",
		"contents": lyrics,
		"bpm":      bpm,
		"pages":    allPages,
		"sections": sectionMarkers,
	}
}

// formatChurchNews — 교회소식 children 재귀 순회하여 계층 텍스트 생성
func formatChurchNews(items []map[string]interface{}, depth int) string {
	var sb strings.Builder
	romanNumerals := []string{"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x"}
	for i, item := range items {
		title, _ := item["title"].(string)
		obj, _ := item["obj"].(string)
		if obj == "-" {
			obj = ""
		}
		tab := strings.Repeat("    ", depth-1)
		// depth별 인덱스 포맷
		var index string
		switch depth {
		case 1:
			index = fmt.Sprintf("%d.", i+1)
		case 2:
			index = fmt.Sprintf("%d)", i+1)
		case 3:
			if i < 26 {
				index = fmt.Sprintf("%c)", 'a'+i)
			}
		default:
			if i < len(romanNumerals) {
				index = fmt.Sprintf("%s)", romanNumerals[i])
			}
		}
		// children 유무에 따라 title 포맷
		hasChildren := false
		if rawChildren, ok := item["children"]; ok {
			childJSON, _ := json.Marshal(rawChildren)
			var ch []map[string]interface{}
			if json.Unmarshal(childJSON, &ch) == nil && len(ch) > 0 {
				hasChildren = true
			}
		}
		displayTitle := title
		if !hasChildren && title != "" {
			displayTitle = title + ":"
		}
		line := strings.TrimRight(fmt.Sprintf("%s%s %s %s", tab, index, displayTitle, obj), " ")
		sb.WriteString(line + "\n")
		// 재귀 처리
		if hasChildren {
			childJSON, _ := json.Marshal(item["children"])
			var ch []map[string]interface{}
			_ = json.Unmarshal(childJSON, &ch)
			sb.WriteString(formatChurchNews(ch, depth+1))
		}
	}
	return sb.String()
}

// fetchDisplayImages — 찬송/교독문 PDF를 Google Drive에서 다운로드 → PNG 변환 → URL 목록 반환
func fetchDisplayImages(title, obj string) []string {
	var category, splitNum string
	switch title {
	case "찬송", "헌금봉헌":
		category = "hymn"
		for _, r := range obj {
			if r >= '0' && r <= '9' {
				splitNum += string(r)
			}
		}
		if splitNum == "" {
			splitNum = obj
		}
	case "성시교독":
		category = "responsive_reading"
		splitNum = strings.Split(obj, ".")[0]
	default:
		return nil
	}

	execPath := path.ExecutePath("easyPreparation")
	cacheRoot := filepath.Join(execPath, "data", "cache")

	// 숫자 0-패딩
	num, err := strconv.Atoi(splitNum)
	var targetNum string
	if err == nil {
		targetNum = fmt.Sprintf("%03d.pdf", num)
	} else {
		targetNum = fmt.Sprintf("%s.pdf", splitNum)
	}
	base := strings.TrimSuffix(targetNum, ".pdf")

	// 캐시: data/cache/{category}/{base}/ 에 PNG 있으면 재사용
	imgDir := filepath.Join(cacheRoot, category, base)
	if cached := findCachedImages(category, base, imgDir); len(cached) > 0 {
		return cached
	}

	// PNG 다운로드 (Oracle Cloud hymn_pages/)
	BroadcastMessage("display_loading", map[string]interface{}{
		"message": fmt.Sprintf("%s/%s PNG 다운로드 중...", category, base),
	})
	if pngPaths := assets.DownloadPNGPages(category, targetNum, cacheRoot); len(pngPaths) > 0 {
		var urls []string
		for _, p := range pngPaths {
			rel, _ := filepath.Rel(cacheRoot, p)
			urls = append(urls, "/display/tmp/"+filepath.ToSlash(rel))
		}
		return urls
	}

	log.Printf("[display] PNG 없음 — %s/%s 건너뜀", category, base)
	return nil
}

// findCachedImages — data/cache/{category}/{base}/ 에서 PNG 파일 URL 목록 반환
func findCachedImages(category, base, imgDir string) []string {
	files, err := os.ReadDir(imgDir)
	if err != nil {
		return nil
	}
	type pageEntry struct {
		num int
		url string
	}
	var entries []pageEntry
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".png") {
			continue
		}
		name := strings.TrimSuffix(f.Name(), ".png")
		n, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		entries = append(entries, pageEntry{
			num: n,
			url: fmt.Sprintf("/display/tmp/%s/%s/%s", category, base, f.Name()),
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].num < entries[j].num })
	urls := make([]string, len(entries))
	for i, e := range entries {
		urls[i] = e.url
	}
	return urls
}

// ── 가사↔페이지 매핑 ──

// fetchHymnLyrics — hymns 테이블에서 가사 텍스트 조회 (PostgreSQL bibleDB)
func fetchHymnLyrics(number int) string {
	if bibleDB == nil {
		return ""
	}
	var lyrics string
	err := bibleDB.QueryRow("SELECT lyrics FROM hymns WHERE hymnbook='new' AND number=?", number).Scan(&lyrics)
	if err != nil {
		return ""
	}
	return lyrics
}

var versePattern = regexp.MustCompile(`(?m)^\d+\.\s`)

// splitVerses — 가사를 절 단위로 분리
func splitVerses(lyrics string) []string {
	lyrics = strings.TrimSpace(lyrics)
	if lyrics == "" {
		return nil
	}

	// 절 번호 패턴 (1. 2. ...) 으로 분리 시도
	locs := versePattern.FindAllStringIndex(lyrics, -1)
	if len(locs) >= 2 {
		var verses []string
		for i, loc := range locs {
			start := loc[0]
			var end int
			if i+1 < len(locs) {
				end = locs[i+1][0]
			} else {
				end = len(lyrics)
			}
			verse := strings.TrimSpace(lyrics[start:end])
			if verse != "" {
				verses = append(verses, verse)
			}
		}
		return verses
	}

	// 절 번호 없으면 빈 줄(\n\n)로 분리
	blocks := strings.Split(lyrics, "\n\n")
	var verses []string
	for _, b := range blocks {
		b = strings.TrimSpace(b)
		if b != "" {
			verses = append(verses, b)
		}
	}
	if len(verses) <= 1 {
		return []string{lyrics}
	}
	return verses
}

// splitIntoChunks — 텍스트를 N줄 단위로 분할
func splitIntoChunks(text string, linesPerChunk int) []string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var chunks []string
	for i := 0; i < len(lines); i += linesPerChunk {
		end := i + linesPerChunk
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, strings.Join(lines[i:end], "\n"))
	}
	if len(chunks) == 0 {
		return []string{text}
	}
	return chunks
}

// mapLyricsToPages — 전체 가사를 PNG 장수에 맞게 균등 분배 (lyricsMap.length == pageCount)
func mapLyricsToPages(verses []string, pageCount int) []string {
	if pageCount <= 0 || len(verses) == 0 {
		return nil
	}

	// 전체 가사를 줄 단위로 펼침
	var allLines []string
	for _, v := range verses {
		for _, line := range strings.Split(strings.TrimSpace(v), "\n") {
			if strings.TrimSpace(line) != "" {
				allLines = append(allLines, line)
			}
		}
	}

	n := len(allLines)
	if n == 0 {
		return nil
	}

	// 각 페이지(PNG)에 연속된 줄을 균등 배분 (누락 없이, lyricsMap[i] ↔ images[i])
	result := make([]string, pageCount)
	for i := 0; i < pageCount; i++ {
		start := i * n / pageCount
		end := (i + 1) * n / pageCount
		if start >= n {
			start = n - 1
		}
		if end > n {
			end = n
		}
		result[i] = strings.Join(allLines[start:end], "\n")
	}
	return result
}

// fetchBibleText — "책명_코드/장:절" 형식에서 성경 본문 조회 (기본 버전)
func fetchBibleText(obj string) (string, string) {
	return fetchBibleTextWithVersion(obj, 1)
}

// fetchBibleTextWithVersion — 지정된 버전으로 성경 본문 조회
func fetchBibleTextWithVersion(obj string, versionID int) (string, string) {
	var texts []string
	var refs []string

	for _, part := range strings.Split(obj, ",") {
		part = strings.TrimSpace(part)
		underIdx := strings.Index(part, "_")
		if underIdx < 0 {
			continue
		}
		korName := part[:underIdx]
		codeAndRange := part[underIdx+1:]

		slashIdx := strings.Index(codeAndRange, "/")
		if slashIdx < 0 {
			continue
		}
		verseRange := codeAndRange[slashIdx+1:]

		text, err := quote.GetQuoteWithVersion(codeAndRange, versionID)
		if err != nil {
			continue
		}
		texts = append(texts, text)
		refs = append(refs, fmt.Sprintf("%s %s", korName, verseRange))
	}

	return strings.Join(texts, "\n"), strings.Join(refs, ", ")
}
