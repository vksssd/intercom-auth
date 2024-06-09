package utils

import (
	"net/http"
	"time"
)

func SetCookie(w http.ResponseWriter, name, value string, expiration time.Duration){
	expirationTime := time.Now().Add(expiration)
	cookie := http.Cookie{
		Name: name, 
		Value: value,
		Expires: expirationTime,
		HttpOnly: true,  //inaccessbile to JS
	}
	http.SetCookie(w, &cookie)
}