package main

import (
	"log"
	"net/http"
	"sync"

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
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan *data.TransferEvent, 256), // 增加 256 个事件的缓冲区
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Println("New WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			log.Println("WebSocket client disconnected")

		case event := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteJSON(event)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}
