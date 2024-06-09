package csrf

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)



type CSRF struct {
	RedisClient *redis.Client
	Ctx context.Context
}

func NewCSRF(redisClient *redis.Client, ctx context.Context) *CSRF {
	return &CSRF{
		RedisClient: redisClient,
		Ctx: ctx,
	}
}

func (s *CSRF)GenerateCSRF(sessionID string)(string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	//store the token in redis with an expiration time
	err := s.RedisClient.Set(s.Ctx, "csrf_"+sessionID, token, 15*time.Minute).Err()
	if err != redis.Nil {
		return "", err
	}

	return token, nil
}

func (s *CSRF) ValidateCSRF(sessionID, token string)(bool,error) {
	storedTOken, err := s.RedisClient.Get(s.Ctx, "csrf_"+sessionID).Result()
	if err == redis.Nil{
		return false, errors.New("CSRF not found in redis")
	}else if err != nil {
		return false, err
	}

	if storedTOken != token {
		return false, errors.New("invalid CSRF token ")
	}

	err = s.RedisClient.Del(s.Ctx, "csrf_"+sessionID).Err()
	if err != nil {
		return false, err
	}

	return true, nil

}