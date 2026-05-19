package data

import (
	"context"
	"database/sql"
	"time"
)

type TransferEventModel struct {
	DB *sql.DB
}

// GetAll 抓取区块链上的事件日志
func (m TransferEventModel) GetAll(fromAddress string, limit int) ([]*TransferEvent, error) {
	query := `
		SELECT id, tx_hash, log_index, block_number, block_hash, from_address, to_address, amount, token_address, created_at
		FROM transfer_events
		WHERE ($1 = '' OR from_address = $1)
		ORDER BY block_number DESC, log_index DESC
		LIMIT $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, fromAddress, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 初始化为空切片而不是 nil，确保 JSON 序列化时变成 [] 而不是 null
	events := []*TransferEvent{}
	for rows.Next() {
		var event TransferEvent
		err := rows.Scan(
			&event.ID,
			&event.TxHash,
			&event.LogIndex,
			&event.BlockNumber,
			&event.BlockHash,
			&event.FromAddress,
			&event.ToAddress,
			&event.Amount,
			&event.TokenAddress,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// Insert 将抓取到的事件日志存入数据库
func (m TransferEventModel) Insert(event *TransferEvent) error {
	// 使用 ON CONFLICT DO NOTHING 极其重要！
	// 这样当以太坊出现微小回滚或重复扫描时，相同的 tx_hash + log_index 会被自动忽略，而不会导致程序崩溃
	query := `
		INSERT INTO transfer_events (tx_hash, log_index, block_number, block_hash, from_address, to_address, amount, token_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tx_hash, log_index) DO NOTHING`

	args := []interface{}{
		event.TxHash,
		event.LogIndex,
		event.BlockNumber,
		event.BlockHash,
		event.FromAddress,
		event.ToAddress,
		event.Amount,
		event.TokenAddress,
	}

	// 设置 3 秒超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}
