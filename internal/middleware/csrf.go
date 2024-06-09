package middleware

import (
	"net/http"
	"github.com/gorilla/csrf"
	"github.com/vksssd/intercom-auth/config"
)

func CSRFMiddleware(cfg *config.CSRFConfig)func(http.Handler) http.Handler {
	return csrf.Protect(
		[]byte(cfg.CSRF_SECRET),
		csrf.Secure(true),
		csrf.HttpOnly(true),
		csrf.Path("/"),
		csrf.MaxAge(3600),
	)
}