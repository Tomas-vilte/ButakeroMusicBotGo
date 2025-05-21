package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/interfaces"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	ErrorMessageNotInVoiceChannel       = "❌ Tenés que estar en un canal de voz para usar este comando, boludo"
	ErrorMessageSongRemovalFailed       = "❌ No se pudo sacar la canción, fijate bien la posición"
	ErrorMessageNoCurrentSong           = "🔇 No hay ninguna canción sonando ahora, maestro"
	ErrorMessageGuildPlayerNotAccesible = "❌ No se pudo agarrar el reproductor de música del servidor, qué bajón"
	ErrorMessageGenericStop             = "❌ Ocurrió un mambo al cortar la reproducción"
	ErrorMessageGenericPlaylist         = "❌ Se arruinó todo al querer ver la lista"
	ErrorMessageGenericPause            = "❌ Se mandó cualquiera al pausar la música"
	ErrorMessageGenericResume           = "❌ No se pudo seguir con la música, qué garronazo"
	ErrorMessageInvalidRemovePosition   = "❌ Tenés que poner un número de posición válido para sacar la canción, dale"

	InfoMessageSearchingSongFmt        = "🔍 Buscando tu tema, dame un toque..."
	SuccessMessageSongAddedFmt         = "✅ Listo, agregué: **%s**"
	SuccessMessagePlayingStopped       = "⏹️ Corté la música, chau"
	SuccessMessageSongSkipped          = "⏭️ Salté esta, a la próxima"
	InfoMessagePlaylistEmpty           = "📭 No hay nada en la lista, agregá algo che"
	SuccessMessageSongRemovedFmt       = "🗑️ Chau **%s**, la sacamos de la lista"
	InfoMessageNowPlayingFmt           = "🎵 Sonando ahora: **%s**"
	SuccessMessagePaused               = "⏸️ Le metí pausa"
	SuccessMessageResumed              = "▶️ Seguimos con el tema"
	InfoMessageSongSkippedNoNextToPlay = "🤷 No hay más temas en la cola, che. Seguimos con este."
	ErrorMessageNothingToSkip          = "🤔 No hay nada sonando para saltar, maestro"
	ErrorMessageSkipGeneric            = "💥 Se mandó una cagada al intentar saltar el tema"
)

type CommandHandler struct {
	storage      ports.InteractionStorage
	logger       logging.Logger
	messenger    interfaces.DiscordMessenger
	guildManager ports.GuildManager
	queueManager ports.PlayRequestService
}

func NewCommandHandler(
	storage ports.InteractionStorage,
	logger logging.Logger,
	guildManager ports.GuildManager,
	messenger interfaces.DiscordMessenger,
	queueManager ports.PlayRequestService,
) *CommandHandler {
	return &CommandHandler{
		storage:      storage,
		logger:       logger,
		messenger:    messenger,
		guildManager: guildManager,
		queueManager: queueManager,
	}
}

func (h *CommandHandler) baseLogger(ctx context.Context, ic *discordgo.InteractionCreate, methodName string, commandName string) logging.Logger {
	return h.logger.With(
		zap.String("component", "CommandHandler"),
		zap.String("method", methodName),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("channel_id", ic.ChannelID),
		zap.String("user_id", ic.Member.User.ID),
		zap.String("command", commandName),
	)
}

func (h *CommandHandler) sendResponse(interaction *discordgo.Interaction, message string) {
	if err := h.messenger.RespondWithMessage(interaction, message); err != nil {
		h.logger.Error("Error al enviar mensaje de respuesta al usuario",
			zap.String("interactionID", interaction.ID),
			zap.Error(err),
		)
	}
}

func (h *CommandHandler) getGuildPlayerAndLog(_ context.Context, ic *discordgo.InteractionCreate, logger logging.Logger) (ports.GuildPlayer, error) {
	guildPlayer, err := h.guildManager.GetGuildPlayer(ic.GuildID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer", zap.Error(err))
		h.sendResponse(ic.Interaction, ErrorMessageGuildPlayerNotAccesible)
		return nil, err
	}
	return guildPlayer, nil
}

