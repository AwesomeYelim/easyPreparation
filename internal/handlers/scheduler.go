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
			{WorshipType: "main_worship", Label: "주일예배", Weekday: 0, Hour: 11, Minute: 0, Enabled: true},
			{WorshipType: "after_worship", Label: "오후예배", Weekday: 0, Hour: 14, Minute: 0, Enabled: true},
			{WorshipType: "wed_worship", Label: "수요예배", Weekday: 3, Hour: 19, Minute: 30, Enabled: true},
			{WorshipType: "fri_worship", Label: "금요예배", Weekday: 5, Hour: 20, Minute: 30, Enabled: true},
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

	// config/{worshipType}.json 로드
	execPath := path.ExecutePath("easyPreparation")
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
			// 썸네일 설정에서 제목 가져오기
			cfg, _ := thumbnail.LoadConfig()
			title := entry.Label
			if cfg != nil {
				_, t := cfg.ResolveTheme(entry.WorshipType, time.Now())
				title = t
			}

			// YouTube 방송 생성 + 스트림 바인딩
			server, key, broadcastID, err := youtube.CreateBroadcastAndBind(title)
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

				// OBS 스트림 설정 (커스텀 RTMP)
				if err := obsM.SetStreamSettings(server, key); err != nil {
					log.Printf("[scheduler] OBS 스트림 설정 실패: %v", err)
				}

				// OBS 송출 시작
				if err := obsM.StartStreaming(); err != nil {
					log.Printf("[scheduler] OBS 스트리밍 시작 실패: %v", err)
				}

				// EnableAutoStart 미작동 대비 — 수동 live 전환
				go func(bid string) {
					if err := youtube.TransitionToLive(bid); err != nil {
						log.Printf("[scheduler] YouTube live 전환 실패: %v", err)
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
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
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
			defer streamStartMu.Unlock()
			obsM := obs.Get()
			ytM := youtube.Get()

			if ytM.IsEnabled() {
				// 0. 기존 방송 정리
				log.Println("[stream] 기존 방송 정리 중...")
				youtube.CleanupBroadcasts()

				// 제목 결정
				title := "라이브 예배"
				cfg, err := thumbnail.LoadConfig()
				if err == nil {
					_, t := cfg.ResolveTheme("main_worship", time.Now())
					title = t
				}

				// 1. YouTube 방송 생성
				server, key, broadcastID, err := youtube.CreateBroadcastAndBind(title)
				if err != nil {
					log.Printf("[stream] YouTube 방송 생성 실패: %v — OBS만 시작", err)
					obsM.StartStreaming()
					return
				}
				log.Printf("[stream] YouTube 방송 준비 완료: %s", broadcastID)

				// 2. 썸네일 (동기 — upcoming 상태에서 업로드해야 반영됨)
				GenerateAndUploadThumbnailTo("main_worship", broadcastID)

				// 3. OBS 스트리밍 중이면 중지
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

				// 4. OBS 스트림 설정
				if err := obsM.SetStreamSettings(server, key); err != nil {
					log.Printf("[stream] OBS 스트림 설정 실패: %v", err)
				}
				time.Sleep(2 * time.Second)

				// 5. OBS 스트리밍 시작
				if err := obsM.StartStreaming(); err != nil {
					log.Printf("[stream] OBS 스트리밍 시작 실패: %v", err)
					return
				}
				log.Println("[stream] OBS 스트리밍 시작 완료")

				// 6. YouTube live 전환
				youtube.TransitionToLive(broadcastID)
			} else {
				obsM.StartStreaming()
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
