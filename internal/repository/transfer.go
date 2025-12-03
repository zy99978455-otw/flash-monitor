package repository

import (
	"github.com/zy99978455-otw/flash-monitor/internal/model"
	"gorm.io/gorm/clause"
)

// SaveTransferEvent 保存转账事件
// 核心逻辑：使用 "INSERT IGNORE" 模式，如果数据已存在（根据 TxHash+LogIndex 唯一索引），则忽略，防止报错
func SaveTransferEvent(event *model.TransferEvent) error {
	result := DB.Clauses(clause.OnConflict{
		DoNothing: true, // 遇到冲突(重复数据)什么都不做
	}).Create(event)
	
	return result.Error
}