func (h *CommandHandler) PlaySong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := trace.WithTraceID(context.Background())
	logger := h.baseLogger(ctx, ic, "PlaySong", "play")

	vs, ok := h.isUserInVoiceChannel(ctx, s, ic)
	if !ok {
		return
	}

	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: InfoMessageSearchingSongFmt},
	}); err != nil {
		logger.Error("Error al enviar respuesta inicial", zap.Error(err))
		return
	}

	originalMsgID, err := h.messenger.GetOriginalResponseID(ic.Interaction)
	if err != nil {
		logger.Warn("No se pudo obtener el ID del mensaje original, se enviará uno nuevo si es necesario.", zap.Error(err))
		originalMsgID = ""
	}

	songInput := ""
	if len(opt.Options) > 0 && opt.Options[0].Type == discordgo.ApplicationCommandOptionString {
		songInput = opt.Options[0].StringValue()
	} else {
		logger.Error("Opción de canción inválida o faltante")
		h.sendResponse(ic.Interaction, "❌ Debes proporcionar el nombre o URL de una canción.")
		return
	}

	resultChan := h.queueManager.Enqueue(ic.GuildID, model.PlayRequestData{
		Ctx:             ctx,
		GuildID:         ic.GuildID,
		ChannelID:       ic.ChannelID,
		VoiceChannelID:  vs.ChannelID,
		UserID:          ic.Member.User.ID,
		SongInput:       songInput,
		RequestedByName: ic.Member.User.Username,
	})

	go func() {
		result := <-resultChan
		var response string
		if result.Err != nil {
			response = fmt.Sprintf("❌ Error: %v", result.Err)
			logger.Error("Error al procesar la canción en la cola", zap.Error(result.Err), zap.String("songTitle", result.SongTitle))
		} else {
			response = fmt.Sprintf(SuccessMessageSongAddedFmt, result.SongTitle)
			logger.Info("Canción agregada exitosamente a la cola", zap.String("songTitle", result.SongTitle))
		}

		var sendErr error
		if originalMsgID != "" {
			sendErr = h.messenger.EditMessageByID(ic.ChannelID, originalMsgID, response)
			if sendErr != nil {
				logger.Error("Error al editar mensaje original, intentando enviar uno nuevo", zap.Error(sendErr))
				sendErr = h.messenger.RespondWithMessage(ic.Interaction, response)
			}
		} else {
			sendErr = h.messenger.RespondWithMessage(ic.Interaction, response)
		}

		if sendErr != nil {
			logger.Error("Error final al enviar/editar mensaje de respuesta para PlaySong", zap.Error(sendErr))
		}
	}()
}

func (h *CommandHandler) StopPlaying(ic *discordgo.InteractionCreate) {
	ctx := trace.WithTraceID(context.Background())
	logger := h.baseLogger(ctx, ic, "StopPlaying", "stop")

	guildPlayer, err := h.getGuildPlayerAndLog(ctx, ic, logger)
	if err != nil {
		return
	}

	if err := guildPlayer.Stop(ctx); err != nil {
		logger.Error("Error al detener la reproducción", zap.Error(err))
		h.sendResponse(ic.Interaction, ErrorMessageGenericStop)
		return
	}

	logger.Debug("Reproducción detenida exitosamente")
	h.sendResponse(ic.Interaction, SuccessMessagePlayingStopped)
}

