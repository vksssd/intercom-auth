package jwt

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/vksssd/intercom-auth/config"
)

type JWTService struct {
	Config *config.JWTConfig
}

func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{Config: cfg}
}


func (j *JWTService) GenerateJWT(username, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"email":email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(j.Config.Secret))
}

func (j *JWTService) ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.Config.Secret), nil
	})
}

func (j *JWTService)Parse(tokenString string)(jwt.MapClaims, error) {
	// log.Println(tokenString)
	token, err := j.ValidateJWT(tokenString)
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