package core

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	myEth "github.com/zy99978455-otw/flash-monitor/pkg/ethereum"
	"github.com/zy99978455-otw/flash-monitor/pkg/logger"
	"go.uber.org/zap"
)

// 定义 USDT 的 Transfer 事件签名
// Transfer(address indexed from, address indexed to, uint256 value)
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
	// 1. 构造查询条件 (FilterQuery)
	// 我们要找：USDT 合约地址 + Transfer 事件
	contractAddr := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7") // USDT 地址
	
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(startHeight),
		ToBlock:   new(big.Int).SetUint64(endHeight),
		Addresses: []common.Address{contractAddr},
		Topics: [][]common.Hash{
			{LogTransferSigHash}, // Topic[0] 必须是 Transfer 事件的 Hash
		},
	}

	// 2. 调用 RPC 接口获取日志
	logs, err := s.Client.EthClient.FilterLogs(ctx, query)
	if err != nil {
		logger.Log.Error("获取日志失败", 
			zap.Uint64("start", startHeight), 
			zap.Uint64("end", endHeight), 
			zap.Error(err),
		)
		return
	}

	// 3. 简单打印一下结果 (证明我们抓到了)
	logger.Log.Info("扫描完成", 
		zap.Uint64("start", startHeight),
		zap.Uint64("end", endHeight),
		zap.Int("抓取到的交易数", len(logs)),
	)

	for _, vLog := range logs {
		logger.Log.Debug("发现一笔转账",
			zap.String("TxHash", vLog.TxHash.Hex()),
			zap.Uint64("Block", vLog.BlockNumber),
		)
	}
}