// internal/config/redis.go
package config

import (
    "os"

    "github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_ADDR"),     // localhost:6379
        Password: os.Getenv("REDIS_PASSWORD"), // пусто если нет
        DB:       0,                           // номер БД (0-15)
    })
} 