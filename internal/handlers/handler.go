package handlers

import (
	"context"

	"github.com/go-redis/redis/v8"
	csrf "github.com/vksssd/intercom-auth/internal/CSRF"
	"github.com/vksssd/intercom-auth/internal/jwt"
	"github.com/vksssd/intercom-auth/internal/session"
)

type AuthHandler struct {
	JWTService *jwt.JWTService
	CSRFService *csrf.CSRF
	RedisClient *redis.Client
	SessionService *session.SessionService
	Ctx context.Context
}

func NewAuthHandler(jwtService *jwt.JWTService, csrf *csrf.CSRF ,redisClient *redis.Client, sessionService *session.SessionService, ctx context.Context) *AuthHandler {
	return &AuthHandler{
		JWTService: jwtService,
		CSRFService: csrf,
		RedisClient: redisClient,
		SessionService: sessionService,
		Ctx: ctx,
	}
}