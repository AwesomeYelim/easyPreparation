package handlers

import (
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/thumbnail"
	"easyPreparation_1.0/internal/youtube"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
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
	// 동일 스케줄 중복 실행 방지: "worshipType_YYYYMMDD" → true
	lastExecuted   = map[string]bool{}
	lastExecutedMu sync.Mutex
	schedulerStop  chan struct{}
	// 수동 방송 시작 중복 방지
	streamStartMu sync.Mutex
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
	data, err := os.ReadFile(schedulePath())
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
	data, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		log.Printf("[scheduler] 설정 저장 실패 (marshal): %v", err)
		return
	}
	if err := os.WriteFile(schedulePath(), data, 0644); err != nil {
		log.Printf("[scheduler] 설정 저장 실패 (write): %v", err)
	}
}

// ── 스케줄러 초기화/루프 ──

func InitScheduler() {
	scheduleMu.Lock()
	scheduleConf = loadScheduleConfig()
	scheduleMu.Unlock()
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

func checkSchedule(now time.Time) {
	scheduleMu.RLock()
	conf := scheduleConf
	scheduleMu.RUnlock()

	weekday := int(now.Weekday())
	for _, entry := range conf.Entries {
		if !entry.Enabled || entry.Weekday != weekday {
			continue
		}
		// 예배 시작 시각
		target := time.Date(now.Year(), now.Month(), now.Day(), entry.Hour, entry.Minute, 0, 0, now.Location())
		diff := target.Sub(now)
		countdownDur := time.Duration(conf.CountdownMinutes) * time.Minute

		if diff > 0 && diff <= countdownDur {
			// 카운트다운 구간
			remaining := int(diff.Seconds())
			minutes := remaining / 60
			seconds := remaining % 60
			BroadcastMessage("schedule_countdown", map[string]interface{}{
				"worshipType": entry.WorshipType,
				"label":       entry.Label,
				"remaining":   remaining,
				"minutes":     minutes,
				"seconds":     seconds,
			})
		}

		// T-0 실행 (±1초 허용)
		if diff >= -1*time.Second && diff <= 1*time.Second {
			execKey := fmt.Sprintf("%s_%s", entry.WorshipType, now.Format("20060102"))
			lastExecutedMu.Lock()
			if lastExecuted[execKey] {
				lastExecutedMu.Unlock()
				continue
			}
			lastExecuted[execKey] = true
			lastExecutedMu.Unlock()

			go executeSchedule(entry, conf.AutoStream)
		}
	}
}

func executeSchedule(entry ScheduleEntry, autoStream bool) {
	log.Printf("[scheduler] 예배 시작: %s (%s)", entry.Label, entry.WorshipType)

	// 스케줄러가 순서를 교체하기 전에 현재 상태를 별도 백업
	execPath := path.ExecutePath("easyPreparation")
	backupPath := filepath.Join(execPath, "data", "display_state_before_schedule.json")
	if src, err := os.ReadFile(displayStatePath()); err == nil {
		os.WriteFile(backupPath, src, 0644)
		log.Printf("[scheduler] 기존 display_state 백업: %s", backupPath)
	}

	// config/{worshipType}.json 로드
	configFile := filepath.Join(execPath, "config", entry.WorshipType+".json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("[scheduler] config 파일 없음: %s — 스킵", configFile)
		BroadcastMessage("schedule_started", map[string]interface{}{
			"worshipType": entry.WorshipType,
			"label":       entry.Label,
			"error":       "config 파일 없음",
		})
		return
	}

	var order []map[string]interface{}
	if err := json.Unmarshal(data, &order); err != nil {
		log.Printf("[scheduler] config 파싱 실패: %v — 스킵", err)
		return
	}

	// 전처리
	for i, item := range order {
		order[i] = preprocessItem(item)
	}

	// 타이머 초기화 + 순서 교체
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

	// 자동 스트리밍: YouTube 방송 생성 → 썸네일 업로드 → OBS 스트림 설정 → 송출 시작
	if autoStream {
		ytM := youtube.Get()
		if ytM.IsEnabled() {
			// 썸네일 설정에서 제목·설명 가져오기
			cfg, _ := thumbnail.LoadConfig()
			title := entry.Label
			description := ""
			if cfg != nil {
				sermonTitle, scripture := sermonDataForWorship(entry.WorshipType)
				_, t := cfg.ResolveTheme(entry.WorshipType, time.Now(), sermonTitle, scripture)
				title = t
				description = cfg.ResolveDescription(entry.WorshipType, time.Now(), sermonTitle, scripture)
			}

			// YouTube 방송 생성 + 스트림 바인딩
			server, key, broadcastID, err := youtube.CreateBroadcastAndBind(title, description)
			if err != nil {
				log.Printf("[scheduler] YouTube 방송 생성 실패: %v — 기존 방식으로 스트리밍", err)
				// YouTube 실패해도 기존 OBS 스트리밍은 시도
				if err := obs.Get().StartStreaming(); err != nil {
					log.Printf("[scheduler] OBS 스트리밍 시작 실패: %v", err)
				}
			} else {
				// 썸네일 생성 + 업로드 (upcoming 상태에서 → 확실히 반영)
				GenerateAndUploadThumbnailTo(entry.WorshipType, broadcastID)

				// OBS 스트리밍 중이면 먼저 중지
				obsM := obs.Get()
				streamStatus := obsM.GetStreamStatus()
				if streamStatus.Active {
					obsM.StopStreaming()
					for i := 0; i < 10; i++ {
						s := obsM.GetStreamStatus()
						if !s.Active {
							break
						}
						time.Sleep(500 * time.Millisecond)
					}
				}

				// OBS 스트림 설정
				if err := obsM.SetStreamSettingsWithBroadcastID(server, key, broadcastID); err != nil {
					log.Printf("[scheduler] OBS 스트림 설정 실패: %v", err)
				}

				// OBS 송출 시작
				if err := obsM.StartStreaming(); err != nil {
					log.Printf("[scheduler] OBS 스트리밍 시작 실패: %v", err)
				}

				// OBS 실제 스트리밍 확인 후 YouTube live 전환
				go func(bid string) {
					if waitForOBSStreaming(obsM, 30) {
						if err := youtube.TransitionToLive(bid); err != nil {
							log.Printf("[scheduler] YouTube live 전환 실패: %v", err)
						}
					} else {
						log.Println("[scheduler] OBS 스트리밍 미확인 — YouTube live 전환 스킵")
					}
				}(broadcastID)
			}
		} else {
			// YouTube 미연결 → 기존 OBS 스트리밍만
			GenerateAndUploadThumbnail(entry.WorshipType)
			if err := obs.Get().StartStreaming(); err != nil {
				log.Printf("[scheduler] OBS 스트리밍 시작 실패: %v", err)
			}
		}
	} else {
		// autoStream 꺼져 있으면 썸네일만 생성
		go GenerateAndUploadThumbnail(entry.WorshipType)
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
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		scheduleMu.Lock()
		scheduleConf = conf
		scheduleMu.Unlock()
		saveScheduleConfig(conf)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Action      string `json:"action"`
		WorshipType string `json:"worshipType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
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
		http.Error(w, "Unknown worshipType", http.StatusBadRequest)
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
		go executeSchedule(*entry, false) // autoStream=false (테스트이므로)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": "스케줄 즉시 실행"})

	default:
		http.Error(w, "Unknown action (countdown|trigger)", http.StatusBadRequest)
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
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Action           string `json:"action"`
		IncludeThumbnail *bool  `json:"includeThumbnail"`
		IncludeTitle     *bool  `json:"includeTitle"`
		IsTest           bool   `json:"isTest"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
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
							sermonTitle, scripture := sermonDataForWorship(worshipType)
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
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}
