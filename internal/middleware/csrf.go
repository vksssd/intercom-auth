package middleware

import (
	"net/http"
	"github.com/gorilla/csrf"
	"github.com/vksssd/intercom-auth/config"
)

func CSRFMiddleware(cfg *config.CSRFConfig)func(http.Handler) http.Handler {
	return csrf.Protect(
		[]byte(cfg.Secret),
		csrf.Secure(true),
		csrf.HttpOnly(true),
		csrf.Path("/"),
		csrf.MaxAge(cfg.Expire),
		// csrf.SameSite(http.SameSite),

	)
}