func (h *CommandHandler) SkipSong(ic *discordgo.InteractionCreate) {
	ctx := trace.WithTraceID(context.Background())
	logger := h.baseLogger(ctx, ic, "SkipSong", "skip")

	guildPlayer, err := h.getGuildPlayerAndLog(ctx, ic, logger)
	if err != nil {
		return
	}

	skipAppErr := guildPlayer.SkipSong(ctx)
	if skipAppErr != nil {
		var appErr *errors_app.AppError
		if errors.As(skipAppErr, &appErr) {
			if appErr.Code == errors_app.ErrCodePlayerNotPlaying {
				logger.Info("Intento de skip pero no había canción reproduciéndose.", zap.String("error_message", appErr.Message))
				h.sendResponse(ic.Interaction, ErrorMessageNothingToSkip)
				return
			}
			if appErr.Code == errors_app.ErrCodePlayerNoNextToSkip {
				logger.Info("Intento de skip, pero no hay siguiente canción. La actual continúa.", zap.String("error_message", appErr.Message))
				h.sendResponse(ic.Interaction, InfoMessageSongSkippedNoNextToPlay)
				return
			}
			logger.Error("Error de aplicación al procesar skip", zap.Error(appErr), zap.String("error_code", string(appErr.Code)))
			h.sendResponse(ic.Interaction, ErrorMessageSkipGeneric)
			return
		}
	}
	logger.Debug("Solicitud de omisión de canción procesada")
	h.sendResponse(ic.Interaction, SuccessMessageSongSkipped)
}

func (h *CommandHandler) ListPlaylist(ic *discordgo.InteractionCreate) {
	ctx := trace.WithTraceID(context.Background())
	logger := h.baseLogger(ctx, ic, "ListPlaylist", "list")

	guildPlayer, err := h.getGuildPlayerAndLog(ctx, ic, logger)
	if err != nil {
		return
	}

	songs, err := guildPlayer.GetPlaylist(ctx)
	if err != nil {
		logger.Error("Error al obtener la lista de reproducción", zap.Error(err))
		h.sendResponse(ic.Interaction, ErrorMessageGenericPlaylist)
		return
	}

	if len(songs) == 0 {
		logger.Debug("Lista de reproducción vacía")
		h.sendResponse(ic.Interaction, InfoMessagePlaylistEmpty)
		return
	}

	message := ""
	for i, song := range songs {
		message += fmt.Sprintf("%d. %s\n", i+1, song.DiscordSong.TitleTrack)
		if i > 15 && len(songs) > 20 {
			message += fmt.Sprintf("... y %d más.", len(songs)-(i+1))
			break
		}
	}

	logger.Debug("Mostrando lista de reproducción", zap.Int("total_canciones", len(songs)))
	if err := h.messenger.Respond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{Title: "🎵 Lista de reproducción:", Description: message},
			},
		},
	}); err != nil {
		logger.Error("Error al enviar mensaje de lista de reproducción", zap.Error(err))
	}
}

func (h *CommandHandler) RemoveSong(ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := trace.WithTraceID(context.Background())
	logger := h.baseLogger(ctx, ic, "RemoveSong", "remove")

	var position int64
	if len(opt.Options) > 0 && opt.Options[0].Type == discordgo.ApplicationCommandOptionInteger {
		position = opt.Options[0].IntValue()
	} else {
		logger.Warn("Opción de posición para remover canción inválida o faltante")
		h.sendResponse(ic.Interaction, ErrorMessageInvalidRemovePosition)
		return
	}

	guildPlayer, err := h.getGuildPlayerAndLog(ctx, ic, logger)
	if err != nil {
		return
	}

	song, err := guildPlayer.RemoveSong(ctx, int(position))
	if err != nil {
		logger.Error("Error al eliminar la canción de la lista", zap.Error(err), zap.Int64("position", position))
		h.sendResponse(ic.Interaction, ErrorMessageSongRemovalFailed)
		return
	}

	logger.Debug("Canción eliminada exitosamente", zap.String("song_title", song.DiscordSong.TitleTrack))
	h.sendResponse(ic.Interaction, fmt.Sprintf(SuccessMessageSongRemovedFmt, song.DiscordSong.TitleTrack))
}

