package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	yt "google.golang.org/api/youtube/v3"

	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/sysopen"
)

// Manager — YouTube API 싱글턴
type Manager struct {
	mu        sync.Mutex
	service   *yt.Service
	enabled   bool
	oauthConf *oauth2.Config
	tokenPath string
}

var instance *Manager

// activeTransitions — broadcastID별 TransitionToLive 중복 실행 방지
var activeTransitions sync.Map

// Init — 서버 시작 시 호출. OAuth 설정 로드 + 토큰 있으면 서비스 초기화
func Init(oauthConfigPath, tokenPath string) {
	m := &Manager{tokenPath: tokenPath}

	data, err := os.ReadFile(oauthConfigPath)
	if err != nil {
		log.Printf("[youtube] OAuth 설정 없음 (%s) — YouTube 비활성", oauthConfigPath)
		instance = m
		return
	}

	var creds struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.Unmarshal(data, &creds); err != nil || creds.ClientID == "" {
		log.Printf("[youtube] OAuth 설정 파싱 실패 — YouTube 비활성")
		instance = m
		return
	}

	m.oauthConf = &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube",
			"https://www.googleapis.com/auth/youtube.upload",
		},
		Endpoint: google.Endpoint,
		RedirectURL: "http://localhost:8080/api/youtube/callback",
	}

	// 기존 토큰이 있으면 서비스 즉시 초기화
	if tok, err := loadToken(tokenPath); err == nil {
		if svc, err := m.createService(tok); err == nil {
			m.service = svc
			m.enabled = true
			log.Println("[youtube] 기존 토큰으로 초기화 성공")
		}
	}

	instance = m
}

// Get — 싱글턴 반환
func Get() *Manager {
	if instance == nil {
		return &Manager{}
	}
	return instance
}

// IsEnabled — YouTube 연동 활성 여부
func (m *Manager) IsEnabled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enabled
}

// GetService — YouTube API 서비스 반환
func (m *Manager) GetService() *yt.Service {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.service
}

// GetAuthURL — OAuth 동의 화면 URL 반환
func (m *Manager) GetAuthURL() (string, error) {
	if m.oauthConf == nil {
		return "", fmt.Errorf("OAuth 설정 없음")
	}
	return m.oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

// HandleCallback — OAuth 콜백 처리 → 토큰 저장 → 서비스 초기화
func (m *Manager) HandleCallback(code string) error {
	if m.oauthConf == nil {
		return fmt.Errorf("OAuth 설정 없음")
	}

	tok, err := m.oauthConf.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("토큰 교환 실패: %w", err)
	}

	if err := saveToken(m.tokenPath, tok); err != nil {
		return fmt.Errorf("토큰 저장 실패: %w", err)
	}

	svc, err := m.createService(tok)
	if err != nil {
		return fmt.Errorf("서비스 생성 실패: %w", err)
	}

	m.mu.Lock()
	m.service = svc
	m.enabled = true
	m.mu.Unlock()

	log.Println("[youtube] OAuth 인증 완료 — YouTube 활성화")
	return nil
}

// UploadThumbnail — 현재 활성/예정 라이브 방송에 썸네일 업로드
func UploadThumbnail(imagePath string) error {
	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled || svc == nil {
		log.Println("[youtube] 비활성 — 썸네일 업로드 스킵")
		return nil
	}

	// 활성 라이브 방송 찾기
	resp, err := svc.LiveBroadcasts.List([]string{"id", "snippet"}).
		BroadcastStatus("active").
		BroadcastType("all").
		Do()
	if err != nil {
		return fmt.Errorf("라이브 방송 조회 실패: %w", err)
	}

	if len(resp.Items) == 0 {
		// upcoming도 확인
		resp, err = svc.LiveBroadcasts.List([]string{"id", "snippet"}).
			BroadcastStatus("upcoming").
			BroadcastType("all").
			Do()
		if err != nil {
			return fmt.Errorf("예정 방송 조회 실패: %w", err)
		}
		if len(resp.Items) == 0 {
			log.Println("[youtube] 활성/예정 라이브 방송 없음 — 썸네일 업로드 스킵")
			return nil
		}
	}

	broadcastID := resp.Items[0].Id
	log.Printf("[youtube] 방송 발견: %s (%s)", resp.Items[0].Snippet.Title, broadcastID)
	return UploadThumbnailToBroadcast(broadcastID, imagePath)
}

