package data

import (
	"context"
	"database/sql"
	"time"
)

type BlockTraceModel struct {
	DB *sql.DB
}

// Insert 负责在数据库中持久化当前区块的扫描轨迹（游标）。
// 它是实现引擎“断点续传”和“防区块链分叉回滚”的核心元数据记录器。
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

func (m BlockTraceModel) InsertTx(ctx context.Context, tx *sql.Tx, trace *BlockTrace) error {
	query := `
		INSERT INTO block_traces (block_number, block_hash, parent_hash)
       	VALUES ($1, $2, $3)
       	ON CONFLICT (block_number) DO NOTHING`

	args := []any{
		trace.BlockNumber, trace.BlockHash, trace.ParentHash,
	}

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

// GetLatest 获取数据库中记录的最新区块扫描轨迹。
// 它是抓取引擎重启时“读取存档”的关键方法。
// 如果数据库为空（首次启动），将安全地返回 (nil, nil) 而不是报错。
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
