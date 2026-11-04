package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"easyPreparation_1.0/internal/httpx"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/safefile"
	"easyPreparation_1.0/internal/thumbnail"
	"easyPreparation_1.0/internal/youtube"
)

// ── 스케줄러 데이터 구조 ──

type ScheduleEntry struct {
	WorshipType string `json:"worshipType"`
	Label       string `json:"label"`
	Weekday     int    `json:"weekday"` // 0=일, 1=월, ..., 6=토
	Hour        int    `json:"hour"`
	Minute      int    `json:"minute"`
	Enabled     bool   `json:"enabled"`
}

type ScheduleConfig struct {
	Entries          []ScheduleEntry `json:"entries"`
	AutoStream       bool            `json:"autoStream"`
	CountdownMinutes int             `json:"countdownMinutes"`
}

var (
	scheduleMu   sync.RWMutex
	scheduleConf ScheduleConfig
	// 같은 단계 중복 실행 방지: "worshipType_YYYYMMDD_prep|start" → 실행 시각.
	// 파일(data/schedule_runs.json)에도 남겨, 앱을 다시 켜도 같은 날 두 번 실행하지 않는다.
	scheduleRuns   = map[string]string{}
	scheduleRunsMu sync.Mutex
	schedulerStop  chan struct{}
	// 수동 방송 시작 중복 방지
	streamStartMu sync.Mutex
)

const (
	// 예배 시각을 놓쳤을 때(절전·앱 늦게 켬) 이 시간 안이면 늦게라도 실행한다
	scheduleGrace = 5 * time.Minute
	// 이 시간 안에 운영자가 순서를 보내거나 넘겼으면 진행 중인 예배로 보고 순서를 바꾸지 않는다
	operatorActiveWindow = 60 * time.Minute
	// 자동 송출 시작 후 OBS 가 실제로 송출을 시작할 때까지 기다리는 초 (OBS 확인 창 대비)
	obsStartWait = 120
)

func schedulePath() string {
	execPath := path.ExecutePath("easyPreparation")
	return filepath.Join(execPath, "data", "schedule.json")
}

func defaultSchedule() ScheduleConfig {
	return ScheduleConfig{
		Entries: []ScheduleEntry{
			{WorshipType: "main_worship", Label: "주일예배", Weekday: 0, Hour: 11, Minute: 0, Enabled: false},
			{WorshipType: "after_worship", Label: "오후예배", Weekday: 0, Hour: 14, Minute: 0, Enabled: false},
			{WorshipType: "wed_worship", Label: "수요예배", Weekday: 3, Hour: 19, Minute: 30, Enabled: false},
			{WorshipType: "fri_worship", Label: "금요예배", Weekday: 5, Hour: 20, Minute: 30, Enabled: false},
		},
		AutoStream:       true,
		CountdownMinutes: 5,
	}
}

func loadScheduleConfig() ScheduleConfig {
	data, err := safefile.ReadJSONWithRecovery(schedulePath())
	if err != nil {
		log.Printf("[scheduler] 설정 파일 없음 — 기본값 생성")
		conf := defaultSchedule()
		saveScheduleConfig(conf)
		return conf
	}
	var conf ScheduleConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		log.Printf("[scheduler] 설정 파싱 실패: %v — 기본값 사용", err)
		return defaultSchedule()
	}
	return conf
}

// resolveCurrentWorshipType — 오늘 요일의 스케줄 항목 중 현재 시각과 가장 가까운 예배 종류 반환
// (수동 "방송 시작" 버튼이 어떤 예배인지 하드코딩 없이 판단할 수 있도록)
// entry.Enabled(자동 시작 on/off)는 무관 — 요일/시각 데이터는 자동 스케줄러가 꺼져 있어도 유효한 예배 일정
func resolveCurrentWorshipType() string {
	scheduleMu.RLock()
	conf := scheduleConf
	scheduleMu.RUnlock()

	now := time.Now()
	weekday := int(now.Weekday())

	var best string
	var bestDiff time.Duration
	for _, entry := range conf.Entries {
		if entry.Weekday != weekday {
			continue
		}
		target := time.Date(now.Year(), now.Month(), now.Day(), entry.Hour, entry.Minute, 0, 0, now.Location())
		diff := target.Sub(now)
		if diff < 0 {
			diff = -diff
		}
		if best == "" || diff < bestDiff {
			best = entry.WorshipType
			bestDiff = diff
		}
	}
	if best == "" {
		return "main_worship"
	}
	return best
}

