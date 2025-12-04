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

// DeleteTransferEventsByBlock 删除指定区块高度的所有交易 (用于回滚)
func DeleteTransferEventsByBlock(blockNumber uint64) error {
	return DB.Where("block_number = ?", blockNumber).Delete(&model.TransferEvent{}).Error
}

// SaveTransferEventsBatch 批量保存转账事件 (性能优化版)
func SaveTransferEventsBatch(events []*model.TransferEvent) error {
	if len(events) == 0 {
		return nil
	}
	// 使用 OnConflict 做去重，防止批量插入时某一条重复导致整体失败
	// 每次插入 100 条，防止 SQL 语句过长
	return DB.Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(events, 100).Error
}