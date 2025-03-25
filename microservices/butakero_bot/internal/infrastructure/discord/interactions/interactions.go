package interactions

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/player"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	ErrorMessageNotInVoiceChannel = "❌ Debes estar en un canal de voz para usar este comando"
	ErrorMessageFailedToAddSong   = "❌ No se pudo agregar la canción. Por favor, inténtalo de nuevo"
	ErrorMessageServerNotFound    = "❌ No se pudo encontrar el servidor. Intenta de nuevo más tarde"
	ErrorMessageNoSongSelected    = "❌ No se seleccionó ninguna canción"
	ErrorMessageNoSongsAvailable  = "📭 No hay canciones disponibles para agregar"
	ErrorMessageSongRemovalFailed = "❌ No se pudo eliminar la canción. Verifica la posición"
	ErrorMessageNoCurrentSong     = "🔇 No se está reproduciendo ninguna canción actualmente"
)

// GuildID representa el ID de un servidor de Discord.
type GuildID string

// InteractionHandler maneja las interacciones de Discord.
type InteractionHandler struct {
	guildsPlayers    map[GuildID]*player.GuildPlayer
	storage          ports.InteractionStorage
	cfg              *config.Config
	logger           logging.Logger
	discordMessenger ports.DiscordMessenger
	storageAudio     ports.StorageAudio
	songService      ports.SongService
}

// NewInteractionHandler crea una nueva instancia de InteractionHandler.
func NewInteractionHandler(
	storage ports.InteractionStorage,
	cfg *config.Config,
	logger logging.Logger,
	discordMessenger ports.DiscordMessenger,
	storageAudio ports.StorageAudio,
	songService ports.SongService,
) *InteractionHandler {
	return &InteractionHandler{
		guildsPlayers:    make(map[GuildID]*player.GuildPlayer),
		storage:          storage,
		cfg:              cfg,
		logger:           logger,
		discordMessenger: discordMessenger,
		storageAudio:     storageAudio,
		songService:      songService,
	}
}

// Ready se llama cuando el bot está listo para recibir interacciones.
func (handler *InteractionHandler) Ready(s *discordgo.Session, _ *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, fmt.Sprintf("con tu vieja /%s", handler.cfg.CommandPrefix)); err != nil {
		handler.logger.Error("Error al actualizar el estado del juego", zap.Error(err))
	}
}

// GuildCreate se llama cuando el bot se une a un nuevo servidor.
func (handler *InteractionHandler) GuildCreate(ctx context.Context, s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	guildPlayer := handler.setupGuildPlayer(GuildID(event.Guild.ID), s)
	handler.guildsPlayers[GuildID(event.Guild.ID)] = guildPlayer
	handler.logger.Debug("Conectando al servidor", zap.String("guildID", event.Guild.ID))
	go func() {
		if err := guildPlayer.Run(ctx); err != nil {
			handler.logger.Error("Error al ejecutar el reproductor", zap.Error(err))
		}
	}()
}

// GuildDelete se llama cuando el bot es removido de un servidor.
func (handler *InteractionHandler) GuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	guildID := GuildID(event.Guild.ID)

	handler.getGuildPlayer(guildID, s)
	delete(handler.guildsPlayers, guildID)
}