func saveScheduleConfig(conf ScheduleConfig) {
	if err := safefile.WriteJSON(schedulePath(), conf); err != nil {
		log.Printf("[scheduler] 설정 저장 실패: %v", err)
	}
}

// ── 스케줄러 초기화/루프 ──

func InitScheduler() {
	scheduleMu.Lock()
	scheduleConf = loadScheduleConfig()
	scheduleMu.Unlock()
	loadScheduleRuns()
	// 기존 goroutine 정리 (double-init 방지)
	if schedulerStop != nil {
		close(schedulerStop)
	}
	schedulerStop = make(chan struct{})
	go schedulerLoop()
	log.Println("[scheduler] 스케줄러 시작")
}

func StopScheduler() {
	if schedulerStop != nil {
		close(schedulerStop)
	}
}

func schedulerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-schedulerStop:
			return
		case t := <-ticker.C:
			checkSchedule(t)
		}
	}
}

func scheduleRunsPath() string {
	execPath := path.ExecutePath("easyPreparation")
	return filepath.Join(execPath, "data", "schedule_runs.json")
}

func loadScheduleRuns() {
	runs := map[string]string{}
	if data, err := os.ReadFile(scheduleRunsPath()); err == nil {
		_ = json.Unmarshal(data, &runs)
	}
	scheduleRunsMu.Lock()
	scheduleRuns = runs
	scheduleRunsMu.Unlock()
}

// claimScheduleRun — 오늘 이 단계를 아직 안 했으면 실행했다고 기록하고 true.
// 오늘 날짜가 아닌 기록은 지운다.
func claimScheduleRun(key, today string) bool {
	scheduleRunsMu.Lock()
	defer scheduleRunsMu.Unlock()
	if _, done := scheduleRuns[key]; done {
		return false
	}
	for k := range scheduleRuns {
		if !strings.Contains(k, "_"+today+"_") {
			delete(scheduleRuns, k)
		}
	}
	scheduleRuns[key] = time.Now().Format(time.RFC3339)
	if err := safefile.WriteJSON(scheduleRunsPath(), scheduleRuns); err != nil {
		log.Printf("[scheduler] 실행 기록 저장 실패: %v", err)
	}
	return true
}

// scheduleStep — 한 예배 항목에 대해 지금 해야 할 일
type scheduleStep struct {
	countdown int  // >0 이면 카운트다운 남은 초 (방송할 것)
	prep      bool // 송출 준비 구간 안 (카운트다운 시작 ~ 예배 시각 + 유예)
	start     bool // 순서 시작 구간 안 (예배 시각 ~ 예배 시각 + 유예)
}

// dueScheduleStep — 시각만으로 판단하는 순수 함수 (중복 실행 방지는 호출하는 쪽에서)
//   prep : 카운트다운 시작 시각(예배 시각 - countdownMinutes)부터. 자동 송출이면 YouTube 방송 준비 + OBS 송출 시작
//          (시청자도 Display 의 카운트다운을 보게 된다)
//   start: 예배 시각부터. 순서 로드 + 첫 항목 장면 전환
// 두 단계 모두 시각을 놓쳤어도(절전·앱 늦게 켬) 예배 시각 + scheduleGrace 안이면 늦게라도 실행한다.
func dueScheduleStep(entry ScheduleEntry, now time.Time, countdownMinutes int) scheduleStep {
	var st scheduleStep
	if !entry.Enabled || entry.Weekday != int(now.Weekday()) {
		return st
	}
	countdownDur := time.Duration(countdownMinutes) * time.Minute
	if countdownDur < 0 {
		countdownDur = 0
	}
	target := time.Date(now.Year(), now.Month(), now.Day(), entry.Hour, entry.Minute, 0, 0, now.Location())
	diff := target.Sub(now)

	if diff > 0 && diff <= countdownDur {
		st.countdown = int(diff.Seconds())
		if st.countdown < 1 {
			st.countdown = 1
		}
	}
	if diff > countdownDur || diff < -scheduleGrace {
		return st
	}
	st.prep = true
	st.start = diff <= 0
	return st
}

