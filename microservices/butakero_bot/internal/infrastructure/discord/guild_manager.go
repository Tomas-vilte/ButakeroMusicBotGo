package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/player"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type GuildManager struct {
	players        map[string]ports.GuildPlayer
	discordSession *discordgo.Session
	storageAudio   ports.StorageAudio
	messenger      ports.DiscordMessenger
	logger         logging.Logger
}

func NewGuildManager(
	session *discordgo.Session,
	storageAudio ports.StorageAudio,
	messenger ports.DiscordMessenger,
	logger logging.Logger,
) ports.GuildManager {
	return &GuildManager{
		players:        make(map[string]ports.GuildPlayer),
		discordSession: session,
		storageAudio:   storageAudio,
		messenger:      messenger,
		logger:         logger,
	}
}

func (g *GuildManager) CreateGuildPlayer(guildID string) (ports.GuildPlayer, error) {
	if guildPlayer, exists := g.players[guildID]; exists {
		return guildPlayer, nil
	}

	voiceChat := voice.NewDiscordVoiceSession(g.discordSession, guildID, g.logger)
	songStorage := inmemory.NewInmemorySongStorage(g.logger)
	stateStorage := inmemory.NewInmemoryStateStorage(g.logger)

	guildPlayer := player.NewGuildPlayer(
		voiceChat,
		songStorage,
		stateStorage,
		g.messenger,
		g.storageAudio,
		g.logger,
	)

	g.players[guildID] = guildPlayer
	return guildPlayer, nil
}

func (g *GuildManager) RemoveGuildPlayer(guildID string) error {
	delete(g.players, guildID)
	return nil
}

func (g *GuildManager) GetGuildPlayer(guildID string) (ports.GuildPlayer, error) {
	guildPlayer, exists := g.players[guildID]
	if !exists {
		return nil, fmt.Errorf("player not found for guild %s", guildID)
	}
	return guildPlayer, nil
}
