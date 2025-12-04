package logger

import (
	"os"

	"github.com/zy99978455-otw/flash-monitor/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func InitLogger() {
	// 1. 定义日志切割配置 (Lumberjack)
	fileWriter := &lumberjack.Logger{
		Filename:   "./logs/monitor.log", // 日志文件路径
		MaxSize:    10,                   // 每个文件最大 10MB
		MaxBackups: 5,                    // 保留最近 5 个文件
		MaxAge:     30,                   // 保留最近 30 天
		Compress:   true,                 // 是否压缩 (zip)
	}

	// 2. 设置编码器 (Encoder)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 时间格式：2023-01-01T12:00:00.000Z
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if config.AppConfig.App.Env == "dev" {
		// 开发环境：控制台用彩色，文件用 JSON (或者也用 Console 格式)
		encoder = zapcore.NewConsoleEncoder(encoderConfig) 
	} else {
		// 生产环境：全部用 JSON
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 3. 同时输出到 控制台 和 文件 (MultiWriteSyncer)
	// HighPriority (Error 等) 和 LowPriority (Info 等) 都打印
	core := zapcore.NewCore(
		encoder, //编码器
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)),
		zap.DebugLevel, // 这里可以根据 config 调整 Level
	)

	// 4. 构建 Logger
	Log = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(Log)
}