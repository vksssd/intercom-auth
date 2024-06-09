package redis

import (
	// "os"

	"github.com/go-redis/redis/v8"
	"github.com/vksssd/intercom-auth/config"
)

var RedisClient *redis.Client

func Init(cfg *config.RedisConfig) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: cfg.URL,
		// Password: cfg.Password,
		// Database: os.Getenv("REDIS_DATABASE")
	})
}