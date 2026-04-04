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

	// 2. 기존 upcoming 방송이 있으면 재사용
	bcResp, err := svc.LiveBroadcasts.List([]string{"id", "snippet", "status"}).
		BroadcastStatus("upcoming").BroadcastType("all").Do()
	if err == nil && len(bcResp.Items) > 0 {
		bc := bcResp.Items[0]
		// 제목만 업데이트
		bc.Snippet.Title = title
		svc.LiveBroadcasts.Update([]string{"snippet"}, bc).Do()

		log.Printf("[youtube] 기존 방송 재사용: %s (%s)", bc.Id, title)
		info := stream.Cdn.IngestionInfo
		return info.IngestionAddress, info.StreamName, bc.Id, nil
	}

	// 3. 새 방송 생성
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

	// 4. 스트림을 방송에 바인딩
	_, err = svc.LiveBroadcasts.Bind(broadcast.Id, []string{"id", "contentDetails"}).StreamId(stream.Id).Do()
	if err != nil {
		return "", "", "", fmt.Errorf("스트림 바인딩 실패: %w", err)
	}
	log.Printf("[youtube] 스트림 바인딩 완료: broadcast=%s, stream=%s", broadcast.Id, stream.Id)

	info := stream.Cdn.IngestionInfo
	return info.IngestionAddress, info.StreamName, broadcast.Id, nil
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

func (m *Manager) createService(tok *oauth2.Token) (*yt.Service, error) {
	client := m.oauthConf.Client(context.Background(), tok)
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

// StatusHandler — GET /api/youtube/status → 연결 상태
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	m := Get()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected": m.IsEnabled(),
	})
}

// DefaultTokenPath — 기본 토큰 파일 경로
func DefaultTokenPath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "youtube_token.json")
}

// DefaultOAuthPath — 기본 OAuth 설정 파일 경로
func DefaultOAuthPath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "config", "google_oauth.json")
}
