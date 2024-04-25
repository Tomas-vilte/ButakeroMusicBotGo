package discord

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/music/fetcher"
	"github.com/Tomas-vilte/GoMusicBot/internal/utils"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"strings"
)

type GuildID string

type SongLookuper interface {
	LookupSongs(ctx context.Context, input string) ([]*bot.Song, error)
}

type InteractionStorage interface {
	SaveSongList(channelID string, list []*bot.Song)
	GetSongList(channelID string) []*bot.Song
	DeleteSongList(channelID string)
}

type InteractionHandler struct {
	ctx                  context.Context
	discordToken         string
	guildsPlayers        map[GuildID]*bot.GuildPlayer
	interactionResponder InteractionResponder
	songLookuper         SongLookuper
	storage              InteractionStorage
	cfg                  *config.Config
	logger               *zap.Logger
}

func NewInteractionHandler(ctx context.Context, discordToken string, interactionResponder InteractionResponder, songLookuper SongLookuper, storage InteractionStorage, cfg *config.Config) *InteractionHandler {
	handler := &InteractionHandler{
		ctx:                  ctx,
		discordToken:         discordToken,
		interactionResponder: interactionResponder,
		guildsPlayers:        make(map[GuildID]*bot.GuildPlayer),
		songLookuper:         songLookuper,
		storage:              storage,
		cfg:                  cfg,
		logger:               zap.NewNop(),
	}
	return handler
}

func (handler *InteractionHandler) WithLogger(l *zap.Logger) *InteractionHandler {
	handler.logger = l
	return handler
}

func (handler *InteractionHandler) Ready(s *discordgo.Session, event *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, fmt.Sprintf("modo sesooo %s", handler.cfg.CommandPrefix)); err != nil {
		handler.logger.Error("Error al actualizar estado de juego", zap.Error(err))
	}
}

func (handler *InteractionHandler) GuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	player := handler.setupGuildPlayer(GuildID(event.Guild.ID))
	handler.guildsPlayers[GuildID(event.Guild.ID)] = player

	handler.logger.Info("conectando al server", zap.String("guild", event.Guild.ID))

	go func() {
		if err := player.Run(handler.ctx); err != nil {
			handler.logger.Error("hubo un error, cuando el jugador estaba corriendo", zap.Error(err))
		}
	}()
}

func (handler *InteractionHandler) GuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	guildID := GuildID(event.Guild.ID)
	player := handler.getGuildPlayer(guildID)
	player.Close()
	delete(handler.guildsPlayers, guildID)
}

func (handler *InteractionHandler) PlaySong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	logger := handler.logger.With(zap.String("guildID", string(ic.GuildID)))

	g, err := s.Guild(ic.GuildID)
	if err != nil {
		logger.Warn("Error al obtener el guild", zap.Error(err))
		handler.interactionResponder.RespondServerError(logger, s, ic.Interaction)
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
		handler.interactionResponder.InteractionRespondMessage(logger, s, ic.Interaction, ErrorMessageNotInVoiceChannel)
	}
	RespondToInteraction(logger, s, ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{GenerateAddingSongEmbed(input, ic.Member)},
		},
	})

	go func(ic *discordgo.InteractionCreate, vs *discordgo.VoiceState) {
		songs, err := handler.songLookuper.LookupSongs(handler.ctx, input)
		if err != nil {
			logger.Info("No se pudieron buscar los metadatos de la canción", zap.Error(err))
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
				logger.Info("no se pudo agregar la canción", zap.Error(err), zap.String("input", input))
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
								{Label: "Agregar canción", Value: "cancion", Emoji: &discordgo.ComponentEmoji{Name: "🎵"}},
								{Label: "Agregar lista de reproducción completa\n", Value: "playlist", Emoji: &discordgo.ComponentEmoji{Name: "🎶"}},
							},
						},
					},
				},
			},
		})

	}(ic, vs)

}

func (handler *InteractionHandler) AddSongOrPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	values := ic.MessageComponentData().Values
	if len(values) == 0 {
		handler.interactionResponder.InteractionRespondMessage(handler.logger, s, ic.Interaction, "😨 Algo salió mal...")
		return
	}

	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("fallo al obtener server", zap.Error(err))
		handler.interactionResponder.RespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	value := values[0]
	songs := handler.storage.GetSongList(ic.ChannelID)
	if len(songs) == 0 {
		handler.interactionResponder.InteractionRespondMessage(handler.logger, s, ic.Interaction, "La interacción ya fue seleccionada")
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
		handler.interactionResponder.InteractionRespondMessage(handler.logger, s, ic.Interaction, "🤷🏽 No estas en un canal de voz kkk, unite a uno para escuchar la musica down.")
		return
	}

	switch value {
	case "playlist":
		for _, song := range songs {
			if err := player.AddSong(&ic.Message.ChannelID, voiceChannelID, song); err != nil {
				handler.logger.Info("no se pudo agregar la canción", zap.Error(err), zap.String("input", song.URL))
			}
		}
		handler.interactionResponder.InteractionRespondMessage(handler.logger, s, ic.Interaction, fmt.Sprintf("➕ Se agregaron %d canciones a la playlist.", len(songs)))
	default:
		song := songs[0]
		if err := player.AddSong(&ic.Message.ChannelID, voiceChannelID, song); err != nil {
			handler.logger.Info("failed to add song", zap.Error(err), zap.String("input", song.URL))
			handler.interactionResponder.InteractionRespondMessage(handler.logger, s, ic.Interaction, "😨 No se pudo agregar la canción")
		} else {
			embed := &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name: "Agregado a la cola",
				},
				Title: song.GetHumanName(),
				URL:   song.URL,
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Pedido por %s", *song.RequestedBy),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Duracion",
						Value: utils.FmtDuration(song.Duration),
					},
				},
			}

			if song.ThumbnailURL != nil {
				embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: *song.ThumbnailURL,
				}
			}

			RespondToInteraction(handler.logger, s, ic.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
		}
	}

	handler.storage.DeleteSongList(ic.ChannelID)
}

func (handler *InteractionHandler) StopPlaying(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("fallo al obtener server", zap.Error(err))
		handler.interactionResponder.RespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))
	if err := player.Stop(); err != nil {
		handler.logger.Info("fallo al pausar cancion", zap.Error(err))
		handler.interactionResponder.RespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	handler.interactionResponder.InteractionRespondMessage(handler.logger, s, ic.Interaction, "⏹️  pausando cancion")
}

func (handler *InteractionHandler) SkipSong(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("fallo al obtener server", zap.Error(err))
		handler.interactionResponder.RespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))
	player.SkipSong()

	handler.interactionResponder.InteractionRespondMessage(handler.logger, s, ic.Interaction, "⏭️ saltando cancion")
}

func (handler *InteractionHandler) ListPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate, acido *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("failed to get guild", zap.Error(err))
		handler.interactionResponder.RespondServerError(handler.logger, s, ic.Interaction)
		return
	}

	player := handler.getGuildPlayer(GuildID(g.ID))
	playlist, err := player.GetPlaylist()
	if err != nil {
		handler.logger.Error("failed to get playlist", zap.Error(err))
		return
	}

	if len(playlist) == 0 {
		handler.interactionResponder.InteractionRespondMessage(handler.logger, s, ic.Interaction, "🫙 esta playlist esta vacia")
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

		RespondToInteraction(handler.logger, s, ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{Title: "Playlist:", Description: message},
				},
			},
		})
	}
}

func getUsersVoiceState(guild *discordgo.Guild, user *discordgo.User) *discordgo.VoiceState {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			return vs
		}
	}
	return nil
}

func (handler *InteractionHandler) setupGuildPlayer(guildID GuildID) *bot.GuildPlayer {
	dg, err := discordgo.New("Bot " + handler.discordToken)
	if err != nil {
		handler.logger.Error("fallo al iniciar session en discord", zap.Error(err))
		return nil
	}

	err = dg.Open()
	if err != nil {
		handler.logger.Error("Error al abrir la conexion a discord", zap.Error(err))
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

func (handler *InteractionHandler) getGuildPlayer(guildID GuildID) *bot.GuildPlayer {
	player, ok := handler.guildsPlayers[guildID]
	if !ok {
		player = handler.setupGuildPlayer(guildID)
		handler.guildsPlayers[guildID] = player
	}

	return player
}
