package handlers

import (
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
