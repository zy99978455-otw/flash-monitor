package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/zy99978455-otw/flash-monitor/internal/data"
)

// Broker 是 SSE 实时广播中心
type Broker struct {
	ctx            context.Context
	Broadcast      chan *data.TransferEvent
	newClients     chan chan *data.TransferEvent
	closingClients chan chan *data.TransferEvent
	clients        map[chan *data.TransferEvent]bool
	logger         *slog.Logger
}

func NewBroker(ctx context.Context, logger *slog.Logger) *Broker {
	return &Broker{
		ctx:            ctx,
		Broadcast:      make(chan *data.TransferEvent, 1),
		newClients:     make(chan chan *data.TransferEvent),
		closingClients: make(chan chan *data.TransferEvent),
		clients:        make(map[chan *data.TransferEvent]bool),
		logger:         logger,
	}
}

// Start 负责在后台无锁地管理所有客户端连接和消息分发
func (b *Broker) Start() {
	for {
		select {
		case s := <-b.newClients:
			b.clients[s] = true
			b.logger.Info("SSE client connected", "total_clients", len(b.clients))

		case s := <-b.closingClients:
			delete(b.clients, s)
			b.logger.Info("SSE client disconnected", "total_clients", len(b.clients))

		case event := <-b.Broadcast:
			for clientMessageChan := range b.clients {
				select {
				case clientMessageChan <- event:
				default:
					b.logger.Warn("dropping slow SSE client")
					delete(b.clients, clientMessageChan)
					close(clientMessageChan)
				}
			}
		}
	}
}

// Handler 是暴露给前端的 HTTP 接口
func (b *Broker) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan *data.TransferEvent, 5000)
	b.newClients <- messageChan

	defer func() {
		b.closingClients <- messageChan
	}()

	ctx := r.Context()

	// 一个每隔 10 秒跳动一次的心跳定时器
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-b.ctx.Done():
			// 🛑 核心新增：系统发出 Ctrl+C 停机指令，主动切断这个长连接！
			return

		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case event := <-messageChan:
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", payload)

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}
