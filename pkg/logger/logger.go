package logger

import (
	"os"

	"github.com/zy99978455-otw/flash-monitor/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局 Logger 对象
var Log *zap.Logger

func InitLogger() {
	var core zapcore.Core

	// 根据配置决定日志格式
	// dev 环境：彩色日志，方便看
	// prod 环境：JSON 日志，方便机器采集
	if config.AppConfig.App.Env == "dev" {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 彩色 Level
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		core = zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.DebugLevel)
	} else {
		jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		core = zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel)
	}

	// AddCaller: 显示日志是哪行代码打的
	Log = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(Log)
}