package config

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"os"
	"path/filepath"

	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store/file_storage"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store/inmemory_storage"
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

func GetPlaylistStore(cfg *Config, guildID string, logger logging.Logger, persistent file_storage.StatePersistent) (store.SongStorage, store.StateStorage) {
	switch cfg.Store.Type {
	case "memory":
		return inmemory_storage.NewInmemorySongStorage(logger), inmemory_storage.NewInmemoryStateStorage(logger)
	case "file":
		if err := os.MkdirAll(cfg.Store.File.Dir, 0755); err != nil {
			panic(err)
		}
		path := filepath.Join(cfg.Store.File.Dir, guildID+".json")
		songStore, err := file_storage.NewFileSongStorage(path, logger, persistent)
		if err != nil {
			panic(err)
		}
		stateStore := inmemory_storage.NewInmemoryStateStorage(logger)
		return songStore, stateStore
	default:
		panic("tipo de store invalido")
	}
}
