package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"time"

	//"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/zy99978455-otw/flash-monitor/internal/indexer"

	"github.com/zy99978455-otw/flash-monitor/internal/data"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	rpc struct {
		mainNode string
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found", err)
	}

	var cfg config

	flag.IntVar(&cfg.port, "port", 4010, "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV"), "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("FLASH_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.StringVar(&cfg.rpc.mainNode, "rpc-main", os.Getenv("ETH_RPC_MAIN"), "Ethereum RPC Node URL")

	flag.Parse()

	// 初始化日志
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// 建立数据库连接池
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Printf("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	logger.Printf("using RPC node: %s", app.config.rpc.mainNode)

	// 初始化抓取引擎
	engine, err := indexer.NewEngine(app.config.rpc.mainNode, app.models, app.logger)
	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go engine.Start(ctx)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(), // 挂载路由
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting flash-monitor HTTP server in %s mode on port %d", app.config.env, app.config.port)

	// ListenAndServe 会一直阻塞在这里监听请求，所以不会再报 deadlock 了
	err = srv.ListenAndServe()
	logger.Fatal(err)
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

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	// 创建一个带 5 秒超时的上下文，防止 Ping 一直卡住
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 实际建立连接测试
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
