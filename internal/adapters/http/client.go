package http

import (
	"broadcast_server/internal/domain"
	"github.com/gorilla/websocket"
)

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
	hub  *domain.Hub
}

func (c *wsClient) Send(msg []byte) bool {
	select {
	case c.send <- msg:
		return true
	default:
		return false
	}
}

func (c *wsClient) Close() {
	_ = c.conn.Close()
}
