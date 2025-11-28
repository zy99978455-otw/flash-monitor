package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/zy99978455-otw/flash-monitor/internal/config"
	"github.com/zy99978455-otw/flash-monitor/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库对象
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	// 1. 读取配置
	dsn := config.AppConfig.Database.Dsn
	
	// 2. 设置日志模式
	var gormLogger logger.Interface
	if config.AppConfig.App.Env == "dev" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// 3. 连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}

	// 4. 设置连接池
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 5. 自动建表 (关键步骤)
	err = DB.AutoMigrate(
		&model.BlockTrace{},
		&model.TransferEvent{},
	)
	if err != nil {
		log.Fatalf("❌ 数据库自动迁移失败: %v", err)
	}

	fmt.Println("✅ 数据库表结构初始化成功")
}