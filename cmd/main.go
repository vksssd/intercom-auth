package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/gorilla/mux"
	"github.com/vksssd/intercom-auth/config"
	"github.com/vksssd/intercom-auth/internal/handlers"
	"github.com/vksssd/intercom-auth/internal/session"
	"github.com/vksssd/intercom-auth/pkg/redis"
)

func main() {

	cfg, err := config.ConfigInit()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	session.Configure("auth-session", 30*60, true, []byte("your-32-byte-secret-key"))

	//pingin redis
	redis.Init(&cfg.Redis)
	pong, err := redis.RedisClient.Ping(context.TODO()).Result()
	if err != nil {
		log.Printf(err.Error())
	}	
	log.Println(pong)

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

	fmt.Println("Ping server is listening on port 8000...")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}
	
}