package utils

import (
	"errors"
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
func SetCookiee(w http.ResponseWriter, name, value string){
	// expirationTime := time.Now().Add(expiration)
	cookie := http.Cookie{
		Name: name, 
		Value: value,
		Path: "/",
		// Expires: expirationTime,
		HttpOnly: true,  //inaccessbile to JS
		Secure: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}


func GetCookie(r *http.Request, name string)(string, error){
	cookie, err := r.Cookie(name)
	if err!= nil {
		 if errors.Is(err, http.ErrNoCookie){
			return "", errors.New("Cookie not found")		
		 }
		 return "" , err
	}

	return cookie.Value, nil
}