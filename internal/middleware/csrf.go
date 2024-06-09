package middleware

import (
	"net/http"
	// "github.com/gorilla/csrf"
	"github.com/vksssd/intercom-auth/config"
	service "github.com/vksssd/intercom-auth/internal/CSRF"
)

func CSRFMiddleware(cfg *config.CSRFConfig, csrfService *service.CSRF)func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {

			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sessionID, err := r.Cookie("session_id")
				if err != nil || sessionID.Value=="" {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}

				//check for csrf token
				csrfToken := r.Header.Get(cfg.Header)
				if csrfToken == "" {
					http.Error(w, "Forbidden: csrf not found", http.StatusForbidden)
					return
				}

				//validate csrf
				valid, err := csrfService.ValidateCSRF(sessionID.Value, csrfToken)
				if !valid || err != nil {
					http.Error(w, "forbidden: not valid csrf ", http.StatusForbidden)
					return
				}

				//generate a new csrf token for next request
				newToken, err := csrfService.GenerateCSRF(sessionID.Value)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}

				w.Header().Set(cfg.Header, newToken)
				next.ServeHTTP(w,r)

			})
		}
	
	
	
	// return csrf.Protect(
	// 	[]byte(cfg.Secret),
	// 	csrf.Secure(true),
	// 	csrf.HttpOnly(true),
	// 	csrf.Path("/"),
	// 	csrf.MaxAge(cfg.Expire),
	// 	// csrf.SameSite(http.SameSite),

	// )
}