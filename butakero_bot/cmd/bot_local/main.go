package main

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/cmd/bot_local/bot"
)

func main() {
	if err := bot.StartBot(); err != nil {
		panic(err)
	}
}
