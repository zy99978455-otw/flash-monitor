package main

import (
	"context"

	"github.com/zy99978455-otw/flash-monitor/internal/config"
	"github.com/zy99978455-otw/flash-monitor/internal/core"
	"github.com/zy99978455-otw/flash-monitor/internal/repository"
	"github.com/zy99978455-otw/flash-monitor/pkg/ethereum"
	"github.com/zy99978455-otw/flash-monitor/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. 初始化基础设施
	config.InitConfig()
	logger.InitLogger()
	repository.InitDB()

	// 2. 连接以太坊节点
	logger.Log.Info("正在连接以太坊 RPC 节点...", zap.String("url", config.AppConfig.Chain.RpcUrl))
	client, err := ethereum.InitClient(config.AppConfig.Chain.RpcUrl)
	if err != nil {
		logger.Log.Fatal("以太坊客户端初始化失败", zap.Error(err))
	}

	// 3. 获取当前最新高度
	currentBlock, err := client.GetBlockNumber(context.Background())
	if err != nil {
		logger.Log.Fatal("获取高度失败", zap.Error(err))
	}
	logger.Log.Info("当前链上高度", zap.Uint64("height", currentBlock))

	// 4. 创建扫描器并试运行 (扫描最近 10 个块)
	scanner := core.NewScanner(client)
	
	start := currentBlock - 10
	end := currentBlock

	logger.Log.Info("开始测试扫描任务...", zap.Uint64("from", start), zap.Uint64("to", end))
	
	scanner.Scan(context.Background(), start, end)
	// 保持程序运行
	select {}	
		
}
