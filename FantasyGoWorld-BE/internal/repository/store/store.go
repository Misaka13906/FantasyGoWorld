package store

import (
	"context"
	"fantasy-go-world-be/internal/config"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis(cfg config.Redis) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	Redis = client
	log.Println("Redis connection initialized successfully")
	return client, nil
}