// UploadThumbnailToBroadcast — 특정 방송 ID에 썸네일 업로드 (upcoming 상태에서 호출해야 확실히 반영됨)
func UploadThumbnailToBroadcast(broadcastID, imagePath string) error {
	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled || svc == nil {
		return fmt.Errorf("YouTube 미연결")
	}

	f, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("썸네일 파일 열기 실패: %w", err)
	}
	defer f.Close()

	_, err = svc.Thumbnails.Set(broadcastID).Media(f).Do()
	if err != nil {
		return fmt.Errorf("썸네일 업로드 실패: %w", err)
	}

	log.Printf("[youtube] 썸네일 업로드 성공: %s → %s", imagePath, broadcastID)
	return nil
}

// UpdateBroadcastTitle — 현재 활성/예정 라이브 방송의 제목을 변경
func UpdateBroadcastTitle(title string) error {
	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled || svc == nil {
		return fmt.Errorf("YouTube 미연결")
	}

	// 활성 방송 찾기
	resp, err := svc.LiveBroadcasts.List([]string{"id", "snippet"}).
		BroadcastStatus("active").BroadcastType("all").Do()
	if err != nil {
		return fmt.Errorf("라이브 방송 조회 실패: %w", err)
	}
	if len(resp.Items) == 0 {
		// upcoming도 확인
		resp, err = svc.LiveBroadcasts.List([]string{"id", "snippet"}).
			BroadcastStatus("upcoming").BroadcastType("all").Do()
		if err != nil {
			return fmt.Errorf("예정 방송 조회 실패: %w", err)
		}
		if len(resp.Items) == 0 {
			return fmt.Errorf("활성/예정 라이브 방송 없음")
		}
	}

	broadcast := resp.Items[0]
	broadcast.Snippet.Title = title

	_, err = svc.LiveBroadcasts.Update([]string{"snippet"}, broadcast).Do()
	if err != nil {
		return fmt.Errorf("방송 제목 변경 실패: %w", err)
	}

	log.Printf("[youtube] 방송 제목 변경: %s → %s", broadcast.Id, title)
	return nil
}

// UpdateBroadcastTitleByID — 특정 broadcast_id의 방송 제목 변경
func UpdateBroadcastTitleByID(broadcastID, title string) error {
	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()
	if !enabled || svc == nil {
		return fmt.Errorf("YouTube 미연결")
	}
	resp, err := svc.LiveBroadcasts.List([]string{"id", "snippet"}).Id(broadcastID).Do()
	if err != nil || len(resp.Items) == 0 {
		return fmt.Errorf("방송 조회 실패: %v", err)
	}
	bc := resp.Items[0]
	bc.Snippet.Title = title
	_, err = svc.LiveBroadcasts.Update([]string{"snippet"}, bc).Do()
	if err != nil {
		return fmt.Errorf("방송 제목 변경 실패: %w", err)
	}
	log.Printf("[youtube] 방송 제목 변경: %s → %s", broadcastID, title)
	return nil
}

