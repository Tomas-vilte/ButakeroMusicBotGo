package discord

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/cache"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/discordmessenger"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/observer"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice/codec"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/metrics"
	"github.com/Tomas-vilte/GoMusicBot/internal/music/fetcher"
	"github.com/Tomas-vilte/GoMusicBot/internal/services/providers"
	"github.com/Tomas-vilte/GoMusicBot/internal/storage/s3_audio"
	"github.com/Tomas-vilte/GoMusicBot/internal/utils"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"strings"
)

// GuildID representa el ID de un servidor de Discord.
type GuildID string

// InteractionHandler maneja las interacciones de Discord.
type InteractionHandler struct {
	guildsPlayers       map[GuildID]*bot.GuildPlayer
	songLookup          fetcher.SongLooker
	storage             InteractionStorage
	cfg                 *config.Config
	logger              logging.Logger
	responseHandler     ResponseHandler
	session             SessionService
	commandUsageCounter metrics.CustomMetric
	realYoutubeClient   providers.YouTubeService
	caching             cache.Manager
	audioCaching        cache.AudioCaching
	executorCommand     fetcher.CommandExecutor
	upload              s3_audio.Uploader
	presenceNotifier    *observer.VoicePresenceNotifier
}

// NewInteractionHandler crea una nueva instancia de InteractionHandler.
func NewInteractionHandler(responseHandler ResponseHandler, session SessionService,
	songLooker fetcher.SongLooker,
	storage InteractionStorage,
	cfg *config.Config, logger logging.Logger,
	metricsPrometheus metrics.CustomMetric,
	manager cache.Manager, audioCaching cache.AudioCaching,
	youtubeClient providers.YouTubeService,
	executorCommand fetcher.CommandExecutor,
	upload s3_audio.Uploader,
	presenceNotifier *observer.VoicePresenceNotifier) *InteractionHandler {

	handler := &InteractionHandler{
		guildsPlayers:       make(map[GuildID]*bot.GuildPlayer),
		songLookup:          songLooker,
		storage:             storage,
		cfg:                 cfg,
		logger:              logger,
		responseHandler:     responseHandler,
		session:             session,
		commandUsageCounter: metricsPrometheus,
		caching:             manager,
		audioCaching:        audioCaching,
		realYoutubeClient:   youtubeClient,
		executorCommand:     executorCommand,
		upload:              upload,
		presenceNotifier:    presenceNotifier,
	}
	return handler
}

// WithLogger establece el logger para InteractionHandler.
func (handler *InteractionHandler) WithLogger(l logging.Logger) *InteractionHandler {
	handler.logger = l
	return handler
}

// Ready se llama cuando el bot está listo para recibir interacciones.
func (handler *InteractionHandler) Ready(s *discordgo.Session, event *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, fmt.Sprintf("con tu vieja /%s", handler.cfg.CommandPrefix)); err != nil {
		handler.logger.Error("falló al actualizar el estado del juego", zap.Error(err))
	}
}

// GuildCreate se llama cuando el bot se une a un nuevo servidor.
func (handler *InteractionHandler) GuildCreate(ctx context.Context, s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	player := handler.setupGuildPlayer(GuildID(event.Guild.ID), s)
	handler.guildsPlayers[GuildID(event.Guild.ID)] = player
	handler.logger.Info("conectado al servidor", zap.String("guildID", event.Guild.ID))
	go func() {
		if err := player.Run(ctx); err != nil {
			handler.logger.Error("ocurrió un error al ejecutar el reproductor", zap.Error(err))
		}
	}()
}

func (handler *InteractionHandler) StartPresenceCheck(s *discordgo.Session) {

	for _, player := range handler.guildsPlayers {
		handler.presenceNotifier.AddObserver(player)
	}

	s.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		handler.logger.Info("Recibido evento VoiceStateUpdate", zap.String("guildID", vs.GuildID), zap.String("channelID", vs.ChannelID))
		handler.presenceNotifier.NotifyObservers(vs)
	})

	handler.logger.Info("Comenzando a escuchar eventos de presencia en el canal de voz")
}