// checkSchedule — 1초마다 호출. 예배마다 prep/start 단계를 하루 한 번씩 실행한다.
func checkSchedule(now time.Time) {
	scheduleMu.RLock()
	conf := scheduleConf
	scheduleMu.RUnlock()

	today := now.Format("20060102")
	for _, entry := range conf.Entries {
		st := dueScheduleStep(entry, now, conf.CountdownMinutes)
		if st.countdown > 0 {
			BroadcastMessage("schedule_countdown", map[string]interface{}{
				"worshipType": entry.WorshipType,
				"label":       entry.Label,
				"remaining":   st.countdown,
				"minutes":     st.countdown / 60,
				"seconds":     st.countdown % 60,
			})
		}
		base := entry.WorshipType + "_" + today + "_"
		if st.prep && claimScheduleRun(base+"prep", today) {
			// 순서에 쓰는 악보·성경 본문을 미리 받아 둔다 — 예배 시각에 순서가 바로 뜨도록
			go warmScheduledOrder(entry)
			if conf.AutoStream {
				go prepareScheduledStream(entry)
			}
		}
		if st.start && claimScheduleRun(base+"start", today) {
			go startScheduledWorship(entry)
		}
	}
}

// scheduleNotice — 자동 시작이 무언가를 건너뛰거나 실패했을 때 운영자 화면에 알린다
func scheduleNotice(entry ScheduleEntry, message string) {
	log.Printf("[scheduler] %s: %s", entry.Label, message)
	BroadcastMessage("schedule_notice", map[string]interface{}{
		"worshipType": entry.WorshipType,
		"label":       entry.Label,
		"message":     message,
	})
}

// prepareScheduledStream — 카운트다운 시작 때 송출을 먼저 시작한다.
// 이미 송출 중이면(운영자가 먼저 켬) 그 방송을 건드리지 않는다.
func prepareScheduledStream(entry ScheduleEntry) {
	obsM := obs.Get()
	if obsM.GetStreamStatus().Active {
		scheduleNotice(entry, "이미 송출 중이라 자동 송출은 건너뜁니다")
		return
	}
	if !streamStartMu.TryLock() {
		scheduleNotice(entry, "방송 시작이 진행 중이라 자동 송출은 건너뜁니다")
		return
	}
	defer streamStartMu.Unlock()
	if !obsM.GetStatus().Connected {
		// 미연결이면 StartStreaming 이 조용히 아무것도 안 하므로 여기서 알린다
		scheduleNotice(entry, "OBS 가 연결되어 있지 않아 자동 송출을 하지 못했습니다")
		return
	}
	log.Printf("[scheduler] 송출 준비: %s (%s)", entry.Label, entry.WorshipType)

	ytM := youtube.Get()
	if !ytM.IsEnabled() {
		// YouTube 미연결 → OBS 에 설정된 송출 대상으로 시작
		go GenerateAndUploadThumbnail(entry.WorshipType)
		if err := obsM.StartStreaming(); err != nil {
			scheduleNotice(entry, "OBS 송출 시작 실패: "+err.Error())
			return
		}
		if !waitForOBSStreaming(obsM, obsStartWait) {
			scheduleNotice(entry, "OBS 가 송출을 시작하지 않았습니다. OBS 송출 설정과 확인 창을 보세요")
		}
		return
	}

	title := entry.Label
	description := ""
	if cfg, _ := thumbnail.LoadConfig(); cfg != nil {
		sermonTitle, scripture := sermonDataPreferLive(entry.WorshipType)
		_, title = cfg.ResolveTheme(entry.WorshipType, time.Now(), sermonTitle, scripture)
		description = cfg.ResolveDescription(entry.WorshipType, time.Now(), sermonTitle, scripture)
	}

	server, key, broadcastID, err := prepareYouTubeBroadcast(entry.WorshipType, title, description, true)
	if err != nil {
		log.Printf("[scheduler] YouTube 방송 생성 실패: %v", err)
		scheduleNotice(entry, "YouTube 방송 만들기 실패, OBS 설정대로 송출합니다")
		if err := obsM.StartStreaming(); err != nil {
			scheduleNotice(entry, "OBS 송출 시작 실패: "+err.Error())
		}
		return
	}
	if err := obsM.SetStreamSettingsWithBroadcastID(server, key, broadcastID); err != nil {
		log.Printf("[scheduler] OBS 스트림 설정 실패: %v", err)
	}
	if err := obsM.StartStreaming(); err != nil {
		scheduleNotice(entry, "OBS 송출 시작 실패: "+err.Error())
		return
	}
	if !waitForOBSStreaming(obsM, obsStartWait) {
		scheduleNotice(entry, "OBS 가 송출을 시작하지 않았습니다. OBS 확인 창이 떠 있는지 보세요")
		return
	}
	if err := youtube.TransitionToLive(broadcastID); err != nil {
		log.Printf("[scheduler] YouTube live 전환 실패: %v", err)
	}
}

