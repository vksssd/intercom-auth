package main

import (
	// "context"
	// "fmt"
	"log"
	"net/http"
	"os"
	"time"
    "github.com/joho/godotenv"

	"github.com/gorilla/mux"
	"github.com/vksssd/intercom-auth/config"
	"github.com/vksssd/intercom-auth/internal/handlers"
	"github.com/vksssd/intercom-auth/pkg/redis"
)

func main() {

	_, err := config.ConfigInit()
	// fmt.Println(cfg,err)

	redis.Init()
	//pingin redis
	// log.Printf(redis.RedisClient.Ping(context.TODO()).Err().Error())
	
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r :=  mux.NewRouter()
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/ping", handlers.HelloHandler)

	server := &http.Server{
        Handler:      r,
        Addr:         os.Getenv("URL"),
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    log.Printf("Server is listening on %s", server.Addr)

	log.Fatal(server.ListenAndServe())
	
}