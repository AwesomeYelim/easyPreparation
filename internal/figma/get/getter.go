package get

import (
	"easyPreparation_1.0/internal/handlers"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const cacheCountFile = ".figma_count"

// GetNodes는 Figma 파일 노드를 가져옵니다.
// 외부에서 직접 호출할 때 사용 (예: 토큰 검증용).
func (i *Info) GetNodes() error {
	nodes, _, err := fetchNodes(*i.Key, *i.Token)
	if err != nil {
		return fmt.Errorf("Figma 파일 조회 실패: %v", err)
	}
	i.Nodes = nodes
	if len(i.Nodes) == 0 {
		return fmt.Errorf("Figma 파일에 노드가 없습니다")
	}
	handlers.BroadcastProgress("Get figma image elements", 1, fmt.Sprintf("Got %d documents", len(i.Nodes)))
	return nil
}

// GetFigmaImage는 Figma 프레임 이미지를 path에 다운로드합니다.
// 캐시된 PNG 수가 저장된 기대 개수와 일치하면 Figma API를 전혀 호출하지 않습니다.
// 디자인을 새로 반영하려면 tmp 폴더를 비우면 됩니다.
func (i *Info) GetFigmaImage(path string, frameName string) {
	_ = os.MkdirAll(path, 0755)

	// 캐시 유효성 검사: PNG 수 == 저장된 기대 수
	if i.loadFromCache(path) {
		expected := readCacheCount(path)
		if expected > 0 && len(i.PathInfo) == expected {
			handlers.BroadcastProgress("Figma cache hit", 1,
				fmt.Sprintf("캐시 사용 (%d개) — 새 이미지를 받으려면 tmp 폴더를 비워주세요", expected))
			return
		}
	}

	// 캐시 미스 → 이때만 Figma API 호출
	handlers.BroadcastProgress("Figma download", 1, fmt.Sprintf("'%s' 이미지 다운로드 중...", frameName))

	if err := i.GetNodes(); err != nil {
		handlers.BroadcastProgress("Figma node error", -1, fmt.Sprintf("노드 조회 실패: %v", err))
		return
	}

	i.AssembledNodes = i.GetFrames(frameName)
	if len(i.AssembledNodes) == 0 {
		handlers.BroadcastProgress("Figma frame error", -1, fmt.Sprintf("'%s' 프레임을 찾을 수 없습니다", frameName))
		return
	}

	// 전체 재다운로드
	i.PathInfo = make(map[string]string)
	_ = os.RemoveAll(path)
	_ = os.MkdirAll(path, 0755)

	for _, node := range i.AssembledNodes {
		if err := i.GetImage(path, node.ID, node.Name); err != nil {
			handlers.BroadcastProgress("Figma image error", -1, fmt.Sprintf("%s 다운로드 실패: %v", node.Name, err))
		}
	}

	// 다운로드된 수를 캐시 메타데이터로 저장
	_ = os.WriteFile(filepath.Join(path, cacheCountFile),
		[]byte(strconv.Itoa(len(i.AssembledNodes))), 0644)
}

// loadFromCache는 캐시 디렉토리의 PNG 파일을 PathInfo에 로드합니다.
func (i *Info) loadFromCache(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			name := strings.TrimSuffix(entry.Name(), ".png")
			i.PathInfo[name] = filepath.Join(path, entry.Name())
		}
	}
	return len(i.PathInfo) > 0
}

// readCacheCount는 저장된 기대 프레임 수를 읽습니다.
func readCacheCount(path string) int {
	data, err := os.ReadFile(filepath.Join(path, cacheCountFile))
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return n
}

func (i *Info) GetImage(createdPath, id, name string) error {
	imgURL, err := fetchImageURL(*i.Key, *i.Token, id)
	if err != nil {
		return fmt.Errorf("이미지 URL 조회 실패: %v", err)
	}

	handlers.BroadcastProgress("Downloading figma images", 1, fmt.Sprintf("%s 다운로드 중...", name))

	data, err := downloadImage(imgURL, *i.Token)
	if err != nil {
		return err
	}

	imgPath := filepath.Join(createdPath, fmt.Sprintf("%s.png", name))
	i.PathInfo[name] = imgPath
	if err := os.WriteFile(imgPath, data, 0666); err != nil {
		return fmt.Errorf("이미지 저장 실패 (%s): %v", name, err)
	}
	return nil
}

// GetFrames는 노드 트리에서 frameName과 일치하는 프레임의 자식들을 반환합니다.
func (i *Info) GetFrames(frameName string) []Node {
	var res []Node
	for _, node := range i.Nodes {
		switch node.Type {
		case "CANVAS":
			res = append(res, (&Info{Nodes: node.Children}).GetFrames(frameName)...)
		case "FRAME":
			if node.Name == frameName {
				i.FrameName = frameName
				res = append(res, (&Info{Nodes: node.Children}).GetFrames("children")...)
			} else if frameName == "children" {
				res = append(res, node)
			}
		}
	}
	if frameName != "children" {
		handlers.BroadcastProgress("Got frame", 1, fmt.Sprintf("'%s' 프레임 %d개 발견", frameName, len(res)))
	}
	return res
}
