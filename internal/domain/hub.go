package domain

import (
	"context"
)

type Client interface {
	Send(msg []byte) bool // неблокирующая попытка отправки; false — если переполнен буфер
	Close()
}

type Hub struct {
	register   chan Client
	unregister chan Client
	broadcast  chan []byte
	clients    map[Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan Client),
		unregister: make(chan Client),
		broadcast:  make(chan []byte, 1024),
		clients:    make(map[Client]struct{}),
	}
}

func (h *Hub) Register() chan<- Client   { return h.register }
func (h *Hub) Unregister() chan<- Client { return h.unregister }
func (h *Hub) Broadcast() chan<- []byte  { return h.broadcast }

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// закрываем всех клиентов
			for c := range h.clients {
				c.Close()
			}
			return

		case c := <-h.register:
			h.clients[c] = struct{}{}

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				c.Close()
			}

		case msg := <-h.broadcast:
			for c := range h.clients {
				if !c.Send(msg) {
					// медленный клиент — отключаем
					delete(h.clients, c)
					c.Close()
				}
			}
		}
	}
}