// CreateBroadcastAndBind — YouTube 라이브 방송 생성 + 스트림 바인딩 + 스트림 키 반환
// 완전 자동화: 방송 생성 → 스트림 연결 → OBS에서 송출하면 라이브
func CreateBroadcastAndBind(title string) (server string, key string, broadcastID string, err error) {
	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled || svc == nil {
		return "", "", "", fmt.Errorf("YouTube 미연결")
	}

	// 1. 기존 스트림 가져오기 (없으면 생성)
	streamsResp, err := svc.LiveStreams.List([]string{"id", "cdn", "snippet"}).Mine(true).Do()
	if err != nil {
		return "", "", "", fmt.Errorf("스트림 조회 실패: %w", err)
	}

	var stream *yt.LiveStream
	if len(streamsResp.Items) > 0 {
		stream = streamsResp.Items[0]
	} else {
		// 스트림 생성
		stream, err = svc.LiveStreams.Insert([]string{"snippet", "cdn"}, &yt.LiveStream{
			Snippet: &yt.LiveStreamSnippet{Title: "easyPreparation Stream"},
			Cdn: &yt.CdnSettings{
				FrameRate:     "30fps",
				IngestionType: "rtmp",
				Resolution:    "1080p",
			},
		}).Do()
		if err != nil {
			return "", "", "", fmt.Errorf("스트림 생성 실패: %w", err)
		}
		log.Printf("[youtube] 새 스트림 생성: %s", stream.Id)
	}

	// 2. 고착 방송 정리
	// testing/testStarting: complete 전환 후 삭제
	// ready: complete 전환 불가(YouTube API 제한) → 직접 삭제
	for _, stuckStatus := range []string{"testStarting", "testing"} {
		stuckResp, err := svc.LiveBroadcasts.List([]string{"id", "status"}).
			BroadcastStatus(stuckStatus).BroadcastType("all").Do()
		if err != nil {
			continue
		}
		for _, bc := range stuckResp.Items {
			if _, err := svc.LiveBroadcasts.Transition("complete", bc.Id, []string{"id"}).Do(); err != nil {
				svc.LiveBroadcasts.Delete(bc.Id).Do()
				log.Printf("[youtube] 고착 방송 삭제: %s (%s)", bc.Id, stuckStatus)
			} else {
				log.Printf("[youtube] 고착 방송 종료: %s (%s→complete)", bc.Id, stuckStatus)
			}
		}
	}
	// ready 상태 방송은 complete 전환 불가 → 직접 삭제 시도
	if readyResp, err := svc.LiveBroadcasts.List([]string{"id", "status"}).
		BroadcastStatus("ready").BroadcastType("all").Do(); err == nil {
		for _, bc := range readyResp.Items {
			if err := svc.LiveBroadcasts.Delete(bc.Id).Do(); err != nil {
				log.Printf("[youtube] 고착 ready 방송 삭제 실패 (무시): %s: %v", bc.Id, err)
			} else {
				log.Printf("[youtube] 고착 ready 방송 삭제: %s", bc.Id)
			}
		}
	}

	// 3. 기존 upcoming 방송이 있으면 재사용 (ready 상태면 삭제 후 새로 생성)
	bcResp, err := svc.LiveBroadcasts.List([]string{"id", "snippet", "status", "contentDetails"}).
		BroadcastStatus("upcoming").BroadcastType("all").Do()
	if err == nil && len(bcResp.Items) > 0 {
		bc := bcResp.Items[0]
		// ready 상태 방송은 autoStart 설정·전환 모두 실패하므로 삭제 후 새로 생성
		if bc.Status != nil && bc.Status.LifeCycleStatus == "ready" {
			log.Printf("[youtube] ready 방송 발견 — 삭제 후 새로 생성: %s", bc.Id)
			if derr := svc.LiveBroadcasts.Delete(bc.Id).Do(); derr != nil {
				log.Printf("[youtube] ready 방송 삭제 실패 (무시): %v", derr)
			}
			// fall through to step 4
		} else {
			// created 상태 방송 재사용
			bcStatus := bc.Status
			bc.Status = nil // Status 포함 시 400 unexpectedPart 에러 발생
			bc.Snippet.Title = title
			bc.ContentDetails.EnableAutoStart = true
			bc.ContentDetails.EnableAutoStop = true
			bc.ContentDetails.ForceSendFields = append(bc.ContentDetails.ForceSendFields, "BroadcastStreamDelayMs", "EnableAutoStart", "EnableAutoStop")
			if _, uerr := svc.LiveBroadcasts.Update([]string{"snippet", "contentDetails"}, bc).Do(); uerr != nil {
				log.Printf("[youtube] 방송 업데이트 실패 (autoStart 미적용 가능): %v", uerr)
			}
			// 스트림 바인딩 (이미 바인딩되어 있으면 무시됨)
			if _, berr := svc.LiveBroadcasts.Bind(bc.Id, []string{"id"}).StreamId(stream.Id).Do(); berr != nil {
				log.Printf("[youtube] 스트림 바인딩 실패: %v", berr)
			}

			info := stream.Cdn.IngestionInfo
			keyPrefix := info.StreamName
			if len(keyPrefix) > 8 {
				keyPrefix = keyPrefix[:8]
			}
			log.Printf("[youtube] 기존 방송 재사용: %s (%s, %s), autoStart=true, stream=%s, key=%s...",
				bc.Id, title, bcStatus.LifeCycleStatus, stream.Id, keyPrefix)
			return info.IngestionAddress, info.StreamName, bc.Id, nil
		}
	}

	// 4. 새 방송 생성
	now := time.Now().Add(1 * time.Minute) // 1분 뒤 시작
	broadcast, err := svc.LiveBroadcasts.Insert([]string{"snippet", "status", "contentDetails"}, &yt.LiveBroadcast{
		Snippet: &yt.LiveBroadcastSnippet{
			Title:              title,
			ScheduledStartTime: now.Format(time.RFC3339),
		},
		Status: &yt.LiveBroadcastStatus{
			PrivacyStatus:           "public",
			SelfDeclaredMadeForKids: false,
		},
		ContentDetails: &yt.LiveBroadcastContentDetails{
			EnableAutoStart: true,
			EnableAutoStop:  true,
		},
	}).Do()
	if err != nil {
		return "", "", "", fmt.Errorf("방송 생성 실패: %w", err)
	}
	log.Printf("[youtube] 방송 생성: %s (%s)", broadcast.Id, title)

	// 5. 스트림을 방송에 바인딩
	_, err = svc.LiveBroadcasts.Bind(broadcast.Id, []string{"id", "contentDetails"}).StreamId(stream.Id).Do()
	if err != nil {
		return "", "", "", fmt.Errorf("스트림 바인딩 실패: %w", err)
	}
	info := stream.Cdn.IngestionInfo
	keyPrefix := info.StreamName
	if len(keyPrefix) > 8 {
		keyPrefix = keyPrefix[:8]
	}
	log.Printf("[youtube] 스트림 바인딩 완료: broadcast=%s, stream=%s, key=%s...", broadcast.Id, stream.Id, keyPrefix)
	return info.IngestionAddress, info.StreamName, broadcast.Id, nil
}

