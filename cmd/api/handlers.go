package main

import (
	"net/http"
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

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	// 准备要返回的数据
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     "1.0.0",
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

func (app *application) listTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// 从数据库中获取真实的转账记录 (默认取最近 20 条)
	events, err := app.models.TransferEvents.GetAll("", 20)
	if err != nil {
		app.logger.Printf("❌ 获取转账记录失败: %v", err)
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
		app.logger.Printf("❌ 写入 JSON 响应失败: %v", err)
	}
}
