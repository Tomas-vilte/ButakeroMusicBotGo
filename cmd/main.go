package main

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Println(cfg)
		panic(err)
	}
}
