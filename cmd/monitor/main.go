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
	// 1. 初始化
	config.InitConfig()
	logger.InitLogger()
	repository.InitDB()

	// 2. 连接节点
	client, err := ethereum.InitClient(config.AppConfig.Chain.RpcUrl)
	if err != nil {
		logger.Log.Fatal("节点连接失败", zap.Error(err))
	}
	scanner := core.NewScanner(client)

	// 3. 决定从哪里开始扫 (StartBlock)
	lastBlock, err := repository.GetLastScannedBlock()
	if err != nil {
		logger.Log.Fatal("查询数据库失败", zap.Error(err))
	}

	var startBlock uint64
	if lastBlock == 0 {
		// 场景 A: 第一次运行，从链上最新高度往前推 50 个块开始
		current, _ := client.GetBlockNumber(context.Background())
		startBlock = current - 50
		logger.Log.Info("✨ 首次运行，从最新高度回溯启动", zap.Uint64("start", startBlock))
	} else {
		// 场景 B: 断点续传，从上次结束的下一个块开始
		startBlock = lastBlock + 1
		logger.Log.Info("🔄 发现历史记录，继续扫描", zap.Uint64("start", startBlock))
	}

	// 4. 开启无限循环扫描 (Loop)
	for {
		// 获取链上最新高度
		currentBlock, err := client.GetBlockNumber(context.Background())
		if err != nil {
			logger.Log.Error("获取最新高度失败，重试中...", zap.Error(err))
			time.Sleep(3 * time.Second)
			continue
		}

		// 如果追上最新高度了，就休息一会儿
		if startBlock > currentBlock {
			logger.Log.Debug("已追上最新高度，等待出块...", zap.Uint64("latest", currentBlock))
			time.Sleep(3 * time.Second) // 以太坊每 12 秒一个块，休息 3 秒比较合适
			continue
		}

		// 每次最多扫 10 个块 (防止一次查太多超时)
		endBlock := startBlock + 10
		if endBlock > currentBlock {
			endBlock = currentBlock
		}

		// 执行扫描
		logger.Log.Info("开始扫描区间", zap.Uint64("from", startBlock), zap.Uint64("to", endBlock))
		scanner.Scan(context.Background(), startBlock, endBlock)

		// 5. 记录状态 (Checkpoint)
		// 注意：这里为了简单，我们直接记录 endBlock。
		// 在 Phase 4 做防回滚时，这里需要存 BlockHash。
		err = repository.SaveBlockTrace(&model.BlockTrace{
			BlockNumber: endBlock,
			BlockHash:   "pending_hash_phase4", // 暂时占位，Phase 4 完善
			ParentHash:  "pending_parent_phase4",
		})
		
		if err != nil {
			logger.Log.Error("保存扫描进度失败", zap.Error(err))
			// 如果保存进度失败，不更新 startBlock，下次重试
			time.Sleep(1 * time.Second)
			continue
		}

		logger.Log.Info("💾 进度已保存", zap.Uint64("当前高度", endBlock))
		// 更新下一次的起点
		startBlock = endBlock + 1
	}
}