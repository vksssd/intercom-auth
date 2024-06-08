package main

import (
	"github.com/vksssd/intercom-auth/config"
	"fmt"
)


func main() {

	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)

	fmt.Printf("Hello, World!")
}