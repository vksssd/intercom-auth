package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
var jwtSecret = []byte("secret")

func GenerateJWT(username, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}
