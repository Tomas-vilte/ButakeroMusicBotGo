package command

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/interfaces"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	ErrorMessageNotInVoiceChannel = "❌ Debes estar en un canal de voz para usar este comando"
	ErrorMessageFailedToAddSong   = "❌ No se pudo agregar la canción. Por favor, inténtalo de nuevo"
	ErrorMessageServerNotFound    = "❌ No se pudo encontrar el servidor. Intenta de nuevo más tarde"
	ErrorMessageSongRemovalFailed = "❌ No se pudo eliminar la canción. Verifica la posición"
	ErrorMessageNoCurrentSong     = "🔇 No se está reproduciendo ninguna canción actualmente"
)

type CommandHandler struct {
	storage      ports.InteractionStorage
	logger       logging.Logger
	songService  ports.SongService
	messenger    interfaces.DiscordMessenger
	guildManager ports.GuildManager
}

func NewCommandHandler(
	storage ports.InteractionStorage,
	logger logging.Logger,
	songService ports.SongService,
	guildManager ports.GuildManager,
	messenger interfaces.DiscordMessenger,
) *CommandHandler {
	return &CommandHandler{
		storage:      storage,
		logger:       logger,
		songService:  songService,
		messenger:    messenger,
		guildManager: guildManager,
	}
}

func (h *CommandHandler) PlaySong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := trace.WithTraceID(context.Background())

	h.logger.Info("SESOOOOOOOOOOOOOOOOOOOOOO", zap.String("GUILD_ID SESO", ic.GuildID))

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "PlaySong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", "play"),
	)

	userID := ic.Member.User.ID

	vs, ok := h.isUserInVoiceChannel(s, ic)
	if !ok {
		logger.Warn("Usuario no está en canal de voz")
		return
	}

	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔍 Buscando tu canción... Esto puede tomar unos momentos.",
		},
	}); err != nil {
		logger.Error("Error al enviar la respuesta inicial", zap.Error(err))
		return
	}

	originalMsgID, err := h.messenger.GetOriginalResponseID(ic.Interaction)
	if err != nil {
		logger.Warn("No se pudo obtener el ID del mensaje original", zap.Error(err))
	}

	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}
	h.logger.Info("SEXO ANALLLL", zap.String("GUILD_ID SEXO ANAL", g.ID))

	go func(ctx context.Context) {
		input := opt.Options[0].StringValue()
		song, err := h.songService.GetOrDownloadSong(ctx, userID, input, "youtube")
		if err != nil {
			logger.Error("Error al obtener canción", zap.Error(err))

			if originalMsgID != "" {
				if editErr := h.messenger.EditMessageByID(ic.ChannelID, originalMsgID,
					"❌ No se pudo encontrar o descargar la canción. Verifica el enlace o inténtalo de nuevo"); editErr != nil {
					logger.Error("Error al editar el mensaje original", zap.Error(editErr))
				}
			}
			return
		}

		logger.Info("Canción obtenida o descargada", zap.String("título", song.TitleTrack))
		h.storage.SaveSongList(ic.ChannelID, []*entity.DiscordEntity{song})

		guildPlayer, err := h.guildManager.GetGuildPlayer(g.ID)
		if err != nil {
			logger.Error("Error al obtener GuildPlayer", zap.String("guildID", g.ID), zap.Error(err))
			return
		}
		playedSong := &entity.PlayedSong{
			DiscordSong:     song,
			RequestedByName: ic.Member.User.Username,
			RequestedByID:   ic.Member.User.ID,
		}

		if err := guildPlayer.AddSong(ctx, &ic.ChannelID, &vs.ChannelID, playedSong); err != nil {
			logger.Error("Error al agregar la canción", zap.String("voice_channel_id", vs.ChannelID), zap.Error(err))

			if originalMsgID != "" {
				if editErr := h.messenger.EditMessageByID(ic.ChannelID, originalMsgID, ErrorMessageFailedToAddSong); editErr != nil {
					logger.Error("Error al editar el mensaje original", zap.Error(editErr))
				}
			}
			return
		}

		logger.Info("Canción agregada a la cola",
			zap.String("título", song.TitleTrack),
			zap.String("voice_channel_id", vs.ChannelID),
		)

		successMessage := "✅ Canción agregada a la cola: " + song.TitleTrack
		if originalMsgID != "" {
			if editErr := h.messenger.EditMessageByID(ic.ChannelID, originalMsgID, successMessage); editErr != nil {
				logger.Error("Error al editar el mensaje original", zap.Error(editErr))
			}
		}
	}(ctx)
}

