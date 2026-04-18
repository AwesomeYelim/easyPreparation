package selfupdate

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// UpdateState — 업데이트 상태 열거형
type UpdateState string

const (
	StateIdle            UpdateState = "idle"
	StateChecking        UpdateState = "checking"
	StateDownloading     UpdateState = "downloading"
	StateDownloaded      UpdateState = "downloaded"
	StateApplying        UpdateState = "applying"
	StateRestartRequired UpdateState = "restart_required"
	StateError           UpdateState = "error"
)

// UpdateStatus — 업데이트 현재 상태 스냅샷
type UpdateStatus struct {
	State          UpdateState `json:"state"`
	Percent        float64     `json:"percent"`
	TotalBytes     int64       `json:"totalBytes"`
	DownloadedBytes int64      `json:"downloadedBytes"`
	Version        string      `json:"version"`
	Error          string      `json:"error,omitempty"`
}

// BroadcastFunc — WS 진행률 브로드캐스트 콜백 타입
// 순환 참조 방지를 위해 handlers 대신 콜백으로 주입
type BroadcastFunc func(messageType string, payload map[string]interface{})

// Updater — 다운로드/적용 싱글턴
type Updater struct {
	mu          sync.RWMutex
	status      UpdateStatus
	cancelFunc  context.CancelFunc
	downloadDir string
	broadcast   BroadcastFunc
}

var (
	updaterOnce sync.Once
	globalUpdater *Updater
)

// GetUpdater — 싱글턴 Updater를 반환합니다.
// SetBroadcast로 WS 콜백을 등록해야 진행률이 전송됩니다.
func GetUpdater() *Updater {
	updaterOnce.Do(func() {
		globalUpdater = &Updater{
			status: UpdateStatus{State: StateIdle},
		}
	})
	return globalUpdater
}

// SetBroadcast — WS 브로드캐스트 함수를 등록합니다.
// init.go에서 호출하여 순환 참조 없이 WS 연결.
func (u *Updater) SetBroadcast(fn BroadcastFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.broadcast = fn
}

// SetDownloadDir — 다운로드 임시 디렉토리를 설정합니다.
// 설정하지 않으면 data/update/ 를 사용합니다.
func (u *Updater) SetDownloadDir(dir string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.downloadDir = dir
}

// GetStatus — 현재 상태 스냅샷을 반환합니다 (읽기 전용).
func (u *Updater) GetStatus() UpdateStatus {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.status
}

// setState — 상태를 갱신하고 WS 브로드캐스트를 전송합니다.
func (u *Updater) setState(s UpdateStatus) {
	u.mu.Lock()
	u.status = s
	fn := u.broadcast
	u.mu.Unlock()

	if fn != nil {
		payload := map[string]interface{}{
			"state":           string(s.State),
			"percent":         s.Percent,
			"totalBytes":      s.TotalBytes,
			"downloadedBytes": s.DownloadedBytes,
			"version":         s.Version,
		}
		if s.Error != "" {
			payload["error"] = s.Error
		}
		fn("update_progress", payload)
	}
}

// downloadDirPath — 다운로드 디렉토리 경로를 반환합니다.
// 미설정 시 data/update/ 사용.
func (u *Updater) downloadDirPath() string {
	u.mu.RLock()
	d := u.downloadDir
	u.mu.RUnlock()
	if d != "" {
		return d
	}
	return filepath.Join("data", "update")
}

// Download — Release에서 현재 플랫폼에 맞는 바이너리를 비동기로 다운로드합니다.
// 이미 다운로드/적용 중이면 에러를 반환합니다.
func (u *Updater) Download(release *Release) error {
	u.mu.Lock()
	state := u.status.State
	u.mu.Unlock()

	if state == StateDownloading || state == StateApplying {
		return fmt.Errorf("이미 업데이트가 진행 중입니다 (상태: %s)", state)
	}

	asset := release.FindAsset()
	if asset == nil {
		return fmt.Errorf("현재 플랫폼에 맞는 바이너리 Asset을 찾을 수 없습니다")
	}

	ctx, cancel := context.WithCancel(context.Background())

	u.mu.Lock()
	u.cancelFunc = cancel
	u.mu.Unlock()

	go u.doDownload(ctx, release, asset)
	return nil
}

