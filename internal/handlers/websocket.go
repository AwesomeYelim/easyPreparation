package handlers

import (
	"easyPreparation_1.0/internal/obs"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // CORS 허용
}

// 커넥션별 write mutex — gorilla/websocket은 동시 쓰기 불가
type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

var clients = make(map[*websocket.Conn]*wsClient)
var clientsMu sync.Mutex

func addClient(conn *websocket.Conn) *wsClient {
	c := &wsClient{conn: conn}
	clientsMu.Lock()
	clients[conn] = c
	clientsMu.Unlock()
	return c
}

func removeClient(conn *websocket.Conn) {
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}

// safeWrite — 커넥션별 mutex로 동시 쓰기 방지
func (c *wsClient) safeWrite(msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

// WebSocket 핸들러
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	wsc := addClient(conn)
	fmt.Println("WebSocket client connected")

	// 새 클라이언트에게 현재 order + idx + churchName 전송
	orderMu.RLock()
	if len(currentOrder) > 0 {
		msg, _ := json.Marshal(map[string]interface{}{
			"type":       "order",
			"items":      currentOrder,
			"idx":        currentIdx,
			"churchName": displayChurchName,
		})
		_ = wsc.safeWrite(msg)
	}
	orderMu.RUnlock()

	// OBS 씬 전환은 항목(idx)이 바뀔 때만 수행 (서브페이지 이동 시 스킵)
	lastObsIdx := -1

	// Keep the connection open + handle incoming messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			removeClient(conn)
			fmt.Println("WebSocket client disconnected")
			break
		}
		// Display HTML이 보고하는 position 처리
		var data map[string]interface{}
		if json.Unmarshal(msg, &data) == nil {
			msgType, _ := data["type"].(string)
			if msgType == "position" {
				if idxFloat, ok := data["idx"].(float64); ok {
					newIdx := int(idxFloat)
					UpdateDisplayIdx(newIdx)
					posPayload := map[string]interface{}{"idx": newIdx}
					subPage := 0
					if sp, ok := data["subPageIdx"].(float64); ok {
						subPage = int(sp)
						posPayload["subPageIdx"] = subPage
					}
					// Display 본인 제외, 제어판에만 전달
					BroadcastMessageExcept("position", posPayload, conn)
					// OBS 씬 전환 (항목이 바뀔 때만)
					if newIdx != lastObsIdx {
						lastObsIdx = newIdx
						if title := GetCurrentTitle(); title != "" {
							go obs.Get().SwitchScene(title)
						}
					}
					// 서버 타이머 업데이트
					OnPositionUpdate(newIdx, subPage)
				}
			}
		}
	}
}

// broadcastTo — 공통 브로드캐스트 (except가 nil이면 전체 전송)
func broadcastTo(msgBytes []byte, except *websocket.Conn) {
	clientsMu.Lock()
	snapshot := make([]*wsClient, 0, len(clients))
	for _, c := range clients {
		snapshot = append(snapshot, c)
	}
	clientsMu.Unlock()

	var failed []*websocket.Conn
	for _, c := range snapshot {
		if c.conn == except {
			continue
		}
		if err := c.safeWrite(msgBytes); err != nil {
			c.conn.Close()
			failed = append(failed, c.conn)
		}
	}

	if len(failed) > 0 {
		clientsMu.Lock()
		for _, conn := range failed {
			delete(clients, conn)
		}
		clientsMu.Unlock()
	}
}

// 특정 conn 제외하고 브로드캐스트 (메시지 발신자에게 다시 보내지 않기 위해)
func BroadcastMessageExcept(messageType string, payload map[string]interface{}, except *websocket.Conn) {
	message := map[string]interface{}{
		"type": messageType,
	}
	for k, v := range payload {
		message[k] = v
	}
	msgBytes, _ := json.Marshal(message)
	broadcastTo(msgBytes, except)
}

func BroadcastMessage(messageType string, payload map[string]interface{}) {
	message := map[string]interface{}{
		"type": messageType,
	}
	for k, v := range payload {
		message[k] = v
	}
	if messageType == "navigate" {
		log.Printf("[broadcast] navigate to %d clients", len(clients))
	}
	msgBytes, _ := json.Marshal(message)
	broadcastTo(msgBytes, nil)
}

func StartKeepAliveBroadcast() {
	ticker := time.NewTicker(50 * time.Second)

	go func() {
		for range ticker.C {
			BroadcastMessage("keepalive", map[string]interface{}{})
		}
	}()
}

func BroadcastProcessDone(target, fileName string) {
	BroadcastMessage("done", map[string]interface{}{
		"target":   target,
		"fileName": fileName,
	})
}

func BroadcastProgress(target string, code int, message string) {
	BroadcastMessage("progress", map[string]interface{}{
		"target":  target,
		"code":    code,
		"message": message,
	})
	log.Print(message)
}
