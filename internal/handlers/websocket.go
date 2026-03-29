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

var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

// WebSocket 핸들러
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	fmt.Println("WebSocket client connected")

	// 새 클라이언트에게 현재 order + idx 전송 (display 창이 늦게 연결되어도 동작)
	orderMu.RLock()
	if len(currentOrder) > 0 {
		msg, _ := json.Marshal(map[string]interface{}{
			"type":  "order",
			"items": currentOrder,
			"idx":   currentIdx,
		})
		_ = conn.WriteMessage(websocket.TextMessage, msg)
	}
	orderMu.RUnlock()

	// OBS 씬 전환은 항목(idx)이 바뀔 때만 수행 (서브페이지 이동 시 스킵)
	lastObsIdx := -1

	// Keep the connection open + handle incoming messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
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
	// 클라이언트 목록 복사 → lock 잡은 채 I/O 하지 않기 위해
	snapshot := make([]*websocket.Conn, 0, len(clients))
	for c := range clients {
		snapshot = append(snapshot, c)
	}
	clientsMu.Unlock()

	var failed []*websocket.Conn
	for _, c := range snapshot {
		if c == except {
			continue
		}
		_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := c.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			c.Close()
			failed = append(failed, c)
		}
	}

	if len(failed) > 0 {
		clientsMu.Lock()
		for _, c := range failed {
			delete(clients, c)
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
	log.Printf(message)

}