// GuildDelete se llama cuando el bot es removido de un servidor.
func (handler *InteractionHandler) GuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	guildID := GuildID(event.Guild.ID)

	player := handler.getGuildPlayer(guildID, s)
	if err := player.Close(); err != nil {
		handler.logger.Error("Hubo un error al cerrar el reproductor", zap.Error(err))
	}
	delete(handler.guildsPlayers, guildID)
}

// PlaySong maneja el comando de reproducción de una canción.
func (handler *InteractionHandler) PlaySong(ctx context.Context, s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	handler.logger.With(zap.String("guildID", ic.GuildID))
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}
	handler.commandUsageCounter.Inc("PlaySong")
	player := handler.getGuildPlayer(GuildID(g.ID), s)
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(opt.Options))
	for _, opt := range opt.Options {
		optionMap[opt.Name] = opt
	}

	input := optionMap["input"].StringValue()
	channelID := ic.ChannelID
	handler.getVoiceChannelMembers(s, channelID)

	vs := getUsersVoiceState(g, ic.Member.User)
	if vs == nil {
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, ErrorMessageNotInVoiceChannel); err != nil {
			handler.logger.Error("falló al responder con el error de no estar en un canal de voz", zap.Error(err))
		}
		return
	}
	if err := handler.responseHandler.Respond(handler.session, ic.Interaction, discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{GenerateAddingSongEmbed(input, ic.Member)},
		},
	}); err != nil {
		handler.logger.Error("fallo al enviar la respuesta diferida", zap.Error(err))
	}

	go func(ic *discordgo.InteractionCreate, vs *discordgo.VoiceState) {
		videoID, err := handler.songLookup.SearchYouTubeVideoID(ctx, input)
		if err != nil {
			handler.logger.Error("Error al buscar el ID del video en YouTube", zap.Error(err), zap.String("input", input))
			if err := handler.responseHandler.CreateFollowupMessage(handler.session, ic.Interaction, discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{GenerateFailedToAddSongEmbed(input, ic.Member)},
			}); err != nil {
				handler.logger.Error("falló al enviar el mensaje de seguimiento de error al buscar el ID del video", zap.Error(err))
			}
			return
		}

		songs, err := handler.songLookup.LookupSongs(ctx, videoID)
		if err != nil {
			handler.logger.Info("falló al buscar la metadata de la canción", zap.Error(err), zap.String("input", input))
			if err := handler.responseHandler.CreateFollowupMessage(handler.session, ic.Interaction, discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{GenerateFailedToAddSongEmbed(input, ic.Member)},
			}); err != nil {
				handler.logger.Error("falló al enviar el mensaje de seguimiento de error al reproducir la cancion", zap.Error(err))
			}
			return
		}

		memberName := getMemberName(ic.Member)
		for i := range songs {
			songs[i].RequestedBy = &memberName
		}

		if len(songs) == 0 {
			if err := handler.responseHandler.CreateFollowupMessage(handler.session, ic.Interaction, discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{GenerateFailedToAddSongEmbed(input, ic.Member)},
			}); err != nil {
				handler.logger.Error("falló al enviar el mensaje de seguimiento de error al agregar la canción", zap.Error(err))
			}
			return
		}

		if len(songs) == 1 {
			song := songs[0]
			if err := player.AddSong(&ic.ChannelID, &vs.ChannelID, song); err != nil {
				handler.logger.Info("falló al agregar la canción", zap.Error(err), zap.String("input", input))
				if err := handler.responseHandler.CreateFollowupMessage(handler.session, ic.Interaction, discordgo.WebhookParams{
					Embeds: []*discordgo.MessageEmbed{GenerateFailedToAddSongEmbed(input, ic.Member)},
				}); err != nil {
					handler.logger.Error("falló al enviar el mensaje de seguimiento de error al agregar la canción", zap.Error(err))
				}
				return
			}
			if err := handler.responseHandler.CreateFollowupMessage(handler.session, ic.Interaction, discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{GenerateAddedSongEmbed(song, ic.Member)},
			}); err != nil {
				handler.logger.Error("falló al enviar el mensaje de seguimiento de canción agregada", zap.Error(err))
			}
			return
		}

		handler.storage.SaveSongList(ic.ChannelID, songs)

		if err := handler.responseHandler.CreateFollowupMessage(handler.session, ic.Interaction, discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{GenerateAskAddPlaylistEmbed(songs, ic.Member)},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID: "add_song_playlist",
							Options: []discordgo.SelectMenuOption{
								{Label: "Agregar canción", Value: "song", Emoji: &discordgo.ComponentEmoji{Name: "🎵"}},
								{Label: "Agregar lista de reproducción completa", Value: "playlist", Emoji: &discordgo.ComponentEmoji{Name: "🎶"}},
							},
						},
					},
				},
			},
		}); err != nil {
			handler.logger.Error("falló al enviar el mensaje de seguimiento de selección de agregar canción o lista de reproducción", zap.Error(err))
		}
	}(ic, vs)
}

