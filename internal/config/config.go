package config

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store"
	"os"
	"path/filepath"
)

type Config struct {
	DiscordToken  string `required:"true"`
	OpenAIToken   string
	GuildID       string
	CommandPrefix string `default:"air"`
	Store         StoreConfig
}

type StoreConfig struct {
	Type string `default:"memory"`
	File FileStoreConfig
}

type FileStoreConfig struct {
	Dir string `default:"./playlist"`
}

func GetPlaylistStore(cfg *Config, guildID string) bot.PlaylistManager {
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
		panic("")
	}

}