func (h *CommandHandler) StopPlaying(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := trace.WithTraceID(context.Background())

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "StopPlaying"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", "stop"),
	)
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, "Ocurrió un error al obtener la información del servidor")
		return
	}

	guildPlayer, err := h.guildManager.GetGuildPlayer(g.ID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer",
			zap.Error(err),
			zap.String("guild_id", g.ID),
		)
		return
	}

	if err := guildPlayer.Stop(ctx); err != nil {
		logger.Error("Error al detener la reproducción", zap.Error(err))
		h.respondWithError(ic, "Ocurrió un error al detener la reproducción")
		return
	}

	logger.Debug("Reproducción detenida exitosamente")

	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹️ Reproducción detenida",
		},
	}); err != nil {
		logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}

func (h *CommandHandler) isUserInVoiceChannel(s *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.VoiceState, bool) {
	ctx := trace.WithTraceID(context.Background())

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "isUserInVoiceChannel"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("method", "isUserInVoiceChannel"),
	)

	if ic.Member == nil {
		if err := h.messenger.RespondWithMessage(ic.Interaction, ErrorMessageNotInVoiceChannel); err != nil {
			logger.Error("Error al enviar mensaje de error de canal de voz", zap.Error(err))
		}
		return nil, false
	}

	vs, err := s.State.VoiceState(ic.GuildID, ic.Member.User.ID)
	if err != nil {
		logger.Warn("Error al obtener estado de voz del usuario", zap.Error(err))
		if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: ErrorMessageNotInVoiceChannel,
			}}); err != nil {
			logger.Error("Error al enviar mensaje de error de canal de voz", zap.Error(err))
		}
		return nil, false
	}

	if vs == nil || vs.ChannelID == "" {
		logger.Warn("Usuario no está en canal de voz")
		if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: ErrorMessageNotInVoiceChannel,
			},
		}); err != nil {
			logger.Error("Error al enviar mensaje de error de canal de voz", zap.Error(err))
		}
		return nil, false
	}

	logger.Debug("Usuario encontrado en canal de voz",
		zap.String("voice_channel_id", vs.ChannelID),
	)
	return vs, true
}

func (h *CommandHandler) SkipSong(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := trace.WithTraceID(context.Background())

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "SkipSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", "skip"),
	)
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Info("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, "Ocurrió un error al obtener la información del servidor")
		return
	}

	guildPlayer, err := h.guildManager.GetGuildPlayer(g.ID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer",
			zap.Error(err),
			zap.String("guild_id", g.ID),
		)
		return
	}
	guildPlayer.SkipSong(ctx)
	logger.Debug("Canción omitida exitosamente")
	h.respondWithError(ic, "⏭️ Canción omitida")
}

