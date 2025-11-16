package websocket

import (
	"github.com/gorilla/websocket"
)

// WebSocketConnection WebSocket连接适配器
type WebSocketConnection struct {
	id string
	*websocket.Conn
}

// NewWebSocketConnection 创建新的WebSocket连接适配器
func NewWebSocketConnection(id string, conn *websocket.Conn) *WebSocketConnection {
	return &WebSocketConnection{
		id:   id,
		Conn: conn,
	}
}
