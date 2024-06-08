package main

import (
	"fmt"
	"github.com/vksssd/intercom-auth/config"

)

func main() {
	num:=config.Hello()
	fmt.Println(num)


	// Call ConfigInit
	// cfg, err := config.Init()
	// if err != nil {
	// 	fmt.Println("Error initializing config:", err)
	// 	return
	// }

	// fmt.Println(cfg)

	cfg, err := config.ConfigInit()
	fmt.Println(cfg,err)

	fmt.Printf("Hello, World!")
}