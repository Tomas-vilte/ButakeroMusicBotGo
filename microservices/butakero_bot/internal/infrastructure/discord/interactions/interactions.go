package interactions

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/application/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/player"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	ErrorMessageNotInVoiceChannel = "❌ Debes estar en un canal de voz"
	ErrorMessageFailedToAddSong   = "❌ No se pudo agregar la canción"
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
	presenceNotifier ports.PresenceSubject
	songService      *service.SongService
}

// NewInteractionHandler crea una nueva instancia de InteractionHandler.
func NewInteractionHandler(
	storage ports.InteractionStorage,
	cfg *config.Config,
	logger logging.Logger,
	discordMessenger ports.DiscordMessenger,
	storageAudio ports.StorageAudio,
	presenceNotifier ports.PresenceSubject,
	songService *service.SongService,
) *InteractionHandler {
	return &InteractionHandler{
		guildsPlayers:    make(map[GuildID]*player.GuildPlayer),
		storage:          storage,
		cfg:              cfg,
		logger:           logger,
		discordMessenger: discordMessenger,
		storageAudio:     storageAudio,
		presenceNotifier: presenceNotifier,
		songService:      songService,
	}
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
	guildPlayer := handler.setupGuildPlayer(GuildID(event.Guild.ID), s)
	handler.guildsPlayers[GuildID(event.Guild.ID)] = guildPlayer
	handler.logger.Info("conectado al servidor", zap.String("guildID", event.Guild.ID))
	go func() {
		if err := guildPlayer.Run(ctx); err != nil {
			handler.logger.Error("error al ejecutar el reproductor", zap.Error(err))
		}
	}()
}

func (handler *InteractionHandler) StartPresenceCheck(s *discordgo.Session) {

	for _, guildPlayer := range handler.guildsPlayers {
		handler.presenceNotifier.AddObserver(guildPlayer)
	}

	s.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		handler.logger.Info("Recibido evento VoiceStateUpdate", zap.String("guildID", vs.GuildID), zap.String("channelID", vs.ChannelID))
		handler.presenceNotifier.NotifyAll(vs)
	})

	handler.logger.Info("Comenzando a escuchar eventos de presencia en el canal de voz")
}

// GuildDelete se llama cuando el bot es removido de un servidor.
func (handler *InteractionHandler) GuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	guildID := GuildID(event.Guild.ID)

	guildPlayer := handler.getGuildPlayer(guildID, s)
	if err := guildPlayer.Close(); err != nil {
		handler.logger.Error("Hubo un error al cerrar el reproductor", zap.Error(err))
	}
	delete(handler.guildsPlayers, guildID)
}

// PlaySong maneja el comando de reproducción de una canción.
func (handler *InteractionHandler) PlaySong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	vs := getUsersVoiceState(g, ic.Member.User)
	if vs == nil {
		if err := handler.discordMessenger.SendText(ic.ChannelID, ErrorMessageNotInVoiceChannel); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	input := opt.Options[0].StringValue()
	song, err := handler.songService.GetOrDownloadSong(context.Background(), input)
	if err != nil {
		handler.logger.Error("Error al obtener canción", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "❌ Error al obtener la canción"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	if err := guildPlayer.AddSong(&ic.ChannelID, &vs.ChannelID, song); err != nil {
		handler.logger.Error("Error al agregar la canción", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, ErrorMessageFailedToAddSong); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.SendText(ic.ChannelID, "✅ Canción agregada a la cola"); err != nil {
		handler.logger.Error("falló al enviar mensaje de confirmación", zap.Error(err))
	}
}

// AddSong maneja la adición de una canción.
func (handler *InteractionHandler) AddSong(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	vs := getUsersVoiceState(g, ic.Member.User)
	if vs == nil {
		if err := handler.discordMessenger.SendText(ic.ChannelID, ErrorMessageNotInVoiceChannel); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	values := ic.MessageComponentData().Values
	if len(values) == 0 {
		if err := handler.discordMessenger.SendText(ic.ChannelID, "No se seleccionó ninguna canción"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	song := &entity.Song{
		URL:   values[0],
		Title: values[0],
	}

	if err := guildPlayer.AddSong(&ic.ChannelID, &vs.ChannelID, song); err != nil {
		handler.logger.Error("Error al agregar la canción", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, ErrorMessageFailedToAddSong); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.SendText(ic.ChannelID, "✅ Canción agregada a la cola"); err != nil {
		handler.logger.Error("falló al enviar mensaje de confirmación", zap.Error(err))
	}
}

// StopPlaying detiene la reproducción de música.
func (handler *InteractionHandler) StopPlaying(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	if err := guildPlayer.Stop(); err != nil {
		handler.logger.Info("falló al detener la reproducción", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Ocurrió un error al detener la reproducción"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.SendText(ic.ChannelID, "⏹️ Reproducción detenida"); err != nil {
		handler.logger.Error("falló al enviar mensaje de confirmación", zap.Error(err))
	}
}

// SkipSong salta la canción actualmente en reproducción.
func (handler *InteractionHandler) SkipSong(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	guildPlayer.SkipSong()
	if err := handler.discordMessenger.SendText(ic.ChannelID, "⏭️ Canción omitida"); err != nil {
		handler.logger.Error("falló al enviar mensaje de confirmación", zap.Error(err))
	}
}

// ListPlaylist lista las canciones en la lista de reproducción actual.
func (handler *InteractionHandler) ListPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	songs, err := guildPlayer.GetPlaylist()
	if err != nil {
		handler.logger.Error("Error al obtener la lista de reproducción", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Error al obtener la lista de reproducción"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if len(songs) == 0 {
		if err := handler.discordMessenger.SendText(ic.ChannelID, "📭 La lista de reproducción está vacía"); err != nil {
			handler.logger.Error("falló al enviar mensaje", zap.Error(err))
		}
		return
	}

	message := "🎵 Lista de reproducción:\n"
	for i, song := range songs {
		message += fmt.Sprintf("%d. %s\n", i+1, song)
	}

	if err := handler.discordMessenger.SendText(ic.ChannelID, message); err != nil {
		handler.logger.Error("falló al enviar la lista de reproducción", zap.Error(err))
	}
}

// RemoveSong elimina una canción de la lista de reproducción.
func (handler *InteractionHandler) RemoveSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	position := opt.IntValue()

	song, err := guildPlayer.RemoveSong(int(position))
	if err != nil {
		handler.logger.Error("Error al eliminar la canción", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Error al eliminar la canción"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.SendText(ic.ChannelID, fmt.Sprintf("🗑️ Canción **%s** eliminada de la lista", song.Title)); err != nil {
		handler.logger.Error("falló al enviar mensaje de confirmación", zap.Error(err))
	}
}

// GetPlayingSong obtiene la canción que se está reproduciendo actualmente.
func (handler *InteractionHandler) GetPlayingSong(s *discordgo.Session, ic *discordgo.InteractionCreate, _ *discordgo.ApplicationCommandInteractionDataOption) {
	g, err := s.State.Guild(ic.GuildID)
	if err != nil {
		handler.logger.Info("falló al obtener el servidor", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Ocurrió un error al obtener la información del servidor"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	guildPlayer := handler.getGuildPlayer(GuildID(g.ID), s)
	song, err := guildPlayer.GetPlayedSong()
	if err != nil {
		handler.logger.Error("Error al obtener la canción actual", zap.Error(err))
		if err := handler.discordMessenger.SendText(ic.ChannelID, "Error al obtener la canción actual"); err != nil {
			handler.logger.Error("falló al enviar mensaje de error", zap.Error(err))
		}
		return
	}

	if song == nil {
		if err := handler.discordMessenger.SendText(ic.ChannelID, "🔇 No se está reproduciendo ninguna canción"); err != nil {
			handler.logger.Error("falló al enviar mensaje", zap.Error(err))
		}
		return
	}

	if err := handler.discordMessenger.SendText(ic.ChannelID, fmt.Sprintf("🎵 Reproduciendo: %s", song.Song.Title)); err != nil {
		handler.logger.Error("falló al enviar mensaje", zap.Error(err))
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

	s.AddHandler(func(session *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		if guildPlayer, ok := handler.guildsPlayers[GuildID(vs.GuildID)]; ok {
			guildPlayer.UpdateVoiceState(session, vs)
		}
	})

	// Registrar el manejador de eventos GuildDelete
	s.AddHandler(handler.GuildDelete)
}