// TransitionToLive — 방송을 live로 전환 (OBS 스트리밍 시작 후 호출)
// 같은 broadcastID에 대해 동시에 하나만 실행됨.
func TransitionToLive(broadcastID string) error {
	// 중복 실행 방지: 동일 broadcastID가 이미 전환 중이면 스킵
	if _, loaded := activeTransitions.LoadOrStore(broadcastID, true); loaded {
		log.Printf("[youtube] TransitionToLive 이미 실행 중: %s — 스킵", broadcastID)
		return nil
	}
	defer activeTransitions.Delete(broadcastID)

	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled || svc == nil {
		return fmt.Errorf("YouTube 미연결")
	}

	// 최대 180초 대기 (90회 × 2초)
	invalidTransitionCount := 0
	var boundStreamID string
	for i := 0; i < 90; i++ {
		time.Sleep(2 * time.Second)
		resp, err := svc.LiveBroadcasts.List([]string{"id", "status", "contentDetails"}).Id(broadcastID).Do()
		if err != nil {
			log.Printf("[youtube] 방송 상태 조회 실패 (재시도): %v", err)
			continue
		}
		if len(resp.Items) == 0 {
			continue
		}
		bc := resp.Items[0]
		status := bc.Status.LifeCycleStatus
		if boundStreamID == "" && bc.ContentDetails != nil {
			boundStreamID = bc.ContentDetails.BoundStreamId
		}
		log.Printf("[youtube] 방송 상태: %s", status)

		switch status {
		case "live":
			log.Printf("[youtube] 라이브 상태 확인")
			return nil
		case "complete":
			return fmt.Errorf("방송이 이미 종료됨")
		case "ready", "testStarting":
			// 30초마다 YouTube 스트림 수신 여부 진단
			if boundStreamID != "" && i > 0 && i%15 == 0 {
				if sr, serr := svc.LiveStreams.List([]string{"id", "status"}).Id(boundStreamID).Do(); serr == nil && len(sr.Items) > 0 {
					ss := sr.Items[0].Status.StreamStatus
					log.Printf("[youtube] 스트림 수신 상태: %s (id=%s)", ss, boundStreamID)
				}
			}
			// autoStart=true여도 YouTube 감지가 느릴 수 있음 — 30초 후부터 10초 간격으로 전환 시도
			// testing 전환 실패 시 monitorStream 비활성 방송으로 간주 → live 직접 전환
			if i >= 15 && i%5 == 0 {
				if _, err := svc.LiveBroadcasts.Transition("testing", broadcastID, []string{"id", "status"}).Do(); err != nil {
					invalidTransitionCount++
					if invalidTransitionCount <= 3 {
						log.Printf("[youtube] ready→testing 전환 대기 중: %v", err)
					}
					// testing 전환 불가 → live 직접 전환 시도 (monitorStream 비활성 방송)
					if _, lerr := svc.LiveBroadcasts.Transition("live", broadcastID, []string{"id", "status"}).Do(); lerr != nil {
						if invalidTransitionCount <= 3 {
							log.Printf("[youtube] ready→live 직접 전환 실패: %v", lerr)
						}
					} else {
						log.Printf("[youtube] ready→live 직접 전환 성공")
						return nil
					}
				} else {
					log.Printf("[youtube] ready→testing 전환 성공")
					invalidTransitionCount = 0
				}
			}
			continue
		case "testing":
			_, err := svc.LiveBroadcasts.Transition("live", broadcastID, []string{"id", "status"}).Do()
			if err != nil {
				log.Printf("[youtube] live 전환 실패 (재시도): %v", err)
				continue
			}
			log.Printf("[youtube] 방송 live 전환 성공: %s", broadcastID)
			return nil
		}
	}
	return fmt.Errorf("방송 live 전환 타임아웃 (180초)")
}

