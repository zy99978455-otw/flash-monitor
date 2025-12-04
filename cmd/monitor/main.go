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

// PipelinePayload 传递给 Worker 的数据包
type PipelinePayload struct {
	Height     uint64
	BlockHash  string
	ParentHash string
	Events     []*model.TransferEvent
}

func main() {
	config.InitConfig()
	logger.InitLogger()
	repository.InitDB()

	client, err := ethereum.InitClient(config.AppConfig.Chain.RpcUrl)
	if err != nil {
		logger.Log.Fatal("节点连接失败", zap.Error(err))
	}
	scanner := core.NewScanner(client)

	// ===========================
	// 1. 启动消费者 Worker (Saver)
	// ===========================
	// 创建一个带缓冲的通道，允许主线程领先 Worker 10 个区块
	saveChan := make(chan *PipelinePayload, 10)

	go func() {
		for payload := range saveChan {
			// A. 批量入库交易
			if len(payload.Events) > 0 {
				err := repository.SaveTransferEventsBatch(payload.Events)
				if err != nil {
					logger.Log.Error("❌ [Worker] 交易入库失败", zap.Error(err))
					// 生产环境这里应该有重试逻辑或死信队列
					continue
				}
			}

			// B. 存档 (Checkpoint)
			err := repository.SaveBlockTrace(&model.BlockTrace{
				BlockNumber: payload.Height,
				BlockHash:   payload.BlockHash,
				ParentHash:  payload.ParentHash,
			})
			if err != nil {
				logger.Log.Error("❌ [Worker] 进度存档失败", zap.Error(err))
				continue
			}

			logger.Log.Info("💾 [Worker] 进度已保存", 
				zap.Uint64("H", payload.Height), 
				zap.Int("Tx数", len(payload.Events)),
			)
		}
	}()

	// ===========================
	// 2. 初始化启动状态 (主线程)
	// ===========================
	lastBlockDB, err := repository.GetLastScannedBlock()
	var currentScanBlock uint64
	// 内存中的 Hash 缓存，用于快速比对防回滚，不需要每次查库
	var lastBlockHashInMemory string 

	if err != nil || lastBlockDB == 0 {
		onChainCurrent, _ := client.GetBlockNumber(context.Background())
		currentScanBlock = onChainCurrent - 50
		logger.Log.Info("✨ 首次启动", zap.Uint64("start", currentScanBlock))
	} else {
		currentScanBlock = lastBlockDB + 1
		// 查出上一个块的 Hash 初始化到内存里
		trace, _ := repository.GetBlockTraceByNumber(lastBlockDB)
		lastBlockHashInMemory = trace.BlockHash
		logger.Log.Info("🔄 断点续传", zap.Uint64("start", currentScanBlock), zap.String("lastHash", lastBlockHashInMemory))
	}

	// ===========================
	// 3. 生产者循环 (Main Loop)
	// ===========================
	for {
		// A. 频率控制
		latestBlock, err := client.GetBlockNumber(context.Background())
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		if currentScanBlock > latestBlock {
			logger.Log.Debug("💤 等待出块...", zap.Uint64("current", currentScanBlock))
			time.Sleep(3 * time.Second)
			continue
		}

		// B. 获取区块头 (用于回滚检测)
		header, err := client.GetBlockHeader(context.Background(), currentScanBlock)
		if err != nil {
			logger.Log.Error("获取区块头失败", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		// C. 【关键】内存回滚检测
		// 如果我们有上一个块的 Hash，且当前块的 Parent 不等于它 -> 回滚！
		if lastBlockHashInMemory != "" && header.ParentHash.Hex() != lastBlockHashInMemory {
			logger.Log.Warn("🚨 检测到回滚 (Reorg)!", 
				zap.Uint64("Height", currentScanBlock),
				zap.String("Expected Parent", lastBlockHashInMemory),
				zap.String("Actual Parent", header.ParentHash.Hex()),
			)

			// 1. 暂停流水线：不再发送新任务
			// 2. 确保 Worker 把手里的活干完 (在这个简单模型里，我们假设 Worker 很快)
			// 3. 执行回滚操作
			prevBlock := currentScanBlock - 1
			repository.DeleteBlockTrace(prevBlock)
			repository.DeleteTransferEventsByBlock(prevBlock)

			// 4. 指针回退
			currentScanBlock = prevBlock
			
			// 5. 更新内存 Hash 为更前一个块的 Hash (需要查库了)
			prevTrace, _ := repository.GetBlockTraceByNumber(currentScanBlock - 1)
			if prevTrace != nil {
				lastBlockHashInMemory = prevTrace.BlockHash
			} else {
				lastBlockHashInMemory = "" // 回退到了起点
			}
			
			logger.Log.Warn("🔙 已回退，重试...", zap.Uint64("NewHeight", currentScanBlock))
			time.Sleep(1 * time.Second)
			continue
		}

		// D. 扫描数据 (生产)
		logger.Log.Info("🔍 [Main] 扫描中...", zap.Uint64("H", currentScanBlock))
		result, err := scanner.Scan(context.Background(), currentScanBlock, currentScanBlock)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// E. 发送给 Worker (非阻塞，除非 Channel 满了)
		saveChan <- &PipelinePayload{
			Height:     result.BlockNumber,
			BlockHash:  header.Hash().Hex(),
			ParentHash: header.ParentHash.Hex(),
			Events:     result.Events,
		}

		// F. 更新内存状态，继续下一个
		lastBlockHashInMemory = header.Hash().Hex()
		currentScanBlock++
	}
}