package http

import (
	"log"
	"net/http"
	"time"

	"context"
	"github.com/gorilla/websocket"

	"broadcast_server/internal/domain"
)

func ServeWS(ctx context.Context, hub *domain.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}
	sender := resolveSender(r)
	remote := shortRemote(r)

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, sendBufSize),
		hub:  hub,
	}

	hub.Register() <- client
	log.Printf("joined: sender=%q remote=%s", sender, remote)

	// reader
	go func() {
		defer func() {
			hub.Unregister() <- client
			log.Printf("left: sender=%q remote=%s", sender, remote)
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
			log.Printf("message: sender=%q remote=%s bytes=%d", sender, remote, len(message))
			out := append([]byte("["+sender+"] "), message...)
			hub.Broadcast() <- out
		}
	}()

	// writer
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
				err = client.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Printf("ws write error: sender=%q remote=%s err=%v", sender, remote, err)
					return
				}
			case <-ticker.C:
				// ping для keep-alive
				_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err = client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("ws ping error: sender=%q remote=%s err=%v", sender, remote, err)
					return
				}
			}
		}
	}()
}
