package version

import "sync"

// Info — 빌드 시 주입되는 버전 정보
type Info struct {
	Version   string
	Commit    string
	BuildTime string
}

var (
	mu       sync.RWMutex
	current  = Info{
		Version:   "dev",
		Commit:    "unknown",
		BuildTime: "unknown",
	}
)

// Set — cmd/*/main.go에서 ldflags로 주입된 값을 등록
func Set(version, commit, buildTime string) {
	mu.Lock()
	defer mu.Unlock()
	if version != "" {
		current.Version = version
	}
	if commit != "" {
		current.Commit = commit
	}
	if buildTime != "" {
		current.BuildTime = buildTime
	}
}

// Get — 현재 버전 정보를 반환 (읽기 전용)
func Get() Info {
	mu.RLock()
	defer mu.RUnlock()
	return current
}
