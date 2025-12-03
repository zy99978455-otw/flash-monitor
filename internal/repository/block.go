package repository

import (
	"errors"

	"github.com/zy99978455-otw/flash-monitor/internal/model"
	"gorm.io/gorm"
)

// GetLastScannedBlock 获取数据库中记录的最新区块高度
func GetLastScannedBlock() (uint64, error) {
	var trace model.BlockTrace
	// 按高度倒序查第一条
	err := DB.Order("block_number desc").First(&trace).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil // 数据库是空的，说明是第一次运行
	}
	
	return trace.BlockNumber, err
}

// SaveBlockTrace 保存区块扫描记录
func SaveBlockTrace(trace *model.BlockTrace) error {
	return DB.Create(trace).Error
}