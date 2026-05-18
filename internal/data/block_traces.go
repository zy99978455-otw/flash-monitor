package data

import (
	"context"
	"database/sql"
	"time"
)

type BlockTraceModel struct {
	DB *sql.DB
}

// Insert 插入新的游标
func (m BlockTraceModel) Insert(trace *BlockTrace) error {
	query := `
		INSERT INTO block_traces (block_number, block_hash, parent_hash, scan_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	args := []interface{}{trace.BlockNumber, trace.BlockHash, trace.ParentHash, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&trace.ID)
}

// GetLatest 获取最新的抓取游标
func (m BlockTraceModel) GetLatest() (*BlockTrace, error) {
	query := `
		SELECT id, block_number, block_hash, parent_hash, scan_time 
		FROM block_traces 
		ORDER BY block_number DESC 
		LIMIT 1`

	var trace BlockTrace
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query).Scan(
		&trace.ID,
		&trace.BlockNumber,
		&trace.BlockHash,
		&trace.ParentHash,
		&trace.ScanTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 还没抓取过，返回空
		}
		return nil, err
	}
	return &trace, nil
}
