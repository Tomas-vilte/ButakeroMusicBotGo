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
	ErrorMessageSongRemovalFailed = "❌ No se pudo eliminar la canción. Verifica la posición"
	ErrorMessageNoCurrentSong     = "🔇 No se está reproduciendo ninguna canción actualmente"
)

type CommandHandler struct {
	Storage      ports.InteractionStorage
	Logger       logging.Logger
	SongService  ports.SongService
	messenger    ports.DiscordMessenger
	EventHandler *events.EventHandler
}

func NewCommandHandler(
	storage ports.InteractionStorage,
	logger logging.Logger,
	songService ports.SongService,
	messenger ports.DiscordMessenger,
	eventHandler *events.EventHandler,
) *CommandHandler {
	return &CommandHandler{
		Storage:      storage,
		Logger:       logger,
		SongService:  songService,
		messenger:    messenger,
		EventHandler: eventHandler,
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
	domainInteraction := toDomainInteraction(ic.Interaction)

	if err := h.messenger.Respond(domainInteraction, entity.InteractionResponse{
		Type:    entity.InteractionResponseChannelMessageWithSource,
		Content: "🔍 Buscando tu canción... Esto puede tomar unos momentos.",
	}); err != nil {
		h.Logger.Error("Error al enviar la respuesta inicial", zap.Error(err))
		return
	}

	go func() {
		song, err := h.SongService.GetOrDownloadSong(context.Background(), input, "youtube")
		if err != nil {
			h.Logger.Error("Error al obtener canción", zap.Error(err))
			if err := h.messenger.EditOriginalResponse(domainInteraction, &entity.WebhookEdit{
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
			if err := h.messenger.EditOriginalResponse(domainInteraction, &entity.WebhookEdit{
				Content: shared.StringPtr(ErrorMessageFailedToAddSong),
			}); err != nil {
				h.Logger.Error("Error al actualizar mensaje de error", zap.Error(err))
			}
			return
		}

		h.Logger.Info("Canción agregada a la cola", zap.String("título", song.TitleTrack))

		if err := h.messenger.EditOriginalResponse(domainInteraction, &entity.WebhookEdit{
			Content: shared.StringPtr("✅ Canción agregada a la cola: " + song.TitleTrack),
		}); err != nil {
			h.Logger.Error("Error al actualizar mensaje de confirmación", zap.Error(err))
		}
	}()
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

	domainInteraction := toDomainInteraction(ic.Interaction)

	if err := h.messenger.Respond(domainInteraction, entity.InteractionResponse{
		Type:    entity.InteractionResponseChannelMessageWithSource,
		Content: "⏹️ Reproducción detenida",
	}); err != nil {
		h.Logger.Error("Error al enviar mensaje de confirmación", zap.Error(err))
	}
}

func (h *CommandHandler) isUserInVoiceChannel(s *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.VoiceState, bool) {
	if ic.Member == nil {
		domainInteraction := toDomainInteraction(ic.Interaction)
		if err := h.messenger.RespondWithMessage(domainInteraction, ErrorMessageNotInVoiceChannel); err != nil {
			h.Logger.Error("Error al enviar mensaje de error de canal de voz", zap.Error(err))
		}
		return nil, false
	}

	domainInteraction := toDomainInteraction(ic.Interaction)

	vs, err := s.State.VoiceState(ic.GuildID, ic.Member.User.ID)
	if err != nil {
		if err := h.messenger.Respond(domainInteraction, entity.InteractionResponse{
			Type:    entity.InteractionResponseChannelMessageWithSource,
			Content: ErrorMessageNotInVoiceChannel,
		}); err != nil {
			h.Logger.Error("Error al enviar mensaje de error de canal de voz", zap.Error(err))
		}
		return nil, false
	}

	if vs == nil || vs.ChannelID == "" {
		if err := h.messenger.Respond(domainInteraction, entity.InteractionResponse{
			Type:    entity.InteractionResponseChannelMessageWithSource,
			Content: ErrorMessageNotInVoiceChannel,
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

	domainInteraction := toDomainInteraction(ic.Interaction)

	if err := h.messenger.Respond(domainInteraction, entity.InteractionResponse{
		Type: entity.InteractionResponseChannelMessageWithSource,
		Embeds: []*entity.Embed{
			{Title: "Lista de reproducción:", Description: message},
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

	domainInteraction := toDomainInteraction(ic.Interaction)

	if err := h.messenger.Respond(domainInteraction, entity.InteractionResponse{
		Type:    entity.InteractionResponseChannelMessageWithSource,
		Content: fmt.Sprintf("🗑️ Canción **%s** eliminada de la lista", song.TitleTrack),
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
	domainInteraction := toDomainInteraction(ic.Interaction)

	if err := h.messenger.Respond(domainInteraction, entity.InteractionResponse{
		Type:    entity.InteractionResponseChannelMessageWithSource,
		Content: fmt.Sprintf("🎵 Reproduciendo: %s", song.DiscordSong.TitleTrack),
	}); err != nil {
		h.Logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

func (h *CommandHandler) respondWithError(ic *discordgo.InteractionCreate, message string) {
	domainInteraction := toDomainInteraction(ic.Interaction)
	if err := h.messenger.RespondWithMessage(domainInteraction, message); err != nil {
		h.Logger.Error("Error al enviar mensaje de error", zap.Error(err))
	}
}

func (h *CommandHandler) handleSongRequest(s *discordgo.Session, ic *discordgo.InteractionCreate, g *discordgo.Guild, vs *discordgo.VoiceState, input string) {
	domainInteraction := toDomainInteraction(ic.Interaction)

	if err := h.messenger.Respond(domainInteraction, entity.InteractionResponse{
		Type:    entity.InteractionResponseDeferredChannelMessageWithSource,
		Content: "🔍 Buscando tu canción... Esto puede tomar unos momentos.",
	}); err != nil {
		h.Logger.Error("Error al enviar la respuesta inicial", zap.Error(err))
		return
	}

	go func() {
		song, err := h.SongService.GetOrDownloadSong(context.Background(), input, "youtube")
		if err != nil {
			h.Logger.Error("Error al obtener canción", zap.Error(err))
			if err := h.messenger.EditOriginalResponse(domainInteraction, &entity.WebhookEdit{
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
			if err := h.messenger.EditOriginalResponse(domainInteraction, &entity.WebhookEdit{
				Content: shared.StringPtr(ErrorMessageFailedToAddSong),
			}); err != nil {
				h.Logger.Error("Error al actualizar mensaje de error", zap.Error(err))
			}
			return
		}

		h.Logger.Info("Canción agregada a la cola", zap.String("título", song.TitleTrack))

		if err := h.messenger.EditOriginalResponse(domainInteraction, &entity.WebhookEdit{
			Content: shared.StringPtr("✅ Canción agregada a la cola: " + song.TitleTrack),
		}); err != nil {
			h.Logger.Error("Error al actualizar mensaje de confirmación", zap.Error(err))
		}
	}()
}

func toDomainInteraction(discordInteraction *discordgo.Interaction) *entity.Interaction {
	if discordInteraction == nil {
		return nil
	}

	var member *entity.Member
	if discordInteraction.Member != nil && discordInteraction.Member.User != nil {
		member = &entity.Member{
			UserID:   discordInteraction.Member.User.ID,
			Username: discordInteraction.Member.User.Username,
		}
	}

	return &entity.Interaction{
		ID:        discordInteraction.ID,
		AppID:     discordInteraction.AppID,
		ChannelID: discordInteraction.ChannelID,
		GuildID:   discordInteraction.GuildID,
		Member:    member,
		Token:     discordInteraction.Token,
	}
}
