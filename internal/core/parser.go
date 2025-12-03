package core

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zy99978455-otw/flash-monitor/internal/model"
)

// ParseTransferLog 将链上原始日志解析为数据库模型
// 核心逻辑：从 Log 的 Topics 和 Data 中提取 From, To, Amount
func ParseTransferLog(log types.Log) *model.TransferEvent {
	// 1. 基础校验：ERC20 Transfer 事件必须有 3 个 Topic (Sig, From, To)
	if len(log.Topics) < 3 {
		return nil
	}

	// 2. 解析 From 和 To
	// Topic 是 32 字节的 Hash，而地址是 20 字节，所以需要 HexToAddress 截取后 20 字节
	from := common.HexToAddress(log.Topics[1].Hex()).Hex()
	to := common.HexToAddress(log.Topics[2].Hex()).Hex()

	// 3. 解析金额 (Amount)
	// Amount 存储在 Data 字段中，是 BigInt 的字节流
	amount := new(big.Int).SetBytes(log.Data)

	// 4. 组装成数据库模型
	event := &model.TransferEvent{
		TxHash:       log.TxHash.Hex(),
		LogIndex:     log.Index,
		BlockNumber:  log.BlockNumber,
		BlockHash:    log.BlockHash.Hex(),
		FromAddress:  strings.ToLower(from), // 转小写，方便后续查询
		ToAddress:    strings.ToLower(to),
		Amount:       amount.String(),       // 存字符串，防止数字太大溢出
		TokenAddress: strings.ToLower(log.Address.Hex()),
	}

	return event
}