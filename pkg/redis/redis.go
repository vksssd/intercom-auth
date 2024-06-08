package redis

import (
	// "os"


	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func Init() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:50713",
		// Database: os.Getenv("REDIS_DATABASE")
	})
}