// AddSongOrPlaylist maneja la adición de una canción o lista de reproducción.
func (handler *InteractionHandler) AddSongOrPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	values := ic.MessageComponentData().Values
	if len(values) == 0 {
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	value := values[0]
	songs := handler.storage.GetSongList(ic.ChannelID)
	if len(songs) == 0 {
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "La interacción ya fue seleccionada"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID), s)

	var voiceChannelID *string = nil

	for _, vs := range g.VoiceStates {
		if vs.UserID == ic.Member.User.ID {
			voiceChannelID = &vs.ChannelID
			break
		}
	}

	if voiceChannelID == nil {
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, ErrorMessageNotInVoiceChannel); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	switch value {
	case "playlist":
		for _, song := range songs {
			if err := player.AddSong(&ic.Message.ChannelID, voiceChannelID, song); err != nil {
				handler.logger.Info("falló al agregar la canción", zap.Error(err), zap.String("input", song.URL))
			}
		}
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, fmt.Sprintf("➕ Se añadieron %d canciones a la lista de reproducción", len(songs))); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
	default:
		song := songs[0]
		if err := player.AddSong(&ic.Message.ChannelID, voiceChannelID, song); err != nil {
			handler.logger.Info("falló al agregar la canción", zap.Error(err), zap.String("input", song.URL))
			if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, ErrorMessageFailedToAddSong); err != nil {
				handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
			}
		} else {
			embed := &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name: "Añadido a la cola",
				},
				Title: song.GetHumanName(),
				URL:   song.URL,
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Solicitado por %s", *song.RequestedBy),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Duración",
						Value: utils.FmtDuration(song.Duration),
					},
				},
			}

			if song.ThumbnailURL != nil {
				embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: *song.ThumbnailURL,
				}
			}

			if err := handler.responseHandler.Respond(handler.session, ic.Interaction, discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			}); err != nil {
				handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
			}
		}
	}
	handler.storage.DeleteSongList(ic.ChannelID)
}

// StopPlaying detiene la reproducción de música.
func (handler *InteractionHandler) StopPlaying(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID), s)
	handler.commandUsageCounter.Inc("StopPlaying")
	if err := player.Stop(); err != nil {
		handler.logger.Info("falló al detener la reproducción", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}
	if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "⏹️  Reproducción detenida"); err != nil {
		handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
	}
}

// SkipSong salta la canción actualmente en reproducción.
func (handler *InteractionHandler) SkipSong(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID), s)
	player.SkipSong()
	handler.commandUsageCounter.Inc("SkipSong")
	if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "⏭️ Canción omitida"); err != nil {
		handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
	}
}