// CleanupBroadcasts — 고착/진행 중인 방송 모두 정리 (complete 전환 또는 삭제)
// active → complete, testing/testStarting/ready → complete 시도 후 삭제, upcoming → 삭제
func CleanupBroadcasts() {
	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled || svc == nil {
		return
	}

	// active/testing/testStarting: complete 전환 시도, 실패 시 삭제
	for _, status := range []string{"active", "testing", "testStarting"} {
		resp, err := svc.LiveBroadcasts.List([]string{"id", "status"}).
			BroadcastStatus(status).BroadcastType("all").Do()
		if err != nil {
			continue
		}
		for _, bc := range resp.Items {
			_, err := svc.LiveBroadcasts.Transition("complete", bc.Id, []string{"id"}).Do()
			if err != nil {
				log.Printf("[youtube] 방송 종료 실패 %s(%s): %v, 삭제 시도", bc.Id, status, err)
				svc.LiveBroadcasts.Delete(bc.Id).Do()
			} else {
				log.Printf("[youtube] 방송 종료: %s (%s→complete)", bc.Id, status)
			}
		}
	}

	// ready: 스트림 없이 고착된 상태 — 삭제
	for _, status := range []string{"ready", "upcoming"} {
		resp, err := svc.LiveBroadcasts.List([]string{"id", "status"}).
			BroadcastStatus(status).BroadcastType("all").Do()
		if err != nil {
			continue
		}
		for _, bc := range resp.Items {
			if err := svc.LiveBroadcasts.Delete(bc.Id).Do(); err != nil {
				log.Printf("[youtube] 방송 삭제 실패 %s(%s): %v", bc.Id, status, err)
			} else {
				log.Printf("[youtube] 방송 삭제: %s (%s)", bc.Id, status)
			}
		}
	}
}

// UploadVideo — YouTube에 동영상 파일을 업로드하고 videoId 반환
// 업로드 완료 후 thumbnailPath가 비어 있지 않으면 썸네일도 설정
func UploadVideo(videoPath, title, description, privacy, thumbnailPath string) (string, error) {
	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled || svc == nil {
		return "", fmt.Errorf("YouTube 미연결")
	}

	f, err := os.Open(videoPath)
	if err != nil {
		return "", fmt.Errorf("비디오 파일 열기 실패: %w", err)
	}
	defer f.Close()

	if privacy == "" {
		privacy = "public"
	}

	video := &yt.Video{
		Snippet: &yt.VideoSnippet{
			Title:       title,
			Description: description,
			CategoryId:  "29",
		},
		Status: &yt.VideoStatus{
			PrivacyStatus: privacy,
		},
	}

	resp, err := svc.Videos.Insert([]string{"snippet", "status"}, video).Media(f).Do()
	if err != nil {
		return "", fmt.Errorf("비디오 업로드 실패: %w", err)
	}
	log.Printf("[youtube] 비디오 업로드 성공: %s (%s)", resp.Id, title)

	if thumbnailPath != "" {
		if err := UploadThumbnailToBroadcast(resp.Id, thumbnailPath); err != nil {
			log.Printf("[youtube] 썸네일 설정 실패 (업로드는 완료): %v", err)
		} else {
			log.Printf("[youtube] 썸네일 설정 완료: %s", resp.Id)
		}
	}

	return resp.Id, nil
}