// startScheduledWorship — 예배 시각에 순서를 로드하고 첫 항목 장면으로 바꾼다.
// 운영자가 최근에 순서를 보내거나 넘겼으면 진행 중인 예배로 보고 그대로 둔다.
func startScheduledWorship(entry ScheduleEntry) {
	if operatorActiveWithin(operatorActiveWindow) {
		scheduleNotice(entry, "이미 진행 중이라 순서를 바꾸지 않았습니다")
		BroadcastMessage("schedule_started", map[string]interface{}{
			"worshipType": entry.WorshipType,
			"label":       entry.Label,
			"skipped":     true,
		})
		return
	}
	executeSchedule(entry)
}

// loadScheduledOrder — config/{worshipType}.json 을 읽어 전처리(성경 본문·악보 받기 등)까지 한 순서
func loadScheduledOrder(entry ScheduleEntry) ([]map[string]interface{}, error) {
	configFile := filepath.Join(path.ExecutePath("easyPreparation"), "config", entry.WorshipType+".json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("예배 순서 파일이 없어 순서를 불러오지 못했습니다")
	}
	var order []map[string]interface{}
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("예배 순서 파일을 읽지 못했습니다")
	}
	for i, item := range order {
		order[i] = preprocessItem(item)
	}
	return order, nil
}

// warmScheduledOrder — 카운트다운 시작 때 순서를 한 번 전처리해 악보 이미지 등을 캐시에 받아 둔다
func warmScheduledOrder(entry ScheduleEntry) {
	start := time.Now()
	if _, err := loadScheduledOrder(entry); err != nil {
		scheduleNotice(entry, err.Error())
		return
	}
	log.Printf("[scheduler] 순서 미리 준비 완료: %s (%s)", entry.Label, time.Since(start).Round(time.Millisecond))
}

func executeSchedule(entry ScheduleEntry) {
	log.Printf("[scheduler] 예배 시작: %s (%s)", entry.Label, entry.WorshipType)

	// 스케줄러가 순서를 교체하기 전에 현재 상태를 별도 백업
	execPath := path.ExecutePath("easyPreparation")
	backupPath := filepath.Join(execPath, "data", "display_state_before_schedule.json")
	if src, err := os.ReadFile(displayStatePath()); err == nil {
		os.WriteFile(backupPath, src, 0644)
		log.Printf("[scheduler] 기존 display_state 백업: %s", backupPath)
	}

	order, err := loadScheduledOrder(entry)
	if err != nil {
		scheduleNotice(entry, err.Error())
		BroadcastMessage("schedule_started", map[string]interface{}{
			"worshipType": entry.WorshipType,
			"label":       entry.Label,
			"error":       err.Error(),
		})
		return
	}

	// 타이머 초기화 + 순서 교체 (스케줄러의 교체는 운영자 활동으로 치지 않는다)
	stopServerTimer()
	orderMu.Lock()
	currentOrder = order
	currentIdx = 0
	orderMu.Unlock()

	// WS broadcast
	BroadcastMessage("order", map[string]interface{}{
		"items":      order,
		"churchName": displayChurchName,
	})
	go saveDisplayState()

	// OBS 씬 전환
	if len(order) > 0 {
		if title, ok := order[0]["title"].(string); ok {
			go obs.Get().SwitchScene(title)
		}
	}

	BroadcastMessage("schedule_started", map[string]interface{}{
		"worshipType": entry.WorshipType,
		"label":       entry.Label,
	})

	log.Printf("[scheduler] 예배 시작 완료: %s", entry.Label)
}

