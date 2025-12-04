package main

import (
	"context"
	"time"

	"github.com/zy99978455-otw/flash-monitor/internal/config"
	"github.com/zy99978455-otw/flash-monitor/internal/core"
	"github.com/zy99978455-otw/flash-monitor/internal/model"
	"github.com/zy99978455-otw/flash-monitor/internal/repository"
	"github.com/zy99978455-otw/flash-monitor/pkg/ethereum"
	"github.com/zy99978455-otw/flash-monitor/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	config.InitConfig()
	logger.InitLogger()
	repository.InitDB()

	client, err := ethereum.InitClient(config.AppConfig.Chain.RpcUrl)
	if err != nil {
		logger.Log.Fatal("节点连接失败", zap.Error(err))
	}
	scanner := core.NewScanner(client)

	// 1. 确定启动高度
	lastBlock, err := repository.GetLastScannedBlock()
	var currentScanBlock uint64
	if err != nil || lastBlock == 0 {
		onChainCurrent, _ := client.GetBlockNumber(context.Background())
		currentScanBlock = onChainCurrent - 50
		logger.Log.Info("✨ 首次启动，从回溯高度开始", zap.Uint64("start", currentScanBlock))
	} else {
		currentScanBlock = lastBlock + 1
		logger.Log.Info("🔄 断点续传", zap.Uint64("start", currentScanBlock))
	}

	// 2. 智能循环
	for {
		// A. 拿到链上最新高度
		latestBlock, err := client.GetBlockNumber(context.Background())
		if err != nil {
			logger.Log.Error("获取链上高度失败", zap.Error(err))
			time.Sleep(3 * time.Second)
			continue
		}

		// B. 如果追上了，就休息
		if currentScanBlock > latestBlock {
			logger.Log.Debug("等待新区块...", zap.Uint64("target", currentScanBlock), zap.Uint64("latest", latestBlock))
			time.Sleep(3 * time.Second)
			continue
		}

		// C. 【核心逻辑】回滚检测 (Reorg Check)
		// 我们准备扫 currentScanBlock。先获取它的区块头信息。
		header, err := client.GetBlockHeader(context.Background(), currentScanBlock)
		if err != nil {
			logger.Log.Error("获取区块头失败", zap.Uint64("height", currentScanBlock), zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		// 只有当我们不是从 0 开始，且数据库里有上一个块的记录时，才需要检查
		// 比如：准备扫 101，我们要检查 101.ParentHash 是否等于 DB 里的 100.Hash
		if currentScanBlock > 0 {
			prevBlockNum := currentScanBlock - 1
			dbBlockTrace, err := repository.GetBlockTraceByNumber(prevBlockNum)
			
			// 如果数据库里有上一个块的记录，进行比对
			if err == nil {
				// 链上 101 的 ParentHash
				parentHashOnChain := header.ParentHash.Hex()
				// 库里 100 的 Hash
				hashInDB := dbBlockTrace.BlockHash

				if parentHashOnChain != hashInDB {
					// 🚨 触发回滚！！！
					logger.Log.Warn("🚨 检测到区块回滚 (Reorg Detected) !!!", 
						zap.Uint64("回滚高度", prevBlockNum),
						zap.String("DB Hash", hashInDB),
						zap.String("Chain Parent", parentHashOnChain),
					)

					// 1. 删除 DB 中上一个块(100) 的 Trace
					repository.DeleteBlockTrace(prevBlockNum)
					// 2. 删除 DB 中上一个块(100) 的 交易记录
					repository.DeleteTransferEventsByBlock(prevBlockNum)
					
					// 3. 指针倒退，重新去扫 100
					currentScanBlock = prevBlockNum
					logger.Log.Warn("🔙 指针已回退，准备重扫", zap.Uint64("new_target", currentScanBlock))
					continue // 跳过本次循环，重新开始
				}
			}
		}

		// D. 正常扫描逻辑 (如果没有回滚，或者回滚修复后)
		logger.Log.Info("正在扫描", zap.Uint64("height", currentScanBlock))
		
		// 这里的 Scan 只需要扫这一个块
		scanner.Scan(context.Background(), currentScanBlock, currentScanBlock)

		// E. 存档 (保存真实的 Hash)
		err = repository.SaveBlockTrace(&model.BlockTrace{
			BlockNumber: header.Number.Uint64(),
			BlockHash:   header.Hash().Hex(),       // ✅ 存真实的 Hash
			ParentHash:  header.ParentHash.Hex(),   // ✅ 存真实的 ParentHash
		})
		
		if err != nil {
			logger.Log.Error("存档失败", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		logger.Log.Info("💾 进度已保存", zap.Uint64("height", currentScanBlock))
		currentScanBlock++
	}
}