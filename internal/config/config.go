package config

import (
	"os"
	"path/filepath"

	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store"
)

type Config struct {
	DiscordToken  string `required:"true"`
	GuildID       string
	CommandPrefix string `required:"true"`
	Store         StoreConfig
}

type StoreConfig struct {
	Type string `default:"memory"`
	File FileStoreConfig
}

type FileStoreConfig struct {
	Dir string `default:"./playlist"`
}

func GetPlaylistStore(cfg *Config, guildID string) bot.GuildPlayerState {
	switch cfg.Store.Type {
	case "memory":
		return store.NewInmemoryGuildPlayerState()
	case "file":
		if err := os.MkdirAll(cfg.Store.File.Dir, 0755); err != nil {
			panic(err)
		}
		path := filepath.Join(cfg.Store.File.Dir, guildID+".json")
		s, err := store.NewFilePlaylistStorage(path)
		if err != nil {
			panic(err)
		}
		return s
	default:
		panic("tipo de store invalido")
	}

}
