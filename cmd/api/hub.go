package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/zy99978455-otw/flash-monitor/internal/data"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许跨域
	},
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan *data.TransferEvent
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	logger     *slog.Logger
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan *data.TransferEvent, 256), // 增加 256 个事件的缓冲区
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		logger:     logger,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.Info("New WebSocket client connected", "active_clients", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				h.logger.Info("WebSocket client disconnected", "active_clients", len(h.clients))

			}

		case event := <-h.broadcast:
			for client := range h.clients {
				err := client.WriteJSON(event)
				if err != nil {
					log.Printf("WebSocket write error, kicking client: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
		}
	}
}
