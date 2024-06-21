package api

import (
	"context"
	"log"

	"github.com/fasthttp/router"
	// "github.com/valyala/fasthttp"
	"github.com/vksssd/intercom-auth/config"
	// csrf "github.com/vksssd/intercom-auth/internal/CSRF"
	"github.com/vksssd/intercom-auth/internal/handlers"
	// "github.com/vksssd/intercom-auth/internal/jwt"
	// "github.com/vksssd/intercom-auth/internal/session"
	"github.com/vksssd/intercom-auth/pkg/redis"
)




func SetupRoutes() *router.Router {

	cfg, err := config.ConfigInit()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	//pingin redis
	redis.Init(&cfg.Redis)
	pong, err := redis.RedisClient.Ping(context.TODO()).Result()
	if err != nil {
		log.Printf(err.Error())
	}	
	log.Println(pong)

	// session, err := session.NewSessionService(*redis.RedisClient, cfg.Session)
	// if err != nil {
	// 	log.Println("session not generated")
	// }
	// jwt := jwt.NewJWTService(&cfg.JWT)
	// csrf := csrf.NewCSRF(redis.RedisClient,context.TODO())
	// session.Configure("auth-session", 30*60, true, []byte("your-32-byte-secret-key"))
	// handler:= handlers.NewAuthHandler(jwt,csrf,redis.RedisClient,session,context.TODO())


	r:= router.New()
	
	r.POST("/ping", handlers.FastHandler)

	return r

}