package model

import (
	"time"
)

// BlockTrace 区块足迹表 (用于防回滚)
type BlockTrace struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	BlockNumber uint64    `gorm:"uniqueIndex;not null" json:"block_number"`
	BlockHash   string    `gorm:"type:varchar(66);not null" json:"block_hash"`
	ParentHash  string    `gorm:"type:varchar(66);not null" json:"parent_hash"`
	ScanTime    time.Time `gorm:"autoCreateTime" json:"scan_time"`
}

// TransferEvent 转账事件表 (业务数据)
type TransferEvent struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TxHash          string    `gorm:"type:varchar(66);uniqueIndex:idx_tx_log;not null" json:"tx_hash"`
	LogIndex        uint      `gorm:"uniqueIndex:idx_tx_log;not null" json:"log_index"`
	BlockNumber     uint64    `gorm:"index;not null" json:"block_number"`
	BlockHash       string    `gorm:"type:varchar(66);not null" json:"block_hash"`
	FromAddress     string    `gorm:"type:varchar(42);index;not null" json:"from_address"`
	ToAddress       string    `gorm:"type:varchar(42);index;not null" json:"to_address"`
	Amount          string    `gorm:"type:varchar(78);not null" json:"amount"`
	TokenAddress    string    `gorm:"type:varchar(42);not null" json:"token_address"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (BlockTrace) TableName() string {
	return "block_trace"
}

func (TransferEvent) TableName() string {
	return "transfer_event"
}