// PlaySong maneja el comando de reproducción de una canción.
func (handler *InteractionHandler) PlaySong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Error("Error al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageServerNotFound); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	handler.getVoiceChannelMembers(s, ic.ChannelID)

	vs := getUsersVoiceState(g, ic.Member.User)
	if vs == nil {
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageNotInVoiceChannel); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	input := opt.Options[0].StringValue()

	if err := handler.discordMessenger.Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔍 Buscando tu canción... Esto puede tomar unos momentos.",
		},
	}); err != nil {
		handler.logger.Error("Error al enviar la respuesta inicial", zap.Error(err))
		return
	}

	go func() {
		song, err := handler.songService.GetOrDownloadSong(context.Background(), input, "youtube")
		if err != nil {
			handler.logger.Error("Error al obtener canción", zap.Error(err))
			if err := handler.discordMessenger.EditOriginalResponse(ic.Interaction, &discordgo.WebhookEdit{
				Content: shared.StringPtr("❌ No se pudo encontrar o descargar la canción. Verifica el enlace o inténtalo de nuevo"),
			}); err != nil {
				handler.logger.Error("Error al actualizar mensaje de error", zap.Error(err))
			}
			return
		}

		handler.logger.Info("Canción obtenida o descargada", zap.String("título", song.TitleTrack))
		handler.storage.SaveSongList(ic.ChannelID, []*entity.DiscordEntity{song})

		guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
		playedSong := &entity.PlayedSong{
			DiscordSong: song,
			RequestedBy: ic.Member.User.Username,
		}

		if err := guildPlayer.AddSong(&ic.ChannelID, &vs.ChannelID, playedSong); err != nil {
			handler.logger.Error("Error al agregar la canción:", zap.Error(err))
			if err := handler.discordMessenger.EditOriginalResponse(ic.Interaction, &discordgo.WebhookEdit{
				Content: shared.StringPtr(ErrorMessageFailedToAddSong),
			}); err != nil {
				handler.logger.Error("Error al actualizar mensaje de error", zap.Error(err))
			}
			return
		}

		handler.logger.Info("Canción agregada a la cola", zap.String("título", song.TitleTrack))

		if err := handler.discordMessenger.EditOriginalResponse(ic.Interaction, &discordgo.WebhookEdit{
			Content: shared.StringPtr("✅ Canción agregada a la cola: " + song.TitleTrack),
		}); err != nil {
			handler.logger.Error("Error al actualizar mensaje de confirmación", zap.Error(err))
		}
	}()
}

