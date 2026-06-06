package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/zy99978455-otw/flash-monitor/internal/data"
	"github.com/zy99978455-otw/flash-monitor/internal/indexer"
	"github.com/zy99978455-otw/flash-monitor/internal/rpc"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	// API 限流器
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	rpc struct {
		urls string //单节点变更为支持逗号分割的多节点配置
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	wg     sync.WaitGroup
	broker *Broker

	// 将 NodeManager 注入到全局 application 结构体中
	nodeManager *rpc.Manager

	// 这是一个用来远程关闭引擎的函数开关
	cancelEngine context.CancelFunc
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found", err)
	}

	var cfg config

	flag.IntVar(&cfg.port, "port", 4010, "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV"), "Environment (development|staging|production)")

	// 数据库配置
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("FLASH_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// 读取 ETH_RPC_URLS
	flag.StringVar(&cfg.rpc.urls, "rpc-urls", os.Getenv("ETH_RPC_URLS"), "Comma-separated Ethereum RPC Node URLs")

	// 限流器配置
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	// 初始化日志
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 建立数据库连接池
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	//// =========================================================================
	// [V2 改造核心] 解析多节点配置并初始化 NodeManager
	// =========================================================================
	rawUrls := strings.Split(cfg.rpc.urls, ",")
	var nodeConfigs []rpc.NodeConfig
	for i, u := range rawUrls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}

		// 动态生成节点配置，按照书写顺序决定优先级
		nodeConfigs = append(nodeConfigs, rpc.NodeConfig{
			Name:     "Node-" + string(rune('A'+i)),
			URL:      u,
			Priority: i + 1,
			Timeout:  10 * time.Second,
		})
	}

	nodeManager, err := rpc.NewManager(nodeConfigs, logger)
	if err != nil {
		logger.Error("failed to initialize rpc node manager", "error", err)
		os.Exit(1)
	}

	// 保证程序退出时一定切断所有RPC心跳与连接
	defer nodeManager.Stop()
	logger.Info("rpc node manager initialized", "node_count", len(nodeConfigs))

	ctx, cancel := context.WithCancel(context.Background())

	// 初始化 SSE Broker
	broker := NewBroker(ctx, logger)
	go broker.Start()

	app := &application{
		config:       cfg,
		logger:       logger,
		models:       data.NewModels(db),
		broker:       broker,
		nodeManager:  nodeManager,
		cancelEngine: cancel,
	}

	// [V2 改造] 初始化抓取引擎。
	engine := indexer.NewEngine(app.nodeManager, app.models, app.logger, broker.Broadcast)
	if err != nil {
		logger.Error("failed to initialize indexer engine", "error", err)
		os.Exit(1)
	}

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		engine.Start(ctx)
	}()

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// openDB 负责初始化数据库连接池并进行 Ping 测试
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// 配置连接池参数
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxLifetime(cfg.db.maxIdleTime)

	// 创建一个带 5 秒超时的上下文，防止 Ping 一直卡住
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 实际建立连接测试
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
