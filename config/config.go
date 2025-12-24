package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
)

func LoadConfig() {
	// 开关: 读取环境变量
	viper.AutomaticEnv()
	// 必须开启开关, viper才能从环境变量读取到该值
	//envConfigPah := viper.Get("VIPER_CONFIG") // 和 os.Getenv("VIPER_CONFIG") 值一致

	configPath := "./config"
	if configEnv := os.Getenv("VIPER_CONFIG"); configEnv != "" {
		fmt.Println("VIPER 环境变量: {}", configEnv)
		configPath = configEnv
	}

	viper.SetConfigType("yaml")     // 配置文件类型
	viper.SetConfigName("config")   // 配置文件名（无需扩展名）
	viper.AddConfigPath(configPath) // 配置文件路径, golang的默认路径是项目根目录而不是当前目录

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config failed: %s \n", err))
	}

	// 设置监听回调
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("检测到配置文件已更改:", e.Name)
	})

	// 开启监听
	viper.WatchConfig()
}
