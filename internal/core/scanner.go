package core

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	
	"github.com/zy99978455-otw/flash-monitor/internal/model"
	
	// 别名引用
	myEth "github.com/zy99978455-otw/flash-monitor/pkg/ethereum"
	"github.com/zy99978455-otw/flash-monitor/pkg/logger"
	"go.uber.org/zap"
)

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

// ScanResult 封装扫描结果
type ScanResult struct {
	BlockNumber uint64
	Events      []*model.TransferEvent
}

// Scan 扫描并返回结果，不再直接入库
func (s *Scanner) Scan(ctx context.Context, startHeight, endHeight uint64) (*ScanResult, error) {
	contractAddr := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7") 
	
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(startHeight),
		ToBlock:   new(big.Int).SetUint64(endHeight),
		Addresses: []common.Address{contractAddr},
		Topics: [][]common.Hash{
			{LogTransferSigHash},
		},
	}

	logs, err := s.Client.EthClient.FilterLogs(ctx, query)
	if err != nil {
		logger.Log.Error("获取日志失败", zap.Error(err))
		return nil, err
	}

	// 内存解析
	var events []*model.TransferEvent
	for _, vLog := range logs {
		event := ParseTransferLog(vLog)
		if event != nil {
			events = append(events, event)
		}
	}

	// 返回结果
	return &ScanResult{
		BlockNumber: startHeight,
		Events:      events,
	}, nil
}