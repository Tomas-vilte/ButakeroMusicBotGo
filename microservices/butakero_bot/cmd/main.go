package main

import "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/cmd/bot_local"

func main() {
	if err := bot_local.StartBot(); err != nil {
		panic(err)
	}
}
