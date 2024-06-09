package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
// var jwtSecret = []byte("secret")

func GenerateJWT(username, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"email":email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}

func Parse(tokenString string)(jwt.MapClaims, error) {
	// log.Println(tokenString)
	token, err := ValidateJWT(tokenString)
	if err != nil || !token.Valid {
		// http.Error(w, "Forbidden")

		return nil, fmt.Errorf("forbidden: token not valid : %v\n%v", err,token)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("forbidden: claim not valid")
	}
	return claims, nil
}