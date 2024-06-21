package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/gorilla/mux"
	"github.com/valyala/fasthttp"
	"github.com/vksssd/intercom-auth/api"
	"github.com/vksssd/intercom-auth/config"
	csrf "github.com/vksssd/intercom-auth/internal/CSRF"
	"github.com/vksssd/intercom-auth/internal/handlers"
	"github.com/vksssd/intercom-auth/internal/jwt"
	"github.com/vksssd/intercom-auth/internal/session"
	"github.com/vksssd/intercom-auth/pkg/redis"
)

func main() {

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

	session, err := session.NewSessionService(*redis.RedisClient, cfg.Session)
	if err != nil {
		log.Println("session not generated")
	}
	jwt := jwt.NewJWTService(&cfg.JWT)
	csrf := csrf.NewCSRF(redis.RedisClient,context.TODO())
	// session.Configure("auth-session", 30*60, true, []byte("your-32-byte-secret-key"))
	handlers:= handlers.NewAuthHandler(jwt,csrf,redis.RedisClient,session,context.TODO())

	r :=  mux.NewRouter()
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/refresh", handlers.RefreshTokenHandler).Methods("POST")
	r.HandleFunc("/ping", handlers.HelloHandler)
	server := &http.Server{
		Handler:      r,
		Addr:         cfg.Server.URL,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Ping server is listening on port 8080 && 8081...")
	go server.ListenAndServe()
	go fasthttp.ListenAndServe(":8081",api.SetupRoutes().Handler)
	
	// if err := server.ListenAndServe(); err != nil {
		// 	fmt.Println("Server error:", err)
		// }
	// if err := fasthttp.ListenAndServe(":8080",api.SetupRoutes().Handler); err != nil {
	// 	log.Fatalf("Error in fast listing ")
	// }

	//infinite loop to keep server on
	for {}
}