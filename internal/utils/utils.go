package utils

import (
	"encoding/base64"
)

func DecodeToken(token []byte) (string,bool) {
	// log.Println(token)
	if token == nil {
		return "", false
	}
	return base64.StdEncoding.EncodeToString(token),true
}