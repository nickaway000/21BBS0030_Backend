package utils

import (
    "log"
    "github.com/gorilla/websocket"
    "net/http"
)


var WebSocketConnections = make([]*websocket.Conn, 0)


var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}


func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    WebSocketConnections = append(WebSocketConnections, conn)
}


func SendWebSocketMessage(message string) {
    for _, conn := range WebSocketConnections {
        if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
            log.Printf("WebSocket send error: %v", err)
            conn.Close()
        }
    }
}