// ListPlaylist lista las canciones en la lista de reproducción actual.
func (handler *InteractionHandler) ListPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID), s)
	handler.commandUsageCounter.Inc("ListPlaylist")
	playlist, err := player.GetPlaylist()
	if err != nil {
		handler.logger.Error("falló al obtener la lista de reproducción", zap.Error(err))
		return
	}

	if len(playlist) == 0 {
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "🫙 La lista de reproducción está vacía"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
	} else {
		builder := strings.Builder{}

		for idx, song := range playlist {
			line := fmt.Sprintf("%d. %s\n", idx+1, song)

			if len(line)+builder.Len() > 4000 {
				builder.WriteString("...")
				break
			}

			builder.WriteString(fmt.Sprintf("%d. %s\n", idx+1, song))
		}

		message := strings.TrimSpace(builder.String())

		if err := handler.responseHandler.Respond(handler.session, ic.Interaction, discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{Title: "Lista de reproducción:", Description: message},
				},
			},
		}); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
	}
}

// RemoveSong elimina una canción de la lista de reproducción.
func (handler *InteractionHandler) RemoveSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID), s)
	handler.commandUsageCounter.Inc("RemoveSong")
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(opt.Options))
	for _, opt := range opt.Options {
		optionMap[opt.Name] = opt
	}

	position := optionMap["position"].IntValue()

	song, err := player.RemoveSong(int(position))
	if err != nil {
		if errors.Is(err, bot.ErrRemoveInvalidPosition) {
			if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "🤷🏽 Posición no válida"); err != nil {
				handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
			}
			return
		}

		handler.logger.Error("falló al eliminar la canción", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al eliminar la cancion"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, fmt.Sprintf("🗑️ Canción **%v** eliminada de la lista de reproducción", song.GetHumanName())); err != nil {
		handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
	}
}

// GetPlayingSong obtiene la canción que se está reproduciendo actualmente.
func (handler *InteractionHandler) GetPlayingSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID), s)
	handler.commandUsageCounter.Inc("GetPlayingSong")
	song, err := player.GetPlayedSong()
	if err != nil {
		handler.logger.Info("falló al obtener la canción en reproducción", zap.Error(err))
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "Ocurrió un error al obtener la canción en reproducción"); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	if song == nil {
		if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, "🔇 No se está reproduciendo ninguna canción en este momento..."); err != nil {
			handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
		}
		return
	}

	if err := handler.responseHandler.RespondWithMessage(handler.session, ic.Interaction, fmt.Sprintf("🎶 %s", song.GetHumanName())); err != nil {
		handler.logger.Error("falló al responder con el error del servidor", zap.Error(err))
	}
}

// setupGuildPlayer configura un reproductor para un servidor dado.
func (handler *InteractionHandler) setupGuildPlayer(guildID GuildID, dg *discordgo.Session) *bot.GuildPlayer {
	dca := codec.NewDCAStreamerImpl(handler.logger)
	voiceChat := voice.NewChatSessionImpl(dg, string(guildID), dca, handler.logger)
	messageSender := discordmessenger.NewMessageSenderImpl(dg, handler.logger)
	fetcherGetDCA := fetcher.NewYoutubeFetcher(handler.logger, handler.caching, handler.realYoutubeClient, handler.audioCaching, handler.executorCommand, handler.upload)
	songStorage, stateStorage := config.GetPlaylistStore(handler.cfg, string(guildID), handler.logger)
	player := bot.NewGuildPlayer(voiceChat, songStorage, stateStorage, fetcherGetDCA.GetDCAData, messageSender, handler.logger)
	return player
}

// getGuildPlayer obtiene un reproductor para un servidor dado.
func (handler *InteractionHandler) getGuildPlayer(guildID GuildID, dg *discordgo.Session) *bot.GuildPlayer {
	player, ok := handler.guildsPlayers[guildID]
	if !ok {
		player = handler.setupGuildPlayer(guildID, dg)
		handler.guildsPlayers[guildID] = player
	}

	return player
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

	s.AddHandler(func(session *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		if player, ok := handler.guildsPlayers[GuildID(vs.GuildID)]; ok {
			player.UpdateVoiceState(session, vs)
		}
	})

	// Registrar el manejador de eventos GuildDelete
	s.AddHandler(handler.GuildDelete)
}
