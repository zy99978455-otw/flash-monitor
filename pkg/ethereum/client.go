package ethereum

import (
	"context"
	"fmt"
	"log"
	"time"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/core/types"
)

// Client 是对 ethclient.Client 的简单封装
// 将来可以在这里加重试逻辑或负载均衡
type Client struct {
	EthClient *ethclient.Client
}

// InitClient 初始化连接
func InitClient(rawUrl string) (*Client, error) {
	// 设置连接超时，防止节点挂了导致程序卡死
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 连接 RPC 节点
	client, err := ethclient.DialContext(ctx, rawUrl)
	if err != nil {
		return nil, fmt.Errorf("无法连接到 RPC 节点: %w", err)
	}

	// 验证连接是否有效 (尝试获取 ChainID)
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("连接成功但无法获取 ChainID: %w", err)
	}
	
	log.Printf("🔌 已连接到以太坊节点, ChainID: %s", chainID.String())

	return &Client{EthClient: client}, nil
}

// GetBlockNumber 获取最新区块高度
func (c *Client) GetBlockNumber(ctx context.Context) (uint64, error) {
	return c.EthClient.BlockNumber(ctx)
}

// GetBlockHeader 获取区块头信息 (包含 Hash 和 ParentHash)
func (c *Client) GetBlockHeader(ctx context.Context, number uint64) (*types.Header, error) {
	// nil 表示获取最新块，我们要传具体的 big.Int
	bigNum := new(big.Int).SetUint64(number)
	return c.EthClient.HeaderByNumber(ctx, bigNum)
}