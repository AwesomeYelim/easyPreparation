package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
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

	// Keep the connection open
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			fmt.Println("WebSocket client disconnected")
			break
		}
	}
}

// 서버가 클라이언트들에게 완료 알림 보내기
func BroadcastProcessDone(target, fileName string) {
	message := map[string]interface{}{
		"type":     "done",
		"target":   target,
		"fileName": fileName,
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
