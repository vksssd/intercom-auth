package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/vksssd/intercom-auth/internal/jwt"
	"github.com/vksssd/intercom-auth/internal/models"
	"github.com/vksssd/intercom-auth/internal/utils"
	"github.com/vksssd/intercom-auth/pkg/redis"
)


func RegisterHandler(w http.ResponseWriter, r *http.Request){
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	hashedPassword, err := utils.Hash(user.Password)
	if err != nil {
		http.Error(w, "server error ", http.StatusInternalServerError)
		return
	}

	err = redis.RedisClient.Set(ctx, user.Username, hashedPassword, 0).Err()
	if err != nil {
		log.Println(err)
		http.Error(w, "Redis Server error", http.StatusInternalServerError)
		return
	}

	result,err := redis.RedisClient.Get(ctx, user.Username).Result()
	
	w.Write([]byte(result))
	w.WriteHeader(http.StatusCreated)
}


func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	storedHash, err := redis.RedisClient.Get(ctx, user.Username).Result()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unauthorized to login", http.StatusUnauthorized)
		return
	}

	// Compare the provided password with the stored hash
	if !utils.CompareHash(user.Password, storedHash) {
		http.Error(w, "Unauthorized to login", http.StatusUnauthorized)
		return
	}

	token, err := jwt.GenerateJWT(user.Username, user.Email)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.Write([]byte(token))
}



func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

