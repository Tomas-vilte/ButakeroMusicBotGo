package commands

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/events"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
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

type CommandHandler struct {
	EventHandler *events.EventHandler
	Storage      ports.InteractionStorage
	Logger       logging.Logger
	SongService  ports.SongService
}

func NewCommandHandler(
	eventHandler *events.EventHandler,
	storage ports.InteractionStorage,
	logger logging.Logger,
	songService ports.SongService,
) *CommandHandler {
	return &CommandHandler{
		EventHandler: eventHandler,
		Storage:      storage,
		Logger:       logger,
		SongService:  songService,
	}
}

func (h *CommandHandler) PlaySong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	vs, ok := h.isUserInVoiceChannel(s, ic)
	if !ok {
		return
	}

	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		h.Logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}

	input := opt.Options[0].StringValue()
	h.handleSongRequest(s, ic, g, vs, input)
}

func (h *CommandHandler) AddSong(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	vs, ok := h.isUserInVoiceChannel(s, ic)
	if !ok {
		return
	}

	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		h.Logger.Info("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}

	values := ic.MessageComponentData().Values
	if len(values) == 0 {
		h.respondWithError(ic, ErrorMessageNoSongSelected)
		return
	}

	songs := h.Storage.GetSongList(ic.ChannelID)
	if len(songs) == 0 {
		h.respondWithError(ic, ErrorMessageNoSongsAvailable)
		return
	}

	guildPlayer := h.EventHandler.GetGuildPlayer(events.GuildID(g.ID), s)
	song := &entity.DiscordEntity{
		URL:        values[0],
		TitleTrack: values[0],
	}

	playedSong := &entity.PlayedSong{
		DiscordSong:     song,
		RequestedByName: ic.Member.User.Username,
	}

	if err := guildPlayer.AddSong(&ic.ChannelID, &vs.ChannelID, playedSong); err != nil {
		h.Logger.Error("Error al agregar la canción", zap.Error(err))
		h.respondWithError(ic, ErrorMessageFailedToAddSong)
		return
	}

	if err := h.EventHandler.Messenger().RespondWithMessage(ic.Interaction, "✅ Canción agregada a la cola"); err != nil {
		h.Logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}

	h.Storage.DeleteSongList(ic.ChannelID)
}

func (h *CommandHandler) StopPlaying(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		h.Logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, "Ocurrió un error al obtener la información del servidor")
		return
	}

	guildPlayer := h.EventHandler.GetGuildPlayer(events.GuildID(g.ID), s)

	if err := guildPlayer.Stop(); err != nil {
		h.Logger.Error("Error al detener la reproducción", zap.Error(err))
		h.respondWithError(ic, "Ocurrió un error al detener la reproducción")
		return
	}

	if err := h.EventHandler.Messenger().Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹️ Reproducción detenida",
		},
	}); err != nil {
		h.Logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}

func (h *CommandHandler) isUserInVoiceChannel(s *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.VoiceState, bool) {
	if ic.Member == nil {
		if err := h.EventHandler.Messenger().Respond(ic.Interaction, discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: ErrorMessageNotInVoiceChannel,
			},
		}); err != nil {
			h.Logger.Error("Error al enviar mensaje de error de canal de voz", zap.Error(err))
		}
		return nil, false
	}

	vs, err := s.State.VoiceState(ic.GuildID, ic.Member.User.ID)
	if err != nil {
		if err := h.EventHandler.Messenger().Respond(ic.Interaction, discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: ErrorMessageNotInVoiceChannel,
			},
		}); err != nil {
			h.Logger.Error("Error al enviar mensaje de error de canal de voz", zap.Error(err))
		}
		return nil, false
	}

	if vs == nil || vs.ChannelID == "" {
		if err := h.EventHandler.Messenger().Respond(ic.Interaction, discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: ErrorMessageNotInVoiceChannel,
			},
		}); err != nil {
			h.Logger.Error("Error al enviar mensaje de error de canal de voz", zap.Error(err))
		}
		return nil, false
	}

	return vs, true
}

func (h *CommandHandler) SkipSong(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		h.Logger.Info("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, "Ocurrió un error al obtener la información del servidor")
		return
	}

	guildPlayer := h.EventHandler.GetGuildPlayer(events.GuildID(g.ID), s)
	guildPlayer.SkipSong()
	h.respondWithError(ic, "⏭️ Canción omitida")
}

