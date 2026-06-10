package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

var (
	client *redis.Client
	once   sync.Once
)

func Init(cfg RedisConfig) {
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		})
		if err := client.Ping(context.Background()).Err(); err != nil {
			panic(fmt.Sprintf("failed to connect redis: %v", err))
		}
	})
}

func GetClient() *redis.Client { return client }
