package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func (app *application) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.logger.Info("❌ WebSocket 升级失败: %v", err)
		return
	}

	app.hub.register <- conn

	// 保持连接，直到客户端断开或发生错误
	defer func() {
		app.hub.unregister <- conn
	}()

	// 这里的读循环必不可少，即使你暂时不处理客户端发来的消息
	// 它可以检测连接是否断开并处理 Ping/Pong 消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				app.logger.Info("❌ WebSocket 读取异常: %v", err)
			}
			break
		}
	}
}
