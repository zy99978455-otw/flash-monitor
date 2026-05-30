package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zy99978455-otw/flash-monitor/internal/validator"
)

type TransferEventModel struct {
	DB *sql.DB
}

func ValidateTransferEvent(v *validator.Validator, event *TransferEvent) {
	// 检验以太坊地址格式
	v.Check(validator.IsEthAddress(event.FromAddress), "from_address", "must be a valid hex-encoded Ethereum address")
	v.Check(validator.IsEthAddress(event.ToAddress), "to_address", "must be a valid hex-encoded Ethereum address")

	// 校验金额（不能为空且必须是正数数字）
	v.Check(event.Amount != "", "amount", "must be provided")

	// 校验交易哈希（长度必须是66位，0x 开头）
	v.Check(len(event.TxHash) == 66, "tx_hash", "must be a valid transaction hash")

	// 校验区块高度
	v.Check(event.BlockNumber > 0, "block_number", "must be a positive integer")
}

// GetAll 抓取区块链上的事件日志 (增强版：支持分页、过滤、排序)
func (m TransferEventModel) GetAll(fromAddress, toAddress string, filters Filters) ([]*TransferEvent, Metadata, error) {
	// 使用 count(*) OVER() 同时获取总行数
	// 使用 fmt.Sprintf 注入排序列和方向
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, tx_hash, log_index, block_number, block_hash, from_address, to_address, amount, token_address, created_at
		FROM transfer_events
		WHERE ($1 = '' OR from_address = $1)
		AND ($2 = '' OR to_address = $2)
		ORDER BY %s %s, log_index DESC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{fromAddress, toAddress, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	events := []*TransferEvent{}

	for rows.Next() {
		var event TransferEvent
		err := rows.Scan(
			&totalRecords,
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
			return nil, Metadata{}, err
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// 计算分页元数据
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return events, metadata, nil
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

func (m TransferEventModel) InsertTx(ctx context.Context, tx *sql.Tx, event *TransferEvent) error {
	query := `
		INSERT INTO transfer_events (tx_hash, log_index, block_number, block_hash, from_address, to_address, amount, token_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tx_hash, log_index) DO NOTHING`

	args := []any{
		event.TxHash,
		event.LogIndex,
		event.BlockNumber,
		event.BlockHash,
		event.FromAddress,
		event.ToAddress,
		event.Amount,
		event.TokenAddress,
	}

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}
