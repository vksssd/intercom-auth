package middleware

import (
	"net/http"

	"github.com/vksssd/intercom-auth/internal/jwt"
	"github.com/vksssd/intercom-auth/internal/session"
)



func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		sess, err := session.Get(r, "auth-session")
		if err != nil {
			http.Error(w, "Server error: Unable to get session", http.StatusInternalServerError)
			return 
		}
		tokenString, ok := sess.Values["auth_token"].(string)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		token, err := jwt.Parse(tokenString)

		//improve it by usein !token.valid
		if err != nil || token == nil  {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return 
		}

		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w,r)

	})
}