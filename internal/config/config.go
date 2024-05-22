package config

import (
	"os"
	"path/filepath"

	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store/file_storage"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store/inmemory_storage"
	"go.uber.org/zap"

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

func GetPlaylistStore(cfg *Config, guildID string, logger *zap.Logger, persistent file_storage.StatePersistent) (store.SongStorage, store.StateStorage) {
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