// AddSong maneja la adición de una canción.
func (handler *InteractionHandler) AddSong(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("Error al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageServerNotFound); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	vs := getUsersVoiceState(g, ic.Member.User)
	if vs == nil {
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageNotInVoiceChannel); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	values := ic.MessageComponentData().Values
	if len(values) == 0 {
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageNoSongSelected); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	songs := handler.storage.GetSongList(ic.ChannelID)
	if len(songs) == 0 {
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageNoSongsAvailable); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	song := &entity.DiscordEntity{
		URL:        values[0],
		TitleTrack: values[0],
	}

	playedSong := &entity.PlayedSong{
		DiscordSong: song,
		RequestedBy: ic.Member.User.Username,
	}

	if err := guildPlayer.AddSong(&ic.ChannelID, &vs.ChannelID, playedSong); err != nil {
		handler.logger.Error("Error al agregar la canción", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageFailedToAddSong); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "✅ Canción agregada a la cola"); err != nil {
		handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}

	handler.storage.DeleteSongList(ic.ChannelID)
}

// StopPlaying detiene la reproducción de música.
func (handler *InteractionHandler) StopPlaying(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Error("Error al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)

	if err := guildPlayer.Stop(); err != nil {
		handler.logger.Error("Error al detener la reproducción", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "Ocurrió un error al detener la reproducción"); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹️ Reproducción detenida",
		},
	}); err != nil {
		handler.logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}

// SkipSong salta la canción actualmente en reproducción.
func (handler *InteractionHandler) SkipSong(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("Error al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	guildPlayer.SkipSong()
	if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "⏭️ Canción omitida"); err != nil {
		handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

// ListPlaylist lista las canciones en la lista de reproducción actual.
func (handler *InteractionHandler) ListPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Error("Error al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	songs, err := guildPlayer.GetPlaylist()
	if err != nil {
		handler.logger.Error("Error al obtener la lista de reproducción", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "Error al obtener la lista de reproducción"); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if len(songs) == 0 {
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "📭 La lista de reproducción está vacía"); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	message := "🎵 Lista de reproducción:\n"
	for i, song := range songs {
		message += fmt.Sprintf("%d. %s\n", i+1, song)
	}

	if err := handler.discordMessenger.Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{Title: "Lista de reproducción:", Description: message},
			},
		},
	}); err != nil {
		handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

// RemoveSong elimina una canción de la lista de reproducción.
func (handler *InteractionHandler) RemoveSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Error("Error al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageServerNotFound); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)

	position := opt.Options[0].IntValue()

	song, err := guildPlayer.RemoveSong(int(position))
	if err != nil {
		handler.logger.Error("Error al eliminar la canción", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageSongRemovalFailed); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🗑️ Canción **%s** eliminada de la lista", song.TitleTrack),
		},
	}); err != nil {
		handler.logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}

// GetPlayingSong obtiene la canción que se está reproduciendo actualmente.
func (handler *InteractionHandler) GetPlayingSong(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Error("Error al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageServerNotFound); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	song, err := guildPlayer.GetPlayedSong()
	if err != nil {
		handler.logger.Error("Error al obtener la canción actual", zap.Error(err))
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, "Error al obtener la información de la canción"); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if song == nil {
		if err := handler.discordMessenger.RespondWithMessage(ic.Interaction, ErrorMessageNoCurrentSong); err != nil {
			handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🎵 Reproduciendo: %s", song.DiscordSong.TitleTrack),
		},
	}); err != nil {
		handler.logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

// setupGuildPlayer configura un reproductor para un servidor dado.
func (handler *InteractionHandler) setupGuildPlayer(guildID GuildID, dg *discordgo.Session) *player.GuildPlayer {
	voiceChat := voice.NewDiscordVoiceSession(dg, string(guildID), handler.logger)
	songStorage := inmemory.NewInmemorySongStorage(handler.logger)
	stateStorage := inmemory.NewInmemoryStateStorage(handler.logger)
	return player.NewGuildPlayer(
		voiceChat,
		songStorage,
		stateStorage,
		handler.discordMessenger,
		handler.storageAudio,
		handler.logger,
	)
}

// getGuildPlayer obtiene un reproductor para un servidor dado.
func (handler *InteractionHandler) getGuildPlayer(guildID GuildID, dg *discordgo.Session) *player.GuildPlayer {
	guildPlayer, ok := handler.guildsPlayers[guildID]
	if !ok {
		guildPlayer = handler.setupGuildPlayer(guildID, dg)
		handler.guildsPlayers[guildID] = guildPlayer
	}
	return guildPlayer
}

// getUsersVoiceState obtiene el estado de voz de un usuario en un servidor dado.
func getUsersVoiceState(guild *discordgo.Guild, user *discordgo.User) *discordgo.VoiceState {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			return vs
		}
	}

	return nil
}

// getVoiceChannelMembers obtiene los miembros presentes en un canal de voz específico.
func (handler *InteractionHandler) getVoiceChannelMembers(s *discordgo.Session, channelID string) {
	channel, err := s.Channel(channelID)
	if err != nil {
		handler.logger.Error("Error al obtener el canal:", zap.Error(err))
		return
	}
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		handler.logger.Error("Error al obtener el guild:", zap.Error(err))
		return
	}
	handler.logger.Info("Miembros en el canal de voz '" + channel.Name + "':")
	for _, voiceState := range guild.VoiceStates {
		userID := voiceState.UserID
		user, err := s.User(userID)
		if err != nil {
			handler.logger.Error("Error al obtener el usuario:", zap.Error(err))
		} else {
			handler.logger.Info("- " + user.Username)
		}
	}

}

// RegisterEventHandlers registra los manejadores de eventos en la sesión de Discord.
func (handler *InteractionHandler) RegisterEventHandlers(s *discordgo.Session, ctx context.Context) {
	// Registrar el manejador de eventos Ready
	s.AddHandler(handler.Ready)

	// Registrar el manejador de eventos GuildCreate
	s.AddHandler(func(session *discordgo.Session, event *discordgo.GuildCreate) {
		handler.GuildCreate(ctx, session, event)
	})

	// Registrar el manejador de eventos GuildDelete
	s.AddHandler(handler.GuildDelete)
}