func (h *CommandHandler) GetPlayingSong(ic *discordgo.InteractionCreate) {
	ctx := trace.WithTraceID(context.Background())
	logger := h.baseLogger(ctx, ic, "GetPlayingSong", "nowplaying")

	guildPlayer, err := h.getGuildPlayerAndLog(ctx, ic, logger)
	if err != nil {
		return
	}

	song, err := guildPlayer.GetPlayedSong(ctx)
	if err != nil {
		logger.Error("Error al obtener la canción actual", zap.Error(err))
		h.sendResponse(ic.Interaction, ErrorMessageGenericPlaylist)
		return
	}

	if song == nil {
		logger.Debug("No hay canción reproduciéndose actualmente")
		h.sendResponse(ic.Interaction, ErrorMessageNoCurrentSong)
		return
	}

	logger.Debug("Mostrando canción actual", zap.String("song_title", song.DiscordSong.TitleTrack))
	h.sendResponse(ic.Interaction, fmt.Sprintf(InfoMessageNowPlayingFmt, song.DiscordSong.TitleTrack))
}

func (h *CommandHandler) PauseSong(ic *discordgo.InteractionCreate) {
	ctx := trace.WithTraceID(context.Background())
	logger := h.baseLogger(ctx, ic, "PauseSong", "pause")

	guildPlayer, err := h.getGuildPlayerAndLog(ctx, ic, logger)
	if err != nil {
		return
	}

	if err := guildPlayer.Pause(ctx); err != nil {
		logger.Error("Error al pausar la reproducción", zap.Error(err))
		h.sendResponse(ic.Interaction, ErrorMessageGenericPause)
		return
	}

	logger.Debug("Reproducción pausada exitosamente")
	h.sendResponse(ic.Interaction, SuccessMessagePaused)
}

func (h *CommandHandler) ResumeSong(ic *discordgo.InteractionCreate) {
	ctx := trace.WithTraceID(context.Background())
	logger := h.baseLogger(ctx, ic, "ResumeSong", "resume")

	guildPlayer, err := h.getGuildPlayerAndLog(ctx, ic, logger)
	if err != nil {
		return
	}

	if err := guildPlayer.Resume(ctx); err != nil {
		logger.Error("Error al reanudar la reproducción", zap.Error(err))
		h.sendResponse(ic.Interaction, ErrorMessageGenericResume)
		return
	}

	logger.Debug("Reproducción reanudada exitosamente")
	h.sendResponse(ic.Interaction, SuccessMessageResumed)
}

func (h *CommandHandler) isUserInVoiceChannel(ctx context.Context, s *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.VoiceState, bool) {
	logger := h.logger.With(
		zap.String("component", "CommandHandlerUtil"),
		zap.String("method", "isUserInVoiceChannel"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", ic.GuildID),
		zap.String("user_id", ic.Member.User.ID),
	)

	if ic.Member == nil {
		logger.Warn("La interacción no tiene información del miembro (probablemente DM o error)")
		h.sendResponse(ic.Interaction, ErrorMessageNotInVoiceChannel)
		return nil, false
	}

	vs, err := s.State.VoiceState(ic.GuildID, ic.Member.User.ID)
	if err != nil {
		logger.Info("Usuario no encontrado en un canal de voz (s.State.VoiceState falló o devolvió nil)", zap.Error(err))
		h.sendResponse(ic.Interaction, ErrorMessageNotInVoiceChannel)
		return nil, false
	}

	if vs.ChannelID == "" {
		logger.Info("Usuario encontrado pero no está conectado a un canal de voz (ChannelID vacío)")
		h.sendResponse(ic.Interaction, ErrorMessageNotInVoiceChannel)
		return nil, false
	}

	logger.Debug("Usuario encontrado en canal de voz", zap.String("voice_channel_id", vs.ChannelID))
	return vs, true
}
