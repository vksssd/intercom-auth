package main

import (
	"./config/config"
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