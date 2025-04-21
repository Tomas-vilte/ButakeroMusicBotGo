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

func NewGuildManager(playerFactory PlayerFactory, logger logging.Logger) ports.GuildManager {
	return &GuildManager{
		playerFactory: playerFactory,
		players:       make(map[string]ports.GuildPlayer),
		logger:        logger,
	}
}

func (g *GuildManager) CreateGuildPlayer(guildID string) (ports.GuildPlayer, error) {
	if guildID == "" {
		return nil, errors_app.NewAppError(
			errors_app.ErrCodeInvalidGuildID,
			"El ID del guild no puede estar vacío",
			nil,
		)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if guildPlayer, exists := g.players[guildID]; exists {
		return guildPlayer, errors_app.NewAppError(
			errors_app.ErrCodeGuildPlayerAlreadyExists,
			fmt.Sprintf("El reproductor para el guild %s ya existe", guildID),
			nil,
		)
	}

	newGuildPlayer, err := g.playerFactory.CreatePlayer(guildID)
	if err != nil {
		g.logger.Error("Error creando guild player",
			zap.String("guildID", guildID),
			zap.Error(err))

		return nil, errors_app.NewAppError(
			errors_app.ErrCodeGuildPlayerCreateFailed,
			"Error al crear el reproductor para el guild",
			err,
		)
	}

	g.players[guildID] = newGuildPlayer
	g.logger.Debug("Nuevo GuildPlayer creado",
		zap.String("guildID", guildID),
		zap.Int("total_players", len(g.players)))

	return newGuildPlayer, nil
}

func (g *GuildManager) RemoveGuildPlayer(guildID string) error {
	if guildID == "" {
		return errors_app.NewAppError(
			errors_app.ErrCodeInvalidGuildID,
			"El ID del guild no puede estar vacío",
			nil,
		)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.players[guildID]; !exists {
		g.logger.Debug("Intento de eliminar un guild player inexistente",
			zap.String("guildID", guildID))
		return errors_app.NewAppError(
			errors_app.ErrCodeGuildPlayerNotFound,
			fmt.Sprintf("Reproductor no encontrado para el guild %s", guildID),
			nil,
		)
	}

	delete(g.players, guildID)
	g.logger.Debug("GuildPlayer eliminado",
		zap.String("guildID", guildID),
		zap.Int("total_players", len(g.players)))

	return nil
}

func (g *GuildManager) GetGuildPlayer(guildID string) (ports.GuildPlayer, error) {
	if guildID == "" {
		return nil, errors_app.NewAppError(
			errors_app.ErrCodeInvalidGuildID,
			"El ID del guild no puede estar vacío",
			nil,
		)
	}

	g.mu.RLock()
	guildPlayer, exists := g.players[guildID]
	g.mu.RUnlock()

	if exists {
		return guildPlayer, nil
	}

	return g.CreateGuildPlayer(guildID)
}
