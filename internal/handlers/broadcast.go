package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"easyPreparation_1.0/internal/httpx"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/thumbnail"
	"easyPreparation_1.0/internal/youtube"
)

// prepareYouTubeBroadcast — 수동(setup-obs)·자동(스케줄러) 방송 시작이 공유하는 앞부분.
// YouTube 방송 생성 + 스트림 바인딩 → 썸네일 생성·업로드 → OBS 가 송출 중이면 정지(최대 5초 대기).
// 이후 단계(OBS 스트림 설정 방식, 송출 시작, live 전환)는 호출자마다 달라 여기서 하지 않는다.
//
// waitThumbnail: true 면 썸네일 업로드가 끝날 때까지 기다린다(스케줄러 — upcoming 상태에서 확실히 반영),
// false 면 백그라운드로 올린다(수동 버튼 — 응답 지연 방지).
func prepareYouTubeBroadcast(worshipType, title, description string, waitThumbnail bool) (server, key, broadcastID string, err error) {
	server, key, broadcastID, err = youtube.CreateBroadcastAndBind(title, description)
	if err != nil {
		return "", "", "", err
	}

	if waitThumbnail {
		GenerateAndUploadThumbnailTo(worshipType, broadcastID)
	} else {
		go GenerateAndUploadThumbnailTo(worshipType, broadcastID)
	}

	// 기존 송출을 새 스트림 키로 바꾸려면 먼저 멈춰야 한다
	obsM := obs.Get()
	if obsM.GetStreamStatus().Active {
		obsM.StopStreaming()
		for i := 0; i < 10; i++ {
			if !obsM.GetStreamStatus().Active {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	return server, key, broadcastID, nil
}

// YouTubeSetupOBSHandler — POST /api/youtube/setup-obs
// YouTube 방송 생성 + 썸네일 업로드 + OBS 스트림 키 설정. 송출 시작은 하지 않는다(사용자가 OBS 에서 시작).
func YouTubeSetupOBSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	title := "라이브 예배"
	worshipType := "main_worship"
	var body struct {
		Title       string `json:"title"`
		WorshipType string `json:"worshipType"`
	}
	if json.NewDecoder(r.Body).Decode(&body) == nil {
		if body.Title != "" {
			title = body.Title
		}
		if body.WorshipType != "" {
			worshipType = body.WorshipType
		}
	}

	description := ""
	if cfg, cErr := thumbnail.LoadConfig(); cErr == nil {
		sermonTitle, scripture := SermonDataForWorship(worshipType)
		description = cfg.ResolveDescription(worshipType, time.Now(), sermonTitle, scripture)
	}

	server, key, broadcastID, err := prepareYouTubeBroadcast(worshipType, title, description, false)
	if err != nil {
		httpx.JSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	if err := obs.Get().SetStreamSettings(server, key); err != nil {
		httpx.JSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": "방송 생성됨, OBS 설정 실패: " + err.Error()})
		return
	}

	log.Printf("[broadcast] setup-obs 완료: broadcastId=%s", broadcastID)
	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"message":     "YouTube 방송 생성 + OBS 스트림 설정 + 썸네일 업로드 완료",
		"broadcastId": broadcastID,
	})
}
