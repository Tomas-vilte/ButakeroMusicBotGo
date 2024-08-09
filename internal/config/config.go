package config

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store/inmemory_storage"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
)

type Config struct {
	DiscordToken  string
	GuildID       string
	CommandPrefix string
	YoutubeApiKey string
	Store         StoreConfig
	BucketName    string
	Region        string
	AccessKey     string
	SecretKey     string
}

type StoreConfig struct {
	Type string
	File FileStoreConfig
}

type FileStoreConfig struct {
	Dir string
}

func GetPlaylistStore(cfg *Config, guildID string, logger logging.Logger) (store.SongStorage, store.StateStorage) {
	switch cfg.Store.Type {
	case "memory":
		return inmemory_storage.NewInmemorySongStorage(logger), inmemory_storage.NewInmemoryStateStorage(logger)
	default:
		panic("tipo de store invalido")
	}
}
