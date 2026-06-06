package rpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

// ErrNoHealthyNodes 统一定义包级错误
var ErrNoHealthyNodes = errors.New("no healthy nodes available")

// NodeConfig 节点配置
type NodeConfig struct {
	Name     string
	URL      string
	Priority int
	Weight   int
	Timeout  time.Duration
}

// NodeStatus 节点状态
type NodeStatus struct {
	IsHealthy     bool
	LastCheckTime time.Time
	LatestBlock   uint64
	ResponseTime  time.Duration
	ErrorCount    int
	SuccessCount  int
	LastError     error
}

// Node 节点实例
type Node struct {
	Config    NodeConfig
	Status    NodeStatus
	Client    *ethclient.Client
	RPCClient *ethrpc.Client
	mu        sync.RWMutex
}

// Manager 节点管理器
type Manager struct {
	nodes               []*Node
	mu                  sync.RWMutex
	healthCheckInterval time.Duration
	maxRetries          int

	//依赖注入
	logger *slog.Logger

	// 后台任务管理
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager 创建节点管理器，强制要求传入 logger
func NewManager(configs []NodeConfig, logger *slog.Logger) (*Manager, error) {
	if len(configs) == 0 {
		return nil, errors.New("at least one node configuration is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		nodes:               make([]*Node, len(configs)),
		healthCheckInterval: 30 * time.Second,
		maxRetries:          3,
		logger:              logger,
		ctx:                 ctx,
		cancel:              cancel,
	}

	for _, config := range configs {
		node, err := m.createNode(config)
		if err != nil {
			m.logger.Warn("failed to initialize node", "name", config.Name, "error", err)
			continue
		}
		m.nodes = append(m.nodes, node)
	}

	if len(m.nodes) == 0 {
		return nil, ErrNoHealthyNodes
	}

	m.background(m.startHealthCheck)

	return m, nil
}

// background 后台协程管理模式
// 捕获panic 防止应用崩溃，并使用WaitGroup 确保优雅退出
func (m *Manager) background(fn func()) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("background task panic recovered", "error", err)
			}
		}()
		fn()
	}()
}

func (m *Manager) createNode(config NodeConfig) (*Node, error) {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	rpcClient, err := ethrpc.DialContext(m.ctx, config.URL)
	if err != nil {
		return nil, fmt.Errorf("dial rpc failed: %w", err)
	}

	client := ethclient.NewClient(rpcClient)

	return &Node{
		Config:    config,
		Client:    client,
		RPCClient: rpcClient,
		Status: NodeStatus{
			IsHealthy:     true,
			LastCheckTime: time.Now(),
		},
	}, nil
}

// GetHealthyNode 获取最优健康节点
func (m *Manager) GetHealthyNode() (*Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var bestNode *Node

	for _, node := range m.nodes {
		node.mu.RLock()
		isHealthy := node.Status.IsHealthy
		priority := node.Config.Priority
		node.mu.RUnlock()

		if isHealthy {
			continue
		}

		if bestNode == nil || priority < bestNode.Config.Priority {
			bestNode = node
		}
	}

	if bestNode == nil {
		return nil, ErrNoHealthyNodes
	}
	return bestNode, nil
}

// ExecuteWithRetry 核心执行器：执行操作并自动重试
func (m *Manager) ExecuteWithRetry(fn func(*ethclient.Client) error) error {
	var lastErr error

	for attempt := 0; attempt < m.maxRetries; attempt++ {
		node, err := m.GetHealthyNode()
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(attempt+1))
			continue
		}

		err = fn(node.Client)
		if err == nil {
			node.mu.Lock()
			node.Status.SuccessCount++
			node.mu.Unlock()
			return nil
		}

		node.mu.Lock()
		node.Status.ErrorCount++
		node.Status.LastError = err

		if node.Status.ErrorCount >= 3 {
			node.Status.IsHealthy = false
			m.logger.Warn("node marked as unhealthy due to repeated failures", "name", node.Config.Name, "error", err)
		}

		node.mu.Unlock()

		lastErr = err
		time.Sleep(time.Second * time.Duration(attempt+1))
	}
	return fmt.Errorf("operation failed after %d retries: %w", m.maxRetries, lastErr)
}

func (m *Manager) startHealthCheck() {
	ticker := time.NewTicker(m.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkAllNodes()
		}
	}
}

func (m *Manager) checkAllNodes() {
	m.mu.RLock()
	nodes := m.nodes
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			m.checkNodeHealth(n)
		}(node)
	}
	wg.Wait()
}

func (m *Manager) checkNodeHealth(node *Node) {
	ctx, cancel := context.WithTimeout(m.ctx, node.Config.Timeout)
	defer cancel()

	startTime := time.Now()
	blockNumber, err := node.Client.BlockNumber(ctx)
	responseTime := time.Since(startTime)

	node.mu.Lock()
	defer node.mu.Unlock()

	node.Status.LastCheckTime = time.Now()
	node.Status.ResponseTime = responseTime

	if err != nil {
		node.Status.ErrorCount++
		node.Status.LastError = err
		if node.Status.ErrorCount >= 3 && node.Status.IsHealthy {
			node.Status.IsHealthy = false
			m.logger.Warn("health check failed, node offline",
				"name", node.Config.Name,
				"response_time", responseTime,
				"error", err)
		}
		return
	}

	node.Status.LatestBlock = blockNumber
	node.Status.LastError = nil

	if !node.Status.IsHealthy {
		m.logger.Info("node recovered and is back online",
			"name", node.Config.Name,
			"block", blockNumber)
	}

	node.Status.IsHealthy = true
	node.Status.ErrorCount = 0
	node.Status.SuccessCount++
}

// Stop 完美契合 Let's Go Further 的优雅停机
func (m *Manager) Stop() {
	m.logger.Info("shutting down rpc node manager...")
	m.cancel()  // 通知所有后台任务停止
	m.wg.Wait() // 阻塞，直到所有后台任务 (如健康检查) 彻底清理完成

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, node := range m.nodes {
		if node.Client != nil {
			node.Client.Close()
		}
	}
	m.logger.Info("rpc node manager stopped gracefully")
}
