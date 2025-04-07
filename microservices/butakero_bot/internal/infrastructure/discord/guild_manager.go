package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/player"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"sync"
)

var ErrGuildPlayerNotFound = errors_app.NewAppError("guild_player_not_found", "Player not found for the specified guild", nil)

type PlayerFactory interface {
	CreatePlayer(guildID string) (ports.GuildPlayer, error)
}

type GuildPlayerFactory struct {
	discordSession *discordgo.Session
	storageAudio   ports.StorageAudio
	messenger      ports.DiscordMessenger
	logger         logging.Logger
}

func NewGuildPlayerFactory(session *discordgo.Session, storageAudio ports.StorageAudio,
	messenger ports.DiscordMessenger, logger logging.Logger) PlayerFactory {
	return &GuildPlayerFactory{
		discordSession: session,
		storageAudio:   storageAudio,
		messenger:      messenger,
		logger:         logger,
	}
}

func (f *GuildPlayerFactory) CreatePlayer(guildID string) (ports.GuildPlayer, error) {
	voiceChat := voice.NewDiscordVoiceSession(f.discordSession, guildID, f.logger)
	songStorage := inmemory.NewInmemorySongStorage(f.logger)
	stateStorage := inmemory.NewInmemoryStateStorage(f.logger)

	return player.NewGuildPlayer(
		player.Config{
			VoiceSession: voiceChat,
			SongStorage:  songStorage,
			StateStorage: stateStorage,
			Messenger:    f.messenger,
			StorageAudio: f.storageAudio,
			Logger:       f.logger,
		},
	), nil
}

type GuildManager struct {
	players       map[string]ports.GuildPlayer
	playerFactory PlayerFactory
	mu            sync.RWMutex
	logger        logging.Logger
}

func NewGuildManager(
	playerFactory PlayerFactory,
	logger logging.Logger,
) ports.GuildManager {
	return &GuildManager{
		playerFactory: playerFactory,
		players:       make(map[string]ports.GuildPlayer),
		logger:        logger,
	}
}

func (g *GuildManager) CreateGuildPlayer(guildID string) (ports.GuildPlayer, error) {
	g.mu.Lock()
	guildPlayer, exists := g.players[guildID]
	g.mu.Unlock()

	if exists {
		return guildPlayer, nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if guildPlayer, exists = g.players[guildID]; exists {
		return guildPlayer, nil
	}

	newGuildPlayer, err := g.playerFactory.CreatePlayer(guildID)
	if err != nil {
		g.logger.Error("Error creando guild player", zap.String("guildID", guildID), zap.Error(err))
		return nil, fmt.Errorf("fallo al crear el guild player: %w", err)
	}

	g.players[guildID] = newGuildPlayer
	g.logger.Debug("Creando nuevo guild player", zap.String("guildID", guildID))

	return newGuildPlayer, nil
}

func (g *GuildManager) RemoveGuildPlayer(guildID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.players[guildID]; !exists {
		return ErrGuildPlayerNotFound
	}

	delete(g.players, guildID)
	g.logger.Debug("Eliminando guild player", zap.String("guildID", guildID))

	return nil
}

func (g *GuildManager) GetGuildPlayer(guildID string) (ports.GuildPlayer, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	guildPlayer, exists := g.players[guildID]
	if !exists {
		return nil, ErrGuildPlayerNotFound
	}
	return guildPlayer, nil
}
