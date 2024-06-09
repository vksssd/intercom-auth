package handlers

import (
	"net/http"
	"time"

	"github.com/vksssd/intercom-auth/internal/jwt"
	"github.com/vksssd/intercom-auth/internal/session"
	"github.com/vksssd/intercom-auth/internal/utils"
)

// Replace "your-session-key" with your actual session key.


func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	sess, err := session.Get(r,"auth-session")
	if err != nil {
		http.Error(w, "Server Error: unable to get session", http.StatusInternalServerError)
		return
	}
	// log.Println(sess)

	tokenString, ok := sess.Values["refresh_token"].(string)
	if !ok {
		http.Error(w, "Forbidden: no valid refresh token found", http.StatusForbidden)
		return
	}

	claims, err := jwt.Parse(tokenString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	username,usernameok := claims["username"].(string)
	email, emailOk := claims["email"].(string)
	if !usernameok || !emailOk {
		http.Error(w, "Forbidden: Invalid token claims", http.StatusForbidden)
		w.Write([]byte(username+"\n"+email))
		return
	}

	newTokenString, err := jwt.GenerateJWT(username, email)
	if err != nil {
		http.Error(w, "Server Error: unable to generate new token", http.StatusInternalServerError)
		return
	}

	sess.Values["auth_token"] = newTokenString

	// Set cookie function
	utils.SetCookie(w, "auth_token", newTokenString, 10000*time.Second)
	
	if err = session.Save(w,r,sess); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+newTokenString)
	w.WriteHeader(http.StatusOK)
}

