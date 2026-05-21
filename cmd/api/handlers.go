package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// 定义交易响应
type TransactionResponse struct {
	TxHash      string `json:"tx_hash"`
	BlockNumber int64  `json:"block_number"`
	From        string `json:"from_address"`
	To          string `json:"to_address"`
	Amount      string `json:"amount"` // 建议用 string 传给前端，防止 JS 精度丢失
	Time        string `json:"timestamp"`
}

func (app *application) listTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// 从数据库中获取真实的转账记录 (默认取最近 20 条)
	events, err := app.models.TransferEvents.GetAll("", 20)
	if err != nil {
		app.logger.Info("❌ 获取转账记录失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 构造返回数据
	env := envelope{
		"status": "success",
		"data":   events,
	}

	// 使用 writeJSON 辅助函数返回
	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logger.Info("❌ 写入 JSON 响应失败: %v", err)
	}
}

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
