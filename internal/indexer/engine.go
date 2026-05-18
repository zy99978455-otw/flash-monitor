package indexer

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/zy99978455-otw/flash-monitor/internal/data" // 👈 确保这里是你的真实模块名

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Engine 抓取器的核心结构体
type Engine struct {
	client *ethclient.Client
	models data.Models
	logger *log.Logger
}

// NewEngine 初始化并返回一个新的抓取引擎
func NewEngine(rpcURL string, models data.Models, logger *log.Logger) (*Engine, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &Engine{client: client, models: models, logger: logger}, nil
}

// Start 启动后台抓取任务 (死循环轮询)
func (e *Engine) Start(ctx context.Context) {
	e.logger.Println("Starting web3 indexer Engine...")

	ticker := time.NewTicker(12 * time.Second) // 以太坊出块大概 12 秒
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.logger.Println("indexer engine gracefully shutting down...")
			return
		case <-ticker.C:
			if err := e.syncBlocks(ctx); err != nil {
				// ✅ 修复 Bug: 这里必须用 Printf 才能打印出 err 的真实内容
				e.logger.Printf("error syncing blocks: %v\n", err)
			}
		}
	}
}

// syncBlocks 负责拉取区块、解码事件并插入 PostgreSQL
func (e *Engine) syncBlocks(ctx context.Context) error {
	header, err := e.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}
	chainHeight := header.Number.Int64()

	var dbHeight int64 = 0
	latestTrace, err := e.models.BlockTraces.GetLatest()
	if err != nil {
		return err
	}

	if latestTrace != nil {
		dbHeight = latestTrace.BlockNumber
	} else {
		dbHeight = chainHeight - 5
	}

	if dbHeight < chainHeight {
		fromBlock := dbHeight + 1
		toBlock := chainHeight

		if toBlock-fromBlock > 2000 {
			toBlock = fromBlock + 2000
		}

		e.logger.Printf("fetching logs from block %d to %d (chain head: %d)", fromBlock, toBlock, chainHeight)

		// 定义 ERC20 Transfer 的签名 Hash
		transferSigHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

		// ✅ 新增：定义 USDT 的主网官方合约地址
		usdtAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")

		// ✅ 新增：把 usdtAddress 放进 Addresses 过滤条件里
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(fromBlock),
			ToBlock:   big.NewInt(toBlock),
			Addresses: []common.Address{usdtAddress}, // 👈 雷达锁定 USDT
			Topics: [][]common.Hash{
				{transferSigHash},
			},
		}

		logs, err := e.client.FilterLogs(ctx, query)
		if err != nil {
			return err
		}
		e.logger.Printf("found %d USDT Transfer logs in current batch", len(logs))

		for _, vLog := range logs {
			if len(vLog.Topics) != 3 {
				continue
			}

			fromAddr := common.HexToAddress(vLog.Topics[1].Hex()).Hex()
			toAddr := common.HexToAddress(vLog.Topics[2].Hex()).Hex()

			amount := new(big.Int)
			amount.SetBytes(vLog.Data)

			event := &data.TransferEvent{
				TxHash:       vLog.TxHash.Hex(),
				LogIndex:     int(vLog.Index),
				BlockNumber:  int64(vLog.BlockNumber),
				BlockHash:    vLog.BlockHash.Hex(),
				FromAddress:  fromAddr,
				ToAddress:    toAddr,
				Amount:       amount.String(),
				TokenAddress: vLog.Address.Hex(),
			}

			// ✅ 修复：不再静默忽略，显式捕获并打印插入错误
			insertErr := e.models.TransferEvents.Insert(event)
			if insertErr != nil {
				e.logger.Printf("❌ 插入交易数据失败 [Tx: %s]: %v\n", event.TxHash, insertErr)
			}
		}

		trace := &data.BlockTrace{
			BlockNumber: toBlock,
			BlockHash:   header.Hash().Hex(),
			ParentHash:  header.ParentHash.Hex(),
		}

		// ✅ 修复：捕获游标插入的错误
		traceErr := e.models.BlockTraces.Insert(trace)
		if traceErr != nil {
			e.logger.Printf("❌ 插入游标数据失败 [Block: %d]: %v\n", toBlock, traceErr)
		}
	}

	return nil
}