// GetStreamInfo — YouTube 라이브 스트림의 RTMP URL + 스트림 키 조회
func GetStreamInfo() (server string, key string, err error) {
	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled || svc == nil {
		return "", "", fmt.Errorf("YouTube 미연결")
	}

	resp, err := svc.LiveStreams.List([]string{"cdn"}).Mine(true).Do()
	if err != nil {
		return "", "", fmt.Errorf("라이브 스트림 조회 실패: %w", err)
	}

	if len(resp.Items) == 0 {
		return "", "", fmt.Errorf("라이브 스트림이 없습니다 (YouTube Studio에서 스트림을 먼저 생성하세요)")
	}

	info := resp.Items[0].Cdn.IngestionInfo
	log.Printf("[youtube] 스트림 정보: server=%s, key=%s...", info.IngestionAddress, info.StreamName[:8])
	return info.IngestionAddress, info.StreamName, nil
}

// ── 내부 함수 ──

// persistingTokenSource — 토큰 갱신 시 자동으로 디스크에 저장
type persistingTokenSource struct {
	src       oauth2.TokenSource
	tokenPath string
	mu        sync.Mutex
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := p.src.Token()
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := saveToken(p.tokenPath, tok); err != nil {
		log.Printf("[youtube] 토큰 자동 저장 실패: %v", err)
	}
	return tok, nil
}

func (m *Manager) createService(tok *oauth2.Token) (*yt.Service, error) {
	base := m.oauthConf.TokenSource(context.Background(), tok)
	src := &persistingTokenSource{src: oauth2.ReuseTokenSource(tok, base), tokenPath: m.tokenPath}
	client := oauth2.NewClient(context.Background(), src)
	return yt.New(client)
}

func loadToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func saveToken(path string, tok *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ── HTTP 핸들러 ──

// AuthHandler — GET /api/youtube/auth → OAuth URL 리디렉트
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	m := Get()
	url, err := m.GetAuthURL()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// OpenAuthHandler — GET /api/youtube/open-auth
// Desktop 모드: Wails WebView 대신 시스템 브라우저로 OAuth 인증 페이지를 엽니다.
func OpenAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	m := Get()
	authURL, err := m.GetAuthURL()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := sysopen.URL(authURL); err != nil {
		log.Printf("[youtube] 브라우저 열기 실패: %v", err)
		http.Error(w, "브라우저 열기 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// CallbackHandler — GET /api/youtube/callback → 토큰 교환 + 저장
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code 파라미터 없음", http.StatusBadRequest)
		return
	}

	m := Get()
	if err := m.HandleCallback(code); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<html><body><h2>YouTube 연결 완료!</h2><p>이 창을 닫아도 됩니다.</p><script>window.close()</script></body></html>`)
}

// StatusHandler — GET /api/youtube/status → 연결 상태 + 방송 목록
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	m := Get()
	m.mu.Lock()
	svc := m.service
	enabled := m.enabled
	m.mu.Unlock()

	result := map[string]interface{}{"connected": enabled}

	if enabled && svc != nil {
		var broadcasts []map[string]interface{}
		for _, status := range []string{"active", "upcoming", "complete"} {
			resp, err := svc.LiveBroadcasts.List([]string{"id", "snippet", "status"}).
				BroadcastStatus(status).BroadcastType("all").Do()
			if err != nil {
				continue
			}
			for _, bc := range resp.Items {
				broadcasts = append(broadcasts, map[string]interface{}{
					"id":        bc.Id,
					"title":     bc.Snippet.Title,
					"status":    bc.Status.LifeCycleStatus,
					"privacy":   bc.Status.PrivacyStatus,
					"channelId": bc.Snippet.ChannelId,
				})
			}
		}
		result["broadcasts"] = broadcasts

		// 스트림 정보
		streamsResp, err := svc.LiveStreams.List([]string{"id", "cdn", "status"}).Mine(true).Do()
		if err == nil && len(streamsResp.Items) > 0 {
			s := streamsResp.Items[0]
			result["stream"] = map[string]interface{}{
				"id":     s.Id,
				"health": s.Status.HealthStatus.Status,
				"status": s.Status.StreamStatus,
				"server": s.Cdn.IngestionInfo.IngestionAddress,
				"key":    s.Cdn.IngestionInfo.StreamName[:8] + "...",
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DefaultTokenPath — 기본 토큰 파일 경로
func DefaultTokenPath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "youtube_token.json")
}

// DefaultOAuthPath — 기본 OAuth 설정 파일 경로
func DefaultOAuthPath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "config", "google_oauth.json")
}
