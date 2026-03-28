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
			if msgType, _ := data["type"].(string); msgType == "position" {
				if idxFloat, ok := data["idx"].(float64); ok {
					newIdx := int(idxFloat)
					UpdateDisplayIdx(newIdx)
					BroadcastMessage("position", map[string]interface{}{"idx": newIdx})
					// OBS 씬 전환
					if title := GetCurrentTitle(); title != "" {
						go obs.Get().SwitchScene(title)
					}
				}
			}
		}
	}
}

func BroadcastMessage(messageType string, payload map[string]interface{}) {
	message := map[string]interface{}{
		"type": messageType,
	}
	for k, v := range payload {
		message[k] = v
	}

	msgBytes, _ := json.Marshal(message)

	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, msgBytes)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
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
