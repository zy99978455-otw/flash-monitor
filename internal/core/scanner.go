package core

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	
	"github.com/zy99978455-otw/flash-monitor/internal/model"
	"github.com/zy99978455-otw/flash-monitor/internal/repository"
	
	//以此别名引用自定义包，防止与官方包名冲突
	myEth "github.com/zy99978455-otw/flash-monitor/pkg/ethereum"
	"github.com/zy99978455-otw/flash-monitor/pkg/logger"
	"go.uber.org/zap"
)

// 定义 USDT 的 Transfer 事件签名
var (
	LogTransferSig     = []byte("Transfer(address,address,uint256)")
	LogTransferSigHash = crypto.Keccak256Hash(LogTransferSig)
)

type Scanner struct {
	Client *myEth.Client
}

func NewScanner(client *myEth.Client) *Scanner {
	return &Scanner{Client: client}
}

// Scan 扫描指定区间的区块
func (s *Scanner) Scan(ctx context.Context, startHeight, endHeight uint64) {
	// 1. 构造查询条件
	contractAddr := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7") // USDT 合约
	
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(startHeight),
		ToBlock:   new(big.Int).SetUint64(endHeight),
		Addresses: []common.Address{contractAddr},
		Topics: [][]common.Hash{
			{LogTransferSigHash}, // 只过滤 Transfer 事件
		},
	}

	// 2. 调用 RPC 接口获取原始日志
	logs, err := s.Client.EthClient.FilterLogs(ctx, query)
	if err != nil {
		logger.Log.Error("获取日志失败", 
			zap.Uint64("start", startHeight), 
			zap.Uint64("end", endHeight), 
			zap.Error(err),
		)
		return
	}

	logger.Log.Info("扫描完成", 
		zap.Uint64("start", startHeight),
		zap.Uint64("end", endHeight),
		zap.Int("raw_logs", len(logs)),
	)

	// 3. 内存缓冲：解析所有日志
	// 使用切片暂存，准备批量插入，减少数据库 IO 次数
	var events []*model.TransferEvent

	for _, vLog := range logs {
		// 调用 parser.go 里的解析逻辑
		event := ParseTransferLog(vLog)
		if event != nil {
			events = append(events, event)
		}
	}

	// 4. 批量入库
	if len(events) > 0 {
		// 调用 repository 的批量插入方法
		err := repository.SaveTransferEventsBatch(events)
		if err != nil {
			logger.Log.Error("批量入库失败", zap.Error(err))
			// 实际生产中，这里可能需要重试机制
			return 
		}

		logger.Log.Info("🚀 批量入库成功", 
			zap.Int("条数", len(events)), 
			zap.Uint64("区块", startHeight),
		)
		
		// 可选：打印第一条交易用于调试，证明数据是对的
		logger.Log.Debug("示例交易", zap.String("Tx", events[0].TxHash))

	} else {
		logger.Log.Info("📭 该区块无有效转账", zap.Uint64("区块", startHeight))
	}
}