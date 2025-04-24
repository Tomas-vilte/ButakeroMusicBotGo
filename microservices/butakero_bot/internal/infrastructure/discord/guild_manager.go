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
	logger := f.logger.With(
		zap.String("component", "GuildPlayerFactory"),
		zap.String("method", "CreatePlayer"),
		zap.String("guildID", guildID),
	)

	logger.Debug("Creando nuevo GuildPlayer")

	voiceChat := voice.NewDiscordVoiceSession(f.discordSession, guildID, f.logger)
	songStorage := inmemory.NewInmemorySongStorage(f.logger)
	stateStorage := inmemory.NewInmemoryStateStorage(f.logger)

	guildPlayer := player.NewGuildPlayer(
		player.Config{
			VoiceSession: voiceChat,
			SongStorage:  songStorage,
			StateStorage: stateStorage,
			Messenger:    f.messenger,
			StorageAudio: f.storageAudio,
			Logger:       f.logger,
		},
	)

	logger.Info("GuildPlayer creado exitosamente")
	return guildPlayer, nil
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
	logger := g.logger.With(
		zap.String("component", "GuildManager"),
		zap.String("method", "CreateGuildPlayer"),
		zap.String("guildID", guildID),
	)
	if guildID == "" {
		err := errors_app.NewAppError(
			errors_app.ErrCodeInvalidGuildID,
			"El ID del guild no puede estar vacío",
			nil,
		)
		logger.Error("Error al crear GuildPlayer", zap.Error(err))
		return nil, err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if guildPlayer, exists := g.players[guildID]; exists {
		logger.Warn("Intento de crear GuildPlayer que ya existe",
			zap.Int("total_players", len(g.players)))
		return guildPlayer, errors_app.NewAppError(
			errors_app.ErrCodeGuildPlayerAlreadyExists,
			fmt.Sprintf("El reproductor para el guild %s ya existe", guildID),
			nil,
		)
	}

	logger.Info("Creando nuevo GuildPlayer para guild")

	newGuildPlayer, err := g.playerFactory.CreatePlayer(guildID)
	if err != nil {
		logger.Error("Error al crear GuildPlayer",
			zap.Error(err),
			zap.Int("total_players", len(g.players)))

		return nil, errors_app.NewAppError(
			errors_app.ErrCodeGuildPlayerCreateFailed,
			"Error al crear el reproductor para el guild",
			err,
		)
	}

	g.players[guildID] = newGuildPlayer
	logger.Info("GuildPlayer creado exitosamente",
		zap.Int("total_players", len(g.players)))

	return newGuildPlayer, nil
}

func (g *GuildManager) RemoveGuildPlayer(guildID string) error {
	logger := g.logger.With(
		zap.String("component", "GuildManager"),
		zap.String("method", "RemoveGuildPlayer"),
		zap.String("guildID", guildID),
	)
	if guildID == "" {
		err := errors_app.NewAppError(
			errors_app.ErrCodeInvalidGuildID,
			"El ID del guild no puede estar vacío",
			nil,
		)
		logger.Error("Error al eliminar GuildPlayer", zap.Error(err))
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.players[guildID]; !exists {
		logger.Warn("Intento de eliminar GuildPlayer inexistente",
			zap.Int("total_players", len(g.players)))
		return errors_app.NewAppError(
			errors_app.ErrCodeGuildPlayerNotFound,
			fmt.Sprintf("Reproductor no encontrado para el guild %s", guildID),
			nil,
		)
	}

	delete(g.players, guildID)
	logger.Info("GuildPlayer eliminado exitosamente",
		zap.Int("total_players", len(g.players)))

	return nil
}

func (g *GuildManager) GetGuildPlayer(guildID string) (ports.GuildPlayer, error) {
	logger := g.logger.With(
		zap.String("component", "GuildManager"),
		zap.String("method", "GetGuildPlayer"),
		zap.String("guildID", guildID),
	)

	if guildID == "" {
		err := errors_app.NewAppError(
			errors_app.ErrCodeInvalidGuildID,
			"El ID del guild no puede estar vacío",
			nil,
		)
		logger.Error("Error al obtener GuildPlayer", zap.Error(err))
		return nil, err
	}

	g.mu.RLock()
	guildPlayer, exists := g.players[guildID]
	g.mu.RUnlock()

	if exists {
		logger.Debug("GuildPlayer encontrado en cache")
		return guildPlayer, nil
	}

	logger.Info("GuildPlayer no encontrado, creando nuevo")
	return g.CreateGuildPlayer(guildID)
}
