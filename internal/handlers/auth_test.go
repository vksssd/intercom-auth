package handlers_test

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/gorilla/mux"
// 	"github.com/vksssd/intercom-auth/config"
// 	"github.com/vksssd/intercom-auth/internal/handlers"
// 	"github.com/vksssd/intercom-auth/internal/models"
// 	"github.com/vksssd/intercom-auth/internal/utils"
// 	"github.com/vksssd/intercom-auth/pkg/redis"
// )

// func init() {
// 	// Initialize the Redis client for testing purposes
// 	cfg,_:= config.ConfigInit()
// 	redis.Init(&cfg.Redis)
// }

// func TestRegisterHandler(t *testing.T) {
	
// 	user := models.User{
// 		Username: "testuser",
// 		Password: "password123",
// 	}

// 	userJSON, _ := json.Marshal(user)

// 	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(userJSON))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(handlers.RegisterHandler)
// 	handler.ServeHTTP(rr, req)

// 	//we have to look into this as the handler is returning http.StatusCreated but why is it taking status ok
// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
// 	}

// 	result := rr.Body.String()
// 	if result == "" {
// 		t.Errorf("handler returned empty body")
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()
// 	redis.RedisClient.Del(ctx, user.Username)
// }

// func TestLoginHandler(t *testing.T) {
// 	// First, register a user
// 	user := models.User{
// 		Username: "testuser",
// 		Password: "password123",
// 	}

// 	hashedPassword, _ := utils.Hash(user.Password)

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()
// 	redis.RedisClient.Set(ctx, user.Username, hashedPassword, 0)

// 	userJSON, _ := json.Marshal(user)

// 	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(userJSON))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(handlers.LoginHandler)
// 	handler.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
// 		return
// 	}

// 	token := rr.Header().Get("Authorization")
// 	if token == "" {
// 		t.Errorf("handler returned no token")
// 	}

// 	redis.RedisClient.Del(ctx, user.Username)
// }

// func TestHelloHandler(t *testing.T) {
// 	req, err := http.NewRequest("GET", "/ping", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(handlers.HelloHandler)
// 	handler.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
// 	}

// 	expected := "pong"
// 	if rr.Body.String() != expected {
// 		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
// 	}
// }

// func TestHandlers(t *testing.T) {
// 	r := mux.NewRouter()
// 	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
// 	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
// 	r.HandleFunc("/ping", handlers.HelloHandler).Methods("GET")

// 	// Add tests here using the mux router
// }