// ── HTTP 핸들러 ──

// ScheduleHandler — GET: 스케줄 조회, POST: 스케줄 수정
func ScheduleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		scheduleMu.RLock()
		conf := scheduleConf
		scheduleMu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(conf)

	case http.MethodPost:
		var conf ScheduleConfig
		if err := json.NewDecoder(r.Body).Decode(&conf); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		scheduleMu.Lock()
		scheduleConf = conf
		scheduleMu.Unlock()
		saveScheduleConfig(conf)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})

	default:
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// ScheduleTestHandler — POST: 스케줄러 테스트
// {action: "countdown", worshipType: "main_worship"} — 10초 카운트다운 테스트
// {action: "trigger", worshipType: "main_worship"}   — 즉시 실행 테스트
func ScheduleTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var body struct {
		Action      string `json:"action"`
		WorshipType string `json:"worshipType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 스케줄에서 해당 타입 찾기
	scheduleMu.RLock()
	conf := scheduleConf
	scheduleMu.RUnlock()

	var entry *ScheduleEntry
	for _, e := range conf.Entries {
		if e.WorshipType == body.WorshipType {
			eCopy := e
			entry = &eCopy
			break
		}
	}
	if entry == nil {
		httpx.Error(w, http.StatusBadRequest, "Unknown worshipType")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch body.Action {
	case "countdown":
		// 10초 카운트다운 테스트 (백그라운드)
		go func() {
			for i := 10; i > 0; i-- {
				BroadcastMessage("schedule_countdown", map[string]interface{}{
					"worshipType": entry.WorshipType,
					"label":       entry.Label,
					"remaining":   i,
					"minutes":     i / 60,
					"seconds":     i % 60,
				})
				time.Sleep(1 * time.Second)
			}
			BroadcastMessage("schedule_started", map[string]interface{}{
				"worshipType": entry.WorshipType,
				"label":       entry.Label,
			})
		}()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": "10초 카운트다운 시작"})

	case "trigger":
		// 즉시 실행 (config 파일 로드 + 전처리 + broadcast)
		go executeSchedule(*entry) // 순서 로드만 (송출은 하지 않음)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": "스케줄 즉시 실행"})

	default:
		httpx.Error(w, http.StatusBadRequest, "Unknown action (countdown|trigger)")
	}
}

// waitForOBSStreaming — OBS 스트리밍이 실제로 시작될 때까지 대기 (최대 maxSec초)
// OBS 30+에서 YouTube RTMP URL 감지 시 팝업이 뜨며 스트리밍 시작이 지연될 수 있음.
// 사용자가 팝업에서 "닫기"를 클릭하면 스트리밍이 시작됨.
func waitForOBSStreaming(obsM interface{ GetStreamStatus() obs.StreamStatus }, maxSec int) bool {
	for i := 0; i < maxSec; i++ {
		time.Sleep(1 * time.Second)
		if obsM.GetStreamStatus().Active {
			log.Printf("[stream] OBS 스트리밍 실제 시작 확인 (%d초 경과)", i+1)
			return true
		}
		if i%5 == 4 {
			log.Printf("[stream] OBS 스트리밍 대기 중... (%d/%d초, 팝업이 뜬 경우 닫기 클릭)", i+1, maxSec)
		}
	}
	log.Printf("[stream] OBS 스트리밍 %d초 내 미시작", maxSec)
	return false
}

// StreamControlHandler — POST: 스트리밍 수동 제어 {action: "start"|"stop"|"status"}
func StreamControlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var body struct {
		Action           string `json:"action"`
		IncludeThumbnail *bool  `json:"includeThumbnail"`
		IncludeTitle     *bool  `json:"includeTitle"`
		IsTest           bool   `json:"isTest"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	includeThumbnail := true
	if body.IncludeThumbnail != nil {
		includeThumbnail = *body.IncludeThumbnail
	}
	includeTitle := true
	if body.IncludeTitle != nil {
		includeTitle = *body.IncludeTitle
	}

	w.Header().Set("Content-Type", "application/json")

	switch body.Action {
	case "start":
		if !streamStartMu.TryLock() {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "이미 방송 시작 중"})
			return
		}
		log.Println("[stream] 방송 시작 요청")

		// 응답 먼저 보내고 백그라운드에서 처리
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
		go func() {
			obsM := obs.Get()
			ytM := youtube.Get()

			if ytM.IsEnabled() {
				var title, worshipType, description string
				if body.IsTest {
					// 테스트 방송 — 실제 예배 테마/썸네일과 분리, 제목으로 바로 구분 가능하게
					title = "TEST " + time.Now().Format("2006-01-02 15:04")
				} else {
					// 예배 종류 판별 (오늘 요일의 스케줄 항목 기준 — 하드코딩 금지)
					worshipType = resolveCurrentWorshipType()
					title = "라이브 예배"
					// 제목 자동 생성 (사용자가 끈 경우 기본 제목 유지)
					if includeTitle {
						cfg, err := thumbnail.LoadConfig()
						if err == nil {
							sermonTitle, scripture := sermonDataPreferLive(worshipType)
							_, t := cfg.ResolveTheme(worshipType, time.Now(), sermonTitle, scripture)
							title = t
							description = cfg.ResolveDescription(worshipType, time.Now(), sermonTitle, scripture)
						}
					}
				}

				// YouTube 방송 생성 + 스트림 키 바인딩
				server, key, broadcastID, err := youtube.CreateBroadcastAndBind(title, description)
				if err != nil {
					log.Printf("[stream] YouTube 방송 생성 실패: %v — OBS만 시작", err)
					obsM.StartStreaming()
					streamStartMu.Unlock()
					return
				}
				log.Printf("[stream] YouTube 방송 준비 완료: %s", broadcastID)

				// 썸네일 (테스트 방송이거나 사용자가 끈 경우 스킵)
				if includeThumbnail && !body.IsTest {
					GenerateAndUploadThumbnailTo(worshipType, broadcastID)
				}

				// OBS 스트리밍 중이면 중지
				if obsM.GetStreamStatus().Active {
					log.Println("[stream] 기존 OBS 스트리밍 중지...")
					obsM.StopStreaming()
					for i := 0; i < 10; i++ {
						if !obsM.GetStreamStatus().Active {
							break
						}
						time.Sleep(500 * time.Millisecond)
					}
				}

				// OBS 스트림 설정
				if err := obsM.SetStreamSettingsWithBroadcastID(server, key, broadcastID); err != nil {
					log.Printf("[stream] OBS 스트림 설정 실패: %v", err)
				}
				time.Sleep(2 * time.Second)

				// OBS 스트리밍 시작
				if err := obsM.StartStreaming(); err != nil {
					log.Printf("[stream] OBS 스트리밍 시작 실패: %v", err)
					streamStartMu.Unlock()
					return
				}
				log.Println("[stream] OBS StartStream 명령 전송 완료 (팝업 뜰 경우 닫기 클릭 필요)")
				streamStartMu.Unlock()

				// 실제 스트리밍 시작 대기 후 YouTube live 전환
				bid := broadcastID
				go func() {
					if waitForOBSStreaming(obsM, 30) {
						youtube.TransitionToLive(bid)
					} else {
						log.Println("[stream] OBS 스트리밍 미확인 — YouTube live 전환 스킵")
					}
				}()
			} else {
				obsM.StartStreaming()
				streamStartMu.Unlock()
			}
		}()

	case "stop":
		err := obs.Get().StopStreaming()
		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})

	case "status":
		status := obs.Get().GetStreamStatus()
		_ = json.NewEncoder(w).Encode(status)

	default:
		httpx.Error(w, http.StatusBadRequest, "Unknown action")
	}
}
