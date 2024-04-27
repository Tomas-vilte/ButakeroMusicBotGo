package discord

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/music/fetcher"
	"github.com/Tomas-vilte/GoMusicBot/internal/utils"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"strings"
)

// GuildID representa el ID de un servidor de Discord.
type GuildID string

// SongLookuper define la interfaz para buscar canciones.
type SongLookuper interface {
	LookupSongs(ctx context.Context, input string) ([]*bot.Song, error)
}

// InteractionStorage define la interfaz para el almacenamiento de interacciones.
type InteractionStorage interface {
	SaveSongList(channelID string, list []*bot.Song)
	GetSongList(channelID string) []*bot.Song
	DeleteSongList(channelID string)
}

// InteractionHandler maneja las interacciones de Discord.
type InteractionHandler struct {
	ctx           context.Context
	discordToken  string
	guildsPlayers map[GuildID]*bot.GuildPlayer
	songLookuper  SongLookuper
	storage       InteractionStorage
	cfg           *config.Config
	logger        *zap.Logger
}

// NewInteractionHandler crea una nueva instancia de InteractionHandler.
func NewInteractionHandler(ctx context.Context, discordToken string, songLookuper SongLookuper, storage InteractionStorage, cfg *config.Config) *InteractionHandler {
	handler := &InteractionHandler{
		ctx:           ctx,
		discordToken:  discordToken,
		guildsPlayers: make(map[GuildID]*bot.GuildPlayer),
		songLookuper:  songLookuper,
		storage:       storage,
		cfg:           cfg,
		logger:        zap.NewNop(),
	}
	return handler
}

// WithLogger establece el logger para InteractionHandler.
func (handler *InteractionHandler) WithLogger(l *zap.Logger) *InteractionHandler {
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
func (handler *InteractionHandler) GuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	player := handler.setupGuildPlayer(GuildID(event.Guild.ID))
	handler.guildsPlayers[GuildID(event.Guild.ID)] = player
	handler.logger.Info("conectado al servidor", zap.String("guildID", event.Guild.ID))

	//// Iniciar goroutine para monitorear la actividad en los canales de voz
	//go func(guildID string) {
	//	ticker := time.NewTicker(1 * time.Second) // Verificar cada minuto
	//
	//	for {
	//		<-ticker.C // Esperar la señal del ticker
	//		fmt.Println("seso")
	//		// Obtener el reproductor del servidor
	//		player := handler.getGuildPlayer(GuildID(guildID))
	//
	//		// Verificar si hay algún usuario en algún canal de voz que no sea el bot
	//		anyUserInVoice := false
	//		for _, vs := range event.Guild.VoiceStates {
	//			// Verificar si el usuario es un miembro y no el bot
	//			if vs.UserID != "" && vs.UserID != s.State.User.ID {
	//				anyUserInVoice = true
	//				break
	//			}
	//		}
	//
	//		// Si no hay usuarios en ningún canal de voz, detener la reproducción
	//		if !anyUserInVoice {
	//			err := player.Stop()
	//			player.LeaveVoiceChannel()
	//			if err != nil {
	//				handler.logger.Error("falló al detener la reproducción por inactividad", zap.Error(err))
	//			}
	//			break // Salir del bucle una vez que se detiene la reproducción
	//		}
	//	}
	//}(event.Guild.ID)
	go func() {
		if err := player.Run(handler.ctx); err != nil {
			handler.logger.Error("ocurrió un error al ejecutar el reproductor", zap.Error(err))
		}
	}()
}

// GuildDelete se llama cuando el bot es removido de un servidor.
func (handler *InteractionHandler) GuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	guildID := GuildID(event.Guild.ID)

	player := handler.getGuildPlayer(guildID)
	player.Close()
	delete(handler.guildsPlayers, guildID)
}

