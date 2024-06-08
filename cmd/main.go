package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/vksssd/intercom-auth/config"
	"github.com/vksssd/intercom-auth/internal/handlers"
	"github.com/vksssd/intercom-auth/pkg/redis"
)

func main() {
	num:=config.Hello()
	fmt.Println(num)


	// Call ConfigInit
	// cfg, err := config.Init()
	// if err != nil {
	// 	fmt.Println("Error initializing config:", err)
	// 	return
	// }

	// fmt.Println(cfg)

	cfg, err := config.ConfigInit()
	fmt.Println(cfg,err)

	redis.Init()


	r :=  mux.NewRouter()
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	server := &http.Server{
		Handler: r,
		Addr: "localhost:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15*time.Second,
	}

	log.Fatal(server.ListenAndServe())
	
}