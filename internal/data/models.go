package data

import (
	"database/sql"
	"time"
)

// BlockTrace 代表 区块扫描轨迹
// BlockNumber 处理 ·断点续传·
// BlockHash和ParentHash 处理 ·分叉与回滚·
type BlockTrace struct {
	ID          int64     `json:"id"`
	BlockNumber int64     `json:"block_number"`
	BlockHash   string    `json:"block_hash"`
	ParentHash  string    `json:"parent_hash"` // 上一个区块的哈希值
	ScanTime    time.Time `json:"scan_time"`   //处理完这个区块的现实时间
}

// TransferEvent 代表以太坊上的转账事件
type TransferEvent struct {
	ID           int64     `json:"id"`
	TxHash       string    `json:"tx_hash"`
	LogIndex     int       `json:"log_index"`
	BlockNumber  int64     `json:"block_number"`
	BlockHash    string    `json:"block_hash"`
	FromAddress  string    `json:"from_address"`
	ToAddress    string    `json:"to_address"`
	Amount       string    `json:"amount"` // 使用 string 防止前端和 Go 处理超大金额时精度丢失
	TokenAddress string    `json:"token_address"`
	CreatedAt    time.Time `json:"created_at"`
}

type Models struct {
	BlockTraces    BlockTraceModel
	TransferEvents TransferEventModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		BlockTraces:    BlockTraceModel{DB: db},
		TransferEvents: TransferEventModel{DB: db},
	}
}