// PlaySong maneja el comando de reproducción de una canción.
func (handler *InteractionHandler) PlaySong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	logger := handler.logger.With(zap.String("guildID", ic.GuildID))

	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		logger.Info("falló al obtener el servidor", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(opt.Options))
	for _, opt := range opt.Options {
		optionMap[opt.Name] = opt
	}

	input := optionMap["input"].StringValue()

	vs := getUsersVoiceState(g, ic.Member.User)
	if vs == nil {
		InteractionRespondMessage(handler.logger, s, ic.Interaction, ErrorMessageNotInVoiceChannel)
	}

	InteractionRespond(handler.logger, s, ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{GenerateAddingSongEmbed(input, ic.Member)},
		},
	})

	go func(ic *discordgo.InteractionCreate, vs *discordgo.VoiceState) {
		songs, err := handler.songLookuper.LookupSongs(handler.ctx, input)
		if err != nil {
			logger.Info("falló al buscar la metadata de la canción", zap.Error(err), zap.String("input", input))
			FollowupMessageCreate(handler.logger, s, ic.Interaction, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{GenerateFailedToAddSongEmbed(input, ic.Member)},
			})
			return
		}

		memberName := getMemberName(ic.Member)
		for i := range songs {
			songs[i].RequestedBy = &memberName
		}

		if len(songs) == 0 {
			FollowupMessageCreate(handler.logger, s, ic.Interaction, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{GenerateFailedToFindSong(input, ic.Member)},
			})
			return
		}

		if len(songs) == 1 {
			song := songs[0]

			if err := player.AddSong(&ic.ChannelID, &vs.ChannelID, song); err != nil {
				logger.Info("falló al agregar la canción", zap.Error(err), zap.String("input", input))
				FollowupMessageCreate(handler.logger, s, ic.Interaction, &discordgo.WebhookParams{
					Embeds: []*discordgo.MessageEmbed{GenerateFailedToAddSongEmbed(input, ic.Member)},
				})
				return
			}

			FollowupMessageCreate(handler.logger, s, ic.Interaction, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{GenerateAddedSongEmbed(song, ic.Member)},
			})
			return
		}

		handler.storage.SaveSongList(ic.ChannelID, songs)

		FollowupMessageCreate(handler.logger, s, ic.Interaction, &discordgo.WebhookParams{
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
		})

	}(ic, vs)
}

// AddSongOrPlaylist maneja la adición de una canción o lista de reproducción.
func (handler *InteractionHandler) AddSongOrPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	values := ic.MessageComponentData().Values
	if len(values) == 0 {
		InteractionRespondMessage(handler.logger, s, ic.Interaction, "😨 Algo salió mal...")
		return
	}

	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	value := values[0]
	songs := handler.storage.GetSongList(ic.ChannelID)
	if len(songs) == 0 {
		InteractionRespondMessage(handler.logger, s, ic.Interaction, "La interacción ya fue seleccionada")
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))

	var voiceChannelID *string = nil

	for _, vs := range g.VoiceStates {
		if vs.UserID == ic.Member.User.ID {
			voiceChannelID = &vs.ChannelID
			break
		}
	}

	if voiceChannelID == nil {
		InteractionRespondMessage(handler.logger, s, ic.Interaction, ErrorMessageNotInVoiceChannel)
		return
	}

	switch value {
	case "playlist":
		for _, song := range songs {
			if err := player.AddSong(&ic.Message.ChannelID, voiceChannelID, song); err != nil {
				handler.logger.Info("falló al agregar la canción", zap.Error(err), zap.String("input", song.URL))
			}
		}
		InteractionRespondMessage(handler.logger, s, ic.Interaction, fmt.Sprintf("➕ Se añadieron %d canciones a la lista de reproducción", len(songs)))
	default:
		song := songs[0]
		if err := player.AddSong(&ic.Message.ChannelID, voiceChannelID, song); err != nil {
			handler.logger.Info("falló al agregar la canción", zap.Error(err), zap.String("input", song.URL))
			InteractionRespondMessage(handler.logger, s, ic.Interaction, ErrorMessageFailedToAddSong)
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

			InteractionRespond(handler.logger, s, ic.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
		}
	}

	handler.storage.DeleteSongList(ic.ChannelID)
}

// StopPlaying detiene la reproducción de música.
func (handler *InteractionHandler) StopPlaying(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))
	if err := player.Stop(); err != nil {
		handler.logger.Info("falló al detener la reproducción", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	InteractionRespondMessage(handler.logger, s, ic.Interaction, "⏹️  Reproducción detenida")
}

// SkipSong salta la canción actualmente en reproducción.
func (handler *InteractionHandler) SkipSong(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))
	player.SkipSong()

	InteractionRespondMessage(handler.logger, s, ic.Interaction, "⏭️ Canción omitida")
}

// ListPlaylist lista las canciones en la lista de reproducción actual.
func (handler *InteractionHandler) ListPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))
	playlist, err := player.GetPlaylist()
	if err != nil {
		handler.logger.Error("falló al obtener la lista de reproducción", zap.Error(err))
		return
	}

	if len(playlist) == 0 {
		InteractionRespondMessage(handler.logger, s, ic.Interaction, "🫙 La lista de reproducción está vacía")
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

		InteractionRespond(handler.logger, s, ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{Title: "Lista de reproducción:", Description: message},
				},
			},
		})
	}
}

// RemoveSong elimina una canción de la lista de reproducción.
func (handler *InteractionHandler) RemoveSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(opt.Options))
	for _, opt := range opt.Options {
		optionMap[opt.Name] = opt
	}

	position := optionMap["position"].IntValue()

	song, err := player.RemoveSong(int(position))
	if err != nil {
		if errors.Is(err, bot.ErrRemoveInvalidPosition) {
			InteractionRespondMessage(handler.logger, s, ic.Interaction, "🤷🏽 Posición no válida")
			return
		}

		handler.logger.Error("falló al eliminar la canción", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	InteractionRespondMessage(handler.logger, s, ic.Interaction, fmt.Sprintf("🗑️ Canción **%v** eliminada de la lista de reproducción", song.GetHumanName()))
}

// GetPlayingSong obtiene la canción que se está reproduciendo actualmente.
func (handler *InteractionHandler) GetPlayingSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))

	song, err := player.GetPlayedSong()
	if err != nil {
		handler.logger.Info("falló al obtener la canción en reproducción", zap.Error(err))
		InteractionRespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	if song == nil {
		InteractionRespondMessage(handler.logger, s, ic.Interaction, "🔇 No se está reproduciendo ninguna canción en este momento...")
		return
	}

	InteractionRespondMessage(handler.logger, s, ic.Interaction, fmt.Sprintf("🎶 %s", song.GetHumanName()))
}

// setupGuildPlayer configura un reproductor para un servidor dado.
func (handler *InteractionHandler) setupGuildPlayer(guildID GuildID) *bot.GuildPlayer {
	dg, err := discordgo.New("Bot " + handler.discordToken)
	if err != nil {
		handler.logger.Error("falló al crear la sesión de Discord", zap.Error(err))
		return nil
	}

	err = dg.Open()
	if err != nil {
		handler.logger.Error("falló al abrir la sesión de Discord", zap.Error(err))
		return nil
	}

	voiceChat := &DiscordVoiceChatSession{
		discordSession: dg,
		guildID:        string(guildID),
	}
	playlistStore := config.GetPlaylistStore(handler.cfg, string(guildID))

	player := bot.NewGuildPlayer(handler.ctx, voiceChat, string(guildID), playlistStore, fetcher.GetDCAData).WithLogger(handler.logger.With(zap.String("guildID", string(guildID))))
	return player
}

// getGuildPlayer obtiene un reproductor para un servidor dado.
func (handler *InteractionHandler) getGuildPlayer(guildID GuildID) *bot.GuildPlayer {
	player, ok := handler.guildsPlayers[guildID]
	if !ok {
		player = handler.setupGuildPlayer(guildID)
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