// doDownload — 실제 다운로드 goroutine
func (u *Updater) doDownload(ctx context.Context, release *Release, asset *ReleaseAsset) {
	version := release.TagName

	u.setState(UpdateStatus{
		State:   StateDownloading,
		Version: version,
	})

	// 다운로드 디렉토리 생성
	dir := u.downloadDirPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		u.setState(UpdateStatus{
			State:   StateError,
			Version: version,
			Error:   fmt.Sprintf("다운로드 디렉토리 생성 실패: %v", err),
		})
		return
	}

	destPath := filepath.Join(dir, asset.Name)

	// 기존 임시 파일 삭제
	_ = os.Remove(destPath)

	// HTTP 다운로드 (연결 타임아웃 30s, 헤더 수신 60s, 전체 10분)
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		u.setState(UpdateStatus{
			State:   StateError,
			Version: version,
			Error:   fmt.Sprintf("요청 생성 실패: %v", err),
		})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			u.setState(UpdateStatus{State: StateIdle, Version: version})
			return
		}
		u.setState(UpdateStatus{
			State:   StateError,
			Version: version,
			Error:   fmt.Sprintf("다운로드 실패: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		u.setState(UpdateStatus{
			State:   StateError,
			Version: version,
			Error:   fmt.Sprintf("다운로드 HTTP 오류: %d", resp.StatusCode),
		})
		return
	}

	totalBytes := asset.Size
	if resp.ContentLength > 0 {
		totalBytes = resp.ContentLength
	}

	f, err := os.Create(destPath)
	if err != nil {
		u.setState(UpdateStatus{
			State:   StateError,
			Version: version,
			Error:   fmt.Sprintf("파일 생성 실패: %v", err),
		})
		return
	}
	defer f.Close()

	// 진행률 추적하며 복사
	var downloaded int64
	buf := make([]byte, 32*1024)
	lastBroadcast := time.Now()

	for {
		if ctx.Err() != nil {
			// 취소 — 임시 파일 정리
			f.Close()
			_ = os.Remove(destPath)
			u.setState(UpdateStatus{State: StateIdle, Version: version})
			return
		}

		nr, readErr := resp.Body.Read(buf)
		if nr > 0 {
			nw, writeErr := f.Write(buf[:nr])
			if writeErr != nil {
				f.Close()
				_ = os.Remove(destPath)
				u.setState(UpdateStatus{
					State:   StateError,
					Version: version,
					Error:   fmt.Sprintf("파일 쓰기 실패: %v", writeErr),
				})
				return
			}
			downloaded += int64(nw)

			// 500ms마다 또는 100KB마다 진행률 브로드캐스트
			var percent float64
			if totalBytes > 0 {
				percent = float64(downloaded) / float64(totalBytes) * 100
			}
			if time.Since(lastBroadcast) > 500*time.Millisecond {
				u.setState(UpdateStatus{
					State:           StateDownloading,
					Percent:         percent,
					TotalBytes:      totalBytes,
					DownloadedBytes: downloaded,
					Version:         version,
				})
				lastBroadcast = time.Now()
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			f.Close()
			_ = os.Remove(destPath)
			u.setState(UpdateStatus{
				State:   StateError,
				Version: version,
				Error:   fmt.Sprintf("다운로드 중 읽기 오류: %v", readErr),
			})
			return
		}
	}
	f.Close()

	// 체크섬 검증 (checksums.txt가 있는 경우)
	checksumAsset := release.ChecksumsAsset()
	if checksumAsset != nil {
		log.Printf("[updater] 체크섬 검증 중: %s", asset.Name)
		if err := VerifyChecksum(destPath, checksumAsset.BrowserDownloadURL); err != nil {
			_ = os.Remove(destPath)
			u.setState(UpdateStatus{
				State:   StateError,
				Version: version,
				Error:   fmt.Sprintf("체크섬 검증 실패: %v", err),
			})
			return
		}
		log.Printf("[updater] 체크섬 검증 완료")
	} else {
		log.Printf("[updater] checksums.txt 없음 — 체크섬 검증 스킵")
	}

	log.Printf("[updater] 다운로드 완료: %s (%.1f MB)", asset.Name, float64(downloaded)/1024/1024)
	u.setState(UpdateStatus{
		State:           StateDownloaded,
		Percent:         100,
		TotalBytes:      totalBytes,
		DownloadedBytes: downloaded,
		Version:         version,
	})
}

// CancelDownload — 진행 중인 다운로드를 취소합니다.
func (u *Updater) CancelDownload() {
	u.mu.Lock()
	cancel := u.cancelFunc
	u.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// Apply — 다운로드된 바이너리로 현재 실행 파일을 교체합니다.
// 상태가 StateDownloaded가 아니면 에러를 반환합니다.
func (u *Updater) Apply() error {
	u.mu.RLock()
	status := u.status
	u.mu.RUnlock()

	if status.State != StateDownloaded {
		return fmt.Errorf("다운로드가 완료된 상태가 아닙니다 (현재: %s)", status.State)
	}

	u.setState(UpdateStatus{
		State:   StateApplying,
		Version: status.Version,
		Percent: 100,
	})

	// 현재 실행 파일 경로
	execPath, err := os.Executable()
	if err != nil {
		u.setState(UpdateStatus{
			State:   StateError,
			Version: status.Version,
			Error:   fmt.Sprintf("실행 파일 경로 확인 실패: %v", err),
		})
		return err
	}

	// 다운로드된 바이너리 경로 찾기
	dir := u.downloadDirPath()
	entries, err := os.ReadDir(dir)
	if err != nil {
		u.setState(UpdateStatus{
			State:   StateError,
			Version: status.Version,
			Error:   fmt.Sprintf("다운로드 디렉토리 읽기 실패: %v", err),
		})
		return err
	}

	var newBinPath string
	for _, e := range entries {
		if !e.IsDir() && e.Name() != "checksums.txt" {
			candidate := filepath.Join(dir, e.Name())
			// .bak 파일은 스킵
			if filepath.Ext(candidate) == ".bak" {
				continue
			}
			newBinPath = candidate
			break
		}
	}

	if newBinPath == "" {
		err := fmt.Errorf("다운로드 디렉토리에서 바이너리를 찾을 수 없습니다: %s", dir)
		u.setState(UpdateStatus{
			State:   StateError,
			Version: status.Version,
			Error:   err.Error(),
		})
		return err
	}

	if err := u.applyBinary(newBinPath, execPath, status.Version); err != nil {
		u.setState(UpdateStatus{
			State:   StateError,
			Version: status.Version,
			Error:   err.Error(),
		})
		return err
	}

	return nil
}

// CleanupBackup — 이전 업데이트로 남은 .bak 파일을 정리합니다.
// 서버 시작 시 호출합니다.
func (u *Updater) CleanupBackup() {
	execPath, err := os.Executable()
	if err != nil {
		return
	}
	bakPath := execPath + ".bak"
	if _, err := os.Stat(bakPath); err == nil {
		if removeErr := os.Remove(bakPath); removeErr != nil {
			log.Printf("[updater] 백업 파일 삭제 실패: %v", removeErr)
		} else {
			log.Printf("[updater] 이전 백업 파일 정리 완료: %s", bakPath)
		}
	}
}