func (h *CommandHandler) ListPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		h.Logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, "Ocurrió un error al obtener la información del servidor")
		return
	}

	guildPlayer := h.EventHandler.GetGuildPlayer(events.GuildID(g.ID), s)
	songs, err := guildPlayer.GetPlaylist()
	if err != nil {
		h.Logger.Error("Error al obtener la lista de reproducción", zap.Error(err))
		h.respondWithError(ic, "Error al obtener la lista de reproducción")
		return
	}

	if len(songs) == 0 {
		h.respondWithError(ic, "📭 La lista de reproducción está vacía")
		return
	}

	message := "🎵 Lista de reproducción:\n"
	for i, song := range songs {
		message += fmt.Sprintf("%d. %s\n", i+1, song)
	}

	if err := h.EventHandler.Messenger().Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{Title: "Lista de reproducción:", Description: message},
			},
		},
	}); err != nil {
		h.Logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

func (h *CommandHandler) RemoveSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		h.Logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}

	guildPlayer := h.EventHandler.GetGuildPlayer(events.GuildID(g.ID), s)
	position := opt.Options[0].IntValue()

	song, err := guildPlayer.RemoveSong(int(position))
	if err != nil {
		h.Logger.Error("Error al eliminar la canción", zap.Error(err))
		h.respondWithError(ic, ErrorMessageSongRemovalFailed)
		return
	}

	if err := h.EventHandler.Messenger().Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🗑️ Canción **%s** eliminada de la lista", song.TitleTrack),
		},
	}); err != nil {
		h.Logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}

func (h *CommandHandler) GetPlayingSong(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		h.Logger.Error("Error al obtener el servidor", zap.Error(err))
		h.respondWithError(ic, ErrorMessageServerNotFound)
		return
	}

	guildPlayer := h.EventHandler.GetGuildPlayer(events.GuildID(g.ID), s)
	song, err := guildPlayer.GetPlayedSong()
	if err != nil {
		h.Logger.Error("Error al obtener la canción actual", zap.Error(err))
		h.respondWithError(ic, "Error al obtener la información de la canción")
		return
	}

	if song == nil {
		h.respondWithError(ic, ErrorMessageNoCurrentSong)
		return
	}

	if err := h.EventHandler.Messenger().Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🎵 Reproduciendo: %s", song.DiscordSong.TitleTrack),
		},
	}); err != nil {
		h.Logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

func (h *CommandHandler) respondWithError(ic *discordgo.InteractionCreate, message string) {
	if err := h.EventHandler.Messenger().RespondWithMessage(ic.Interaction, message); err != nil {
		h.Logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

func (h *CommandHandler) handleSongRequest(s *discordgo.Session, ic *discordgo.InteractionCreate, g *discordgo.Guild, vs *discordgo.VoiceState, input string) {
	if err := h.EventHandler.Messenger().Respond(ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔍 Buscando tu canción... Esto puede tomar unos momentos.",
		},
	}); err != nil {
		h.Logger.Error("Error al enviar la respuesta inicial", zap.Error(err))
		return
	}

	go func() {
		song, err := h.SongService.GetOrDownloadSong(context.Background(), input, "youtube")
		if err != nil {
			h.Logger.Error("Error al obtener canción", zap.Error(err))
			if err := h.EventHandler.Messenger().EditOriginalResponse(ic.Interaction, &discordgo.WebhookEdit{
				Content: shared.StringPtr("❌ No se pudo encontrar o descargar la canción. Verifica el enlace o inténtalo de nuevo"),
			}); err != nil {
				h.Logger.Error("Error al actualizar mensaje de error", zap.Error(err))
			}
			return
		}

		h.Logger.Info("Canción obtenida o descargada", zap.String("título", song.TitleTrack))
		h.Storage.SaveSongList(ic.ChannelID, []*entity.DiscordEntity{song})

		guildPlayer := h.EventHandler.GetGuildPlayer(events.GuildID(g.ID), s)
		playedSong := &entity.PlayedSong{
			DiscordSong:     song,
			RequestedByName: ic.Member.User.Username,
			RequestedByID:   ic.Member.User.ID,
		}

		if err := guildPlayer.AddSong(&ic.ChannelID, &vs.ChannelID, playedSong); err != nil {
			h.Logger.Error("Error al agregar la canción:", zap.Error(err))
			if err := h.EventHandler.Messenger().EditOriginalResponse(ic.Interaction, &discordgo.WebhookEdit{
				Content: shared.StringPtr(ErrorMessageFailedToAddSong),
			}); err != nil {
				h.Logger.Error("Error al actualizar mensaje de error", zap.Error(err))
			}
			return
		}

		h.Logger.Info("Canción agregada a la cola", zap.String("título", song.TitleTrack))

		if err := h.EventHandler.Messenger().EditOriginalResponse(ic.Interaction, &discordgo.WebhookEdit{
			Content: shared.StringPtr("✅ Canción agregada a la cola: " + song.TitleTrack),
		}); err != nil {
			h.Logger.Error("Error al actualizar mensaje de confirmación", zap.Error(err))
		}
	}()
}
