package config

import (
	"log"

	"github.com/spf13/viper"
)

// 全局配置变量
var AppConfig *Config

// Config 映射结构体
type Config struct {
	App      AppSettings     `mapstructure:"app"`
	Chain    ChainConfig    `mapstructure:"chain"`
	Database DatabaseConfig `mapstructure:"database"`
}

// AppSettings 应用配置
type AppSettings struct {
	Env  string `mapstructure:"env"`
	Name string `mapstructure:"name"`
}

// ChainSettings 链配置
type ChainConfig struct {
	RpcUrl          string `mapstructure:"rpc_url"`
	ContractAddress string `mapstructure:"contract_address"`
}

// DatabaseSettings 数据库配置
type DatabaseConfig struct {
	Dsn string `mapstructure:"dsn"`
}

// InitConfig 初始化配置
func InitConfig() {
	viper.SetConfigName("config")   // 配置文件名 (不带后缀)
	viper.SetConfigType("yaml")     // 文件类型
	viper.AddConfigPath("./configs") // 查找路径 (相对于项目根目录)

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("❌ 读取配置文件失败: %v", err)
	}

	// 解析到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("❌ 解析配置文件失败: %v", err)
	}
}