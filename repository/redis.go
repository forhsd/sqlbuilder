package repository

import (
	"context"

	"github.com/forhsd/logger"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var (
	RDB *redis.Client
)

func InitRedis() *redis.Client {
	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal("Error reading config file: %s", err)
	}

	addr := viper.GetString("redis.address")
	password := viper.GetString("redis.password")
	dbIndex := viper.GetInt("redis.database")

	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbIndex,
	})

	// 测试 Redis 连接是否正常
	pong, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis: %v", err)
	}
	logger.Trace("Connected to Redis, Ping Response: %s", pong)

	return RDB
}
