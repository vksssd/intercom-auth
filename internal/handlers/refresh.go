package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/vksssd/intercom-auth/internal/jwt"
	"github.com/vksssd/intercom-auth/internal/utils"
)

// Replace "your-session-key" with your actual session key.
var (
	store = sessions.NewCookieStore([]byte("your-session-key"))
)

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	tokenString, ok := session.Values["refresh_token"].(string)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	claims, err := jwt.Parse(tokenString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	username := claims["username"].(string)
	email := claims["email"].(string)
	newTokenString, err := jwt.GenerateJWT(username, email)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	session.Values["auth_token"] = newTokenString

	// Set cookie function
	utils.SetCookie(w, "auth_token", newTokenString, 10000*time.Second)
	
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+newTokenString)
	w.WriteHeader(http.StatusOK)
}

