package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/vksssd/intercom-auth/internal/models"
	"github.com/vksssd/intercom-auth/internal/utils"
)


func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request){
	
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err!=nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Email == "" || user.Password == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.Hash(user.Password)
	if err != nil {
		http.Error(w, "server error ", http.StatusInternalServerError)
		return
	}

	// userData := map[string]interface{}{
	// 	"username": user.Username,
	// 	"email":user.Email,
	// 	"password":hashedPassword,
	// }

	// Check if the username already exists in Redis
	exists, err := h.RedisClient.Exists(h.Ctx, user.Username).Result()
	if err != nil {
		http.Error(w, "Redis server error", http.StatusInternalServerError)
		return
	}

	if exists > 0 {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}


	err = h.RedisClient.HSet(h.Ctx, user.Username, "username", user.Username, "email", user.Email, "password", hashedPassword).Err()
	if err != nil {
		log.Println(err)
		http.Error(w, "Redis Server error", http.StatusInternalServerError)
		return
	}

	result,err := h.RedisClient.HGetAll(h.Ctx, user.Username).Result()
	
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result["username"]))
}


func (h *AuthHandler)LoginHandler(w http.ResponseWriter, r *http.Request) {
	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "Missing username  or password", http.StatusBadRequest)
		return
	}
	
	storedUserData, err := h.RedisClient.HGetAll(h.Ctx, user.Username).Result()
	if err != nil || len(storedUserData) == 0{
		log.Println(err)
		http.Error(w, "Unauthorized to login", http.StatusUnauthorized)
		return
	}

	storedHash:=storedUserData["password"]

	// Compare the provided password with the stored hash
	if !utils.CompareHash(user.Password, storedHash) {
		http.Error(w, "Unauthorized to login", http.StatusUnauthorized)
		return
	}

	email := storedUserData["email"]
	token, err := h.JWTService.GenerateJWT(user.Username, email)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	refreshtoken, err := h.JWTService.GenerateJWT(user.Username, email)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	sess, err := h.SessionService.Get(r, h.SessionService.SessionConfig.Name)
	if err != nil {
		http.Error(w, "Server error: Unable to get session", http.StatusInternalServerError)
		return
	}
	
    // if sess == nil {
		//     sess, err = session.New(r, "auth-session")
		//     if err != nil {
			//         http.Error(w, "Server error: Unable to create session", http.StatusInternalServerError)
			//         return
			//     }
			// }
			
			// log.Printf("Token type: %T, Value: %v", token, token)
			// log.Printf("RefreshToken type: %T, Value: %v", refreshtoken, refreshtoken)
			
	sess.Values["auth_token"]=token
	sess.Values["refresh_token"]=refreshtoken
	// sess.Values["created_at"] = time.Now()
	
	if err := h.SessionService.Save(w,r,sess); err != nil {
		http.Error(w, "Server error: Unable to save session", http.StatusInternalServerError)
		return
	}

	// csrf := CSRF.NewCSRF(nil, nil) /// update this

	csrfToken, err := h.CSRFService.GenerateCSRF(sess.ID)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	utils.SetCookiee(w,"session_id",sess.ID)

	utils.SetCookie(w, "auth_token", token, 10000*time.Second)

	w.Header().Set("Authorization","Bearer "+token)
	w.Header().Set("auth_token",token)
	w.Header().Set("refresh_token",refreshtoken)

	// csrfToken := csrf.Token(r)
	w.Header().Set("X-CSRF-Token", csrfToken)
	
	w.Write([]byte(token+"\n"+user.Username+"\n"+email))
}


func (h *AuthHandler)HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong\n "))
}


func FastHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.AppendBytes([]byte("Hello world"))
	json.NewEncoder(ctx).Encode("hello fast world")
}