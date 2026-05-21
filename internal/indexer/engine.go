package indexer

import (
	"context"
	"log/slog"
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
	logger *slog.Logger
	events chan *data.TransferEvent
}

// NewEngine 初始化并返回一个新的抓取引擎
func NewEngine(rpcURL string, models data.Models, logger *slog.Logger, events chan *data.TransferEvent) (*Engine, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &Engine{
		client: client,
		models: models,
		logger: logger,
		events: events,
	}, nil
}

// Start 启动后台抓取任务 (死循环轮询)
func (e *Engine) Start(ctx context.Context) {
	e.logger.Info("Starting web3 indexer Engine...")

	ticker := time.NewTicker(12 * time.Second) // 以太坊出块大概 12 秒
	defer ticker.Stop()

	for {
		// 监听两个通道
		select {
		case <-ctx.Done(): //外部取消信号，优雅停机
			e.logger.Info("indexer engine gracefully shutting down...")
			return
		case <-ticker.C: //定时器触发信号
			if err := e.syncBlocks(ctx); err != nil {
				// 这里必须用 Printf 才能打印出 err 的真实内容
				e.logger.Info("error syncing blocks: %v\n", err)
			}
		}
	}
}

// syncBlocks 是抓取引擎的核心同步处理器。
// 它负责执行一个完整的“提取-转换-加载 (ETL)”工作流：
// 1. 水位探测：比对本地数据库游标与主网最新区块高度，处理冷启动与落后追赶。
// 2. 安全抓取：采用最大 2000 区块的批量限制，防止 Infura 节点请求过载。
// 3. 精准过滤：使用 FilterQuery 直接在链上锁定 USDT 合约的 Transfer 事件。
// 4. 数据落盘：将清洗后的交易数据和最新进度游标事务性地写入 PostgreSQL。
//
// 如果在同步过程中遇到网络错误，将返回 error 交由主循环重试；
// 对于单条数据的插入失败，采取记录日志并跳过的策略，以保障整体引擎不宕机。
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
		dbHeight = chainHeight - 5 //避免链重组织（reorg）导致数据错误
	}

	if dbHeight < chainHeight {
		fromBlock := dbHeight + 1 //下一个未同步块
		toBlock := chainHeight    //当前链高度

		// USDT 太活跃了，50 个块可能超过 10,000 条记录（Infura 的限制）
		// 我们把步长缩小到 5 个块，确保请求不会过大
		if toBlock-fromBlock > 5 {
			toBlock = fromBlock + 5
		}

		e.logger.Info("fetching logs from block %d to %d (chain head: %d)", fromBlock, toBlock, chainHeight)

		// 定义 ERC20 Transfer 的签名 Hash
		transferSigHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

		// 定义 USDT 的主网官方合约地址
		usdtAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")

		// 把 usdtAddress 放进 Addresses 过滤条件里
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
		e.logger.Info("found %d USDT Transfer logs in current batch", len(logs))

		// 遍历事件并解析
		for _, vLog := range logs {
			if len(vLog.Topics) != 3 {
				continue
			}

			/*
				ERC20 Transfer 事件：
					topics[0] → event signature
					topics[1] → from
					topics[2] → to
					data → amount
			*/
			fromAddr := common.HexToAddress(vLog.Topics[1].Hex()).Hex()
			toAddr := common.HexToAddress(vLog.Topics[2].Hex()).Hex()

			amount := new(big.Int)
			amount.SetBytes(vLog.Data)

			// 🐳 巨鲸过滤：只关注大于 50,000 USDT 的交易 (USDT 有 6 位小数)
			minAmount := new(big.Int).Mul(big.NewInt(50000), big.NewInt(1000000))
			if amount.Cmp(minAmount) < 0 {
				continue
			}

			// 封装事件并写入数据库
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
				e.logger.Info("❌ 插入交易数据失败 [Tx: %s]: %v\n", event.TxHash, insertErr)
			} else {
				// 只有成功插入（或 ON CONFLICT 被忽略但我们也视作成功逻辑）的数据才推送到实时通道
				// 注意：这里为了简单直接发送。如果是 ON CONFLICT 忽略的情况，可能不需要重复推。
				// 但由于 Indexer 逻辑，新抓到的通常都是新数据。
				if e.events != nil {
					e.events <- event
				}
			}
		}

		// 更新区块游标
		trace := &data.BlockTrace{
			BlockNumber: toBlock,
			BlockHash:   header.Hash().Hex(),
			ParentHash:  header.ParentHash.Hex(),
		}

		traceErr := e.models.BlockTraces.Insert(trace)
		if traceErr != nil {
			e.logger.Info("❌ 插入游标数据失败 [Block: %d]: %v\n", toBlock, traceErr)
		}
	}

	return nil
}
