package handlers

import (
	"context"
	"encoding/json"
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
		http.Error(w, "SErver error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}


func LoginHandler(w http.ResponseWriter, r *http.Request){

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	_= json.NewDecoder(r.Body).Decode(&user)
		
	storedHash, err := redis.RedisClient.Get(ctx, user.Username).Result()
		if err ==  nil || !utils.CompareHash(user.Password, storedHash){
			http.Error(w, "Invalid Credintails", http.StatusUnauthorized)
			return
		}
		
	token, err := jwt.GenerateJWT(user.Username, user.Email)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)

}