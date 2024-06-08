package utils

import (
	"testing"

)

func TestHash(t *testing.T){
	password:= "mohan"

	hashed, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash function return an error: %v", err)
	}

	if hashed == "" {
		t.Fatal("Hash fucntion returned an empty string")
	}

	//test comparehsh function
	if !CompareHash(password, hashed){
		t.Fatal("Compare function false for correct password")
	}

	incorrect := "wrong"
	if CompareHash(incorrect,hashed){
		t.Fatal("compare hash function returned true for wrong password")
	}

}