package http

import (
	"broadcast_server/internal/domain"
	"broadcast_server/internal/events"
	"context"
	"github.com/gorilla/websocket"
	"log"
	stdhttp "net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = 20 * time.Second
	maxMessageSize = 4096
	sendBufSize    = 64
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *stdhttp.Request) bool {
		return true
	},
}

func ServeWS(ctx context.Context, hub *domain.Hub, evBus *events.Subject[events.Event], w stdhttp.ResponseWriter, r *stdhttp.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return
	}
	sender := r.URL.Query().Get("sender")

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, sendBufSize),
		hub:  hub,
	}

	hub.Register() <- client
	if evBus != nil {
		evBus.Publish(ctx, events.Event{
			Type:   events.EventJoined,
			Sender: sender,
			At:     time.Now(),
		})
	}

	go func() {
		defer func() {
			hub.Unregister() <- client
			if evBus != nil {
				evBus.Publish(ctx, events.Event{
					Type:   events.EventLeft,
					Sender: sender,
					At:     time.Now(),
				})
			}
		}()

		conn.SetReadLimit(maxMessageSize)
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if mt != websocket.TextMessage {
				continue
			}

			hub.Broadcast() <- message

			if evBus != nil {
				evBus.Publish(ctx, events.Event{
					Type:    events.EventBroadcast,
					Sender:  sender,
					Payload: message,
					At:      time.Now(),
				})
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-client.send:
				_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}

			case <-ticker.C:
				_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()
}