func (h *CommandHandler) ListPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := trace.WithTraceID(context.Background())

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "ListPlaylist"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", "list"),
	)
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, "Ocurrió un error al obtener la información del servidor")
		return
	}

	guildPlayer, err := h.guildManager.GetGuildPlayer(g.ID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer",
			zap.Error(err),
			zap.String("guild_id", g.ID),
		)
		return
	}
	songs, err := guildPlayer.GetPlaylist(ctx)
	if err != nil {
		logger.Error("Error al obtener la lista de reproducción", zap.Error(err))
		h.respondWithError(ic, "Error al obtener la lista de reproducción")
		return
	}

	if len(songs) == 0 {
		logger.Debug("Lista de reproducción vacía")
		h.respondWithError(ic, "📭 La lista de reproducción está vacía")
		return
	}

	message := "🎵 Lista de reproducción:\n"
	for i, song := range songs {
		message += fmt.Sprintf("%d. %s\n", i+1, song)
	}

	logger.Debug("Mostrando lista de reproducción", zap.Int("total_canciones", len(songs)))

	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{Title: "Lista de reproducción:", Description: message},
			},
		},
	}); err != nil {
		logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

func (h *CommandHandler) RemoveSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := trace.WithTraceID(context.Background())

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "RemoveSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", "remove"),
	)
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}

	guildPlayer, err := h.guildManager.GetGuildPlayer(g.ID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer",
			zap.Error(err),
			zap.String("guild_id", g.ID),
		)
		return
	}
	position := opt.Options[0].IntValue()

	song, err := guildPlayer.RemoveSong(ctx, int(position))
	if err != nil {
		logger.Error("Error al eliminar la canción", zap.Error(err))
		h.respondWithError(ic, ErrorMessageSongRemovalFailed)
		return
	}

	logger.Debug("Canción eliminada exitosamente",
		zap.String("song_title", song.TitleTrack),
	)

	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🗑️ Canción **%s** eliminada de la lista", song.TitleTrack),
		},
	}); err != nil {
		logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}

func (h *CommandHandler) GetPlayingSong(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := trace.WithTraceID(context.Background())

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "GetPlayingSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", "nowplaying"),
	)
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}

	guildPlayer, err := h.guildManager.GetGuildPlayer(g.ID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer",
			zap.Error(err),
			zap.String("guild_id", g.ID),
		)
		return
	}
	song, err := guildPlayer.GetPlayedSong(ctx)
	if err != nil {
		logger.Error("Error al obtener la canción actual", zap.Error(err))
		h.respondWithError(ic, "Error al obtener la información de la canción")
		return
	}

	if song == nil {
		logger.Debug("No hay canción reproduciéndose actualmente")
		h.respondWithError(ic, ErrorMessageNoCurrentSong)
		return
	}
	logger.Debug("Mostrando canción actual",
		zap.String("song_title", song.DiscordSong.TitleTrack),
	)

	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🎵 Reproduciendo: %s", song.DiscordSong.TitleTrack),
		},
	}); err != nil {
		logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

func (h *CommandHandler) respondWithError(ic *discordgo.InteractionCreate, message string) {
	if err := h.messenger.RespondWithMessage(ic.Interaction, message); err != nil {
		h.logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

func (h *CommandHandler) PauseSong(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	ctx := trace.WithTraceID(context.Background())

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "PauseSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", "pause"),
	)
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}

	guildPlayer, err := h.guildManager.GetGuildPlayer(g.ID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer",
			zap.Error(err),
			zap.String("guild_id", g.ID),
		)
		return
	}

	if err := guildPlayer.Pause(ctx); err != nil {
		logger.Error("Error al pausar la reproducción", zap.Error(err))
		h.respondWithError(ic, "❌ Ocurrió un error al pausar la reproducción")
		return
	}

	logger.Debug("Reproducción pausada exitosamente")

	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏸️ Reproducción pausada",
		},
	}); err != nil {
		logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}

func (h *CommandHandler) ResumeSong(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	ctx := trace.WithTraceID(context.Background())

	logger := h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", "ResumeSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", "resume"),
	)
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}

	guildPlayer, err := h.guildManager.GetGuildPlayer(g.ID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer",
			zap.Error(err),
			zap.String("guild_id", g.ID),
		)
		return
	}

	if err := guildPlayer.Resume(ctx); err != nil {
		logger.Error("Error al reanudar la reproducción", zap.Error(err))
		h.respondWithError(ic, "❌ Ocurrió un error al reanudar la reproducción")
		return
	}

	logger.Debug("Reproducción reanudada exitosamente")

	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "▶️ Reproducción reanudada",
		},
	}); err != nil {
		logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}
