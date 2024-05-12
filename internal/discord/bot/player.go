package bot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"io"
	"time"
)

var (
	// ErrNoSongs indica que no hay canciones disponibles.
	ErrNoSongs = errors.New("canción no disponible")
	// ErrRemoveInvalidPosition indica que la posición de eliminación de la canción es inválida.
	ErrRemoveInvalidPosition = errors.New("posición inválida")
)

// Trigger representa un disparador para comandos relacionados con la reproducción de música.
type Trigger struct {
	Command        string
	VoiceChannelID *string
	TextChannelID  *string
}

// DCADataGetter es una función para obtener datos de audio codificados en DCA para una canción específica.
type DCADataGetter func(ctx context.Context, song *voice.Song) (io.Reader, error)

// GuildPlayerState define métodos para gestionar la lista de reproducción de música.
type GuildPlayerState interface {
	PrependSong(*voice.Song) error
	AppendSong(*voice.Song) error
	RemoveSong(int) (*voice.Song, error)
	ClearPlaylist() error
	GetSongs() ([]*voice.Song, error)
	PopFirstSong() (*voice.Song, error)
	SetCurrentSong(*voice.PlayedSong) error
	GetCurrentSong() (*voice.PlayedSong, error)
	SetVoiceChannel(string) error
	GetVoiceChannel() (string, error)
	SetTextChannel(string) error
	GetTextChannel() (string, error)
}

// GuildPlayer es el reproductor de música para un servidor específico en Discord.
type GuildPlayer struct {
	state           GuildPlayerState
	ctx             context.Context
	triggerCh       chan Trigger
	session         voice.VoiceChatSession
	songCtxCancel   context.CancelFunc
	dCADataGetter   DCADataGetter
	audioBufferSize int
	logger          *zap.Logger
	voiceChannelMap map[string]VoiceChannelInfo
	message         voice.ChatMessageSender
}

// VoiceChannelInfo contiene información sobre un canal de voz y su estado.
type VoiceChannelInfo struct {
	GuildID         string
	BotID           string
	GuildName       string
	VoiceChannelID  string
	TextChannelID   string
	TextChannelName string
	Members         []*discordgo.Member
	PlayingSong     *voice.PlayMessage
	LastUpdated     time.Time
}

// NewGuildPlayer crea una nueva instancia de GuildPlayer con los parámetros proporcionados.
func NewGuildPlayer(ctx context.Context, session voice.VoiceChatSession, state GuildPlayerState, dCADataGetter DCADataGetter, message voice.ChatMessageSender) *GuildPlayer {
	return &GuildPlayer{
		ctx:             ctx,
		state:           state,
		triggerCh:       make(chan Trigger),
		session:         session,
		logger:          zap.NewNop(),
		dCADataGetter:   dCADataGetter,
		audioBufferSize: 1024 * 1024, // 1 MiB
		voiceChannelMap: make(map[string]VoiceChannelInfo),
		message:         message,
	}
}

// WithLogger establece el logger para el GuildPlayer y devuelve el mismo GuildPlayer.
func (p *GuildPlayer) WithLogger(l *zap.Logger) *GuildPlayer {
	p.logger = l
	return p
}

// PrintVoiceChannelInfo imprime la información sobre los servidores y los canales de voz donde se está usando el bot.
func (p *GuildPlayer) PrintVoiceChannelInfo() {
	for _, info := range p.voiceChannelMap {
		fmt.Printf("Servidor: %s (%s)\n", info.GuildName, info.GuildID)
		fmt.Printf("ID Canal de voz: %s\n", info.VoiceChannelID)
		fmt.Printf("Nombre del canal de texto: %s\n", info.TextChannelName)
		fmt.Printf("ID del bot: %s\n", info.BotID)
		fmt.Printf("Actualizacion: %v\n", info.LastUpdated.Local())
		fmt.Println("Miembros:")
		for _, member := range info.Members {
			fmt.Printf("- %s (%s)\n", member.User.Username, member.User.ID)
		}
		fmt.Println("Canción reproduciéndose:")
		if currentSong, err := p.state.GetCurrentSong(); err != nil {
			fmt.Println("Error al obtener la canción actual:", err)
		} else if currentSong != nil {
			fmt.Printf("- %s\n", currentSong.Song.Title)
		} else {
			fmt.Println("- No hay canción reproduciéndose")
		}
		fmt.Println()
	}
}

// UpdateVoiceState actualiza el mapa de información sobre los canales de voz.
func (p *GuildPlayer) UpdateVoiceState(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	// Obtener información sobre el servidor
	guild, err := s.State.Guild(vs.GuildID)
	if err != nil {
		p.logger.Error("Error al obtener información del servidor", zap.Error(err))
		return
	}

	// Verificar si el bot está en el canal de voz
	var voiceChannelID string
	for _, vs := range guild.VoiceStates {
		if vs.UserID == s.State.User.ID {
			voiceChannelID = vs.ChannelID
			break
		}
	}

	if voiceChannelID == "" {
		p.logger.Info("El bot no está en ningún canal de voz en este servidor")
		return
	}

	// Obtener información sobre el canal de voz
	channel, err := s.State.Channel(voiceChannelID)
	if err != nil {
		p.logger.Error("Error al obtener información del canal de voz", zap.Error(err))
		return
	}

	// Obtener los miembros presentes en el canal de voz
	var members []*discordgo.Member
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == voiceChannelID {
			member, err := s.State.Member(guild.ID, vs.UserID)
			if err != nil {
				p.logger.Error("Error al obtener información del miembro", zap.Error(err))
			} else {
				members = append(members, member)
			}
		}
	}

	// Actualizar el mapa de canales de voz solo si es una nueva entrada
	p.voiceChannelMap[vs.GuildID] = VoiceChannelInfo{
		GuildID:         vs.GuildID,
		GuildName:       guild.Name,
		VoiceChannelID:  voiceChannelID,
		TextChannelName: channel.Name,
		Members:         members,
		LastUpdated:     time.Now(),
		BotID:           vs.Member.User.ID,
	}
}

// Close cierra el reproductor de música.
func (p *GuildPlayer) Close() error {
	p.songCtxCancel()
	return p.session.Close()
}

// GetVoiceChannelInfo devuelve el mapa con toda la información de los canales de voz y su estado.
func (p *GuildPlayer) GetVoiceChannelInfo() map[string]VoiceChannelInfo {
	return p.voiceChannelMap
}

// StartListeningEvents inicia la escucha de eventos relevantes.
func (p *GuildPlayer) StartListeningEvents(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		p.UpdateVoiceState(s, vs)
	})
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.PrintVoiceChannelInfo()
			case <-p.ctx.Done():
				return
			}
		}
	}()
	p.logger.Debug("Comenzando a escuchar eventos relevantes")
}

// SendMessage envía un mensaje al canal de texto del servidor.
func (p *GuildPlayer) SendMessage(message string) {
	channel, err := p.state.GetTextChannel()
	if err != nil {
		p.logger.Error("Error al obtener el canal de texto para enviar mensaje", zap.Error(err))
		return
	}

	if err := p.message.SendMessage(channel, message); err != nil {
		p.logger.Error("Error al enviar mensaje al canal de texto", zap.Error(err))
	}
	p.logger.Debug("Mensaje enviado al canal de texto")
}

// AddSong agrega una o más canciones a la lista de reproducción.
func (p *GuildPlayer) AddSong(textChannelID, voiceChannelID *string, songs ...*voice.Song) error {
	for _, song := range songs {
		if err := p.state.AppendSong(song); err != nil {
			p.logger.Error("Error al agregar canción a la lista de reproducción", zap.Error(err))
			return fmt.Errorf("al agregar canción: %w", err)
		}
	}

	go func() {
		p.triggerCh <- Trigger{
			Command:        "play",
			VoiceChannelID: voiceChannelID,
			TextChannelID:  textChannelID,
		}
	}()

	p.logger.Debug("Canciones agregadas a la lista de reproducción", zap.Int("cantidad", len(songs)))
	return nil
}

// SkipSong salta la canción actual.
func (p *GuildPlayer) SkipSong() {
	if p.songCtxCancel != nil {
		p.songCtxCancel()
		p.logger.Debug("Canción actual saltada")
	}
}

// Stop detiene la reproducción y limpia la lista de reproducción.
func (p *GuildPlayer) Stop() error {
	if err := p.state.ClearPlaylist(); err != nil {
		p.logger.Error("Error al limpiar la lista de reproducción", zap.Error(err))
		return fmt.Errorf("al limpiar la lista de reproducción: %w", err)
	}

	if p.songCtxCancel != nil {
		p.songCtxCancel()
		p.logger.Debug("Reproducción detenida y lista de reproducción limpia")
	}

	return nil
}

// RemoveSong elimina una canción de la lista de reproducción por posición.
func (p *GuildPlayer) RemoveSong(position int) (*voice.Song, error) {
	song, err := p.state.RemoveSong(position)
	if err != nil {
		p.logger.Error("Error al eliminar canción de la lista de reproducción", zap.Error(err))
		return nil, fmt.Errorf("al eliminar canción: %w", err)
	}

	p.logger.Debug("Canción eliminada de la lista de reproducción", zap.String("título", song.Title))
	return song, nil
}

// GetPlaylist obtiene la lista de reproducción actual.
func (p *GuildPlayer) GetPlaylist() ([]string, error) {
	songs, err := p.state.GetSongs()
	if err != nil {
		p.logger.Error("Error al obtener la lista de reproducción", zap.Error(err))
		return nil, fmt.Errorf("al obtener canciones: %w", err)
	}

	playlist := make([]string, len(songs))
	for i, song := range songs {
		playlist[i] = song.GetHumanName()
	}

	p.logger.Debug("Lista de reproducción obtenida", zap.Int("cantidad", len(playlist)))
	return playlist, nil
}

// GetPlayedSong obtiene la canción que se está reproduciendo actualmente.
func (p *GuildPlayer) GetPlayedSong() (*voice.PlayedSong, error) {
	currentSong, err := p.state.GetCurrentSong()
	if err != nil {
		p.logger.Error("Error al obtener la canción que se está reproduciendo actualmente", zap.Error(err))
		return nil, fmt.Errorf("al obtener la canción que se está reproduciendo actualmente: %w", err)
	}
	p.logger.Debug("Canción que se está reproduciendo actualmente obtenida")
	return currentSong, nil
}

// JoinVoiceChannel se une al canal de voz especificado.
func (p *GuildPlayer) JoinVoiceChannel(channelID, textChannelID string) {
	p.triggerCh <- Trigger{
		Command:        "join",
		VoiceChannelID: &channelID,
		TextChannelID:  &textChannelID,
	}
	p.logger.Debug("Intentando unirse al canal de voz", zap.String("canal", channelID))
}

// LeaveVoiceChannel abandona el canal de voz actual.
func (p *GuildPlayer) LeaveVoiceChannel() error {
	err := p.session.LeaveVoiceChannel()
	if err != nil {
		p.logger.Error("Error al salir del canal de voz", zap.Error(err))
		return err
	}
	p.logger.Debug("Saliendo del canal de voz")
	return nil
}

// Run inicia el bucle principal del reproductor de música.
func (p *GuildPlayer) Run(ctx context.Context) error {
	currentSong, err := p.state.GetCurrentSong()
	if err != nil {
		p.logger.Info("falló al obtener la canción actual", zap.Error(err))
	} else if currentSong != nil {
		currentSong.StartPosition += currentSong.Position

		if err := p.state.PrependSong(&currentSong.Song); err != nil {
			p.logger.Info("falló al agregar la canción actual en la lista de reproducción", zap.Error(err))
		}
	}

	songs, err := p.state.GetSongs()
	if err != nil {
		return fmt.Errorf("al obtener canciones: %w", err)
	}

	if len(songs) > 0 {
		voiceChannel, err := p.state.GetVoiceChannel()
		if err != nil {
			return fmt.Errorf("al obtener el canal de voz: %w", err)
		}
		textChannel, err := p.state.GetTextChannel()
		if err != nil {
			return fmt.Errorf("al obtener el canal de texto: %w", err)
		}

		go func() {
			p.triggerCh <- Trigger{
				Command:        "play",
				VoiceChannelID: &voiceChannel,
				TextChannelID:  &textChannel,
			}
		}()
	}

	for {
		p.logger.Debug("Esperando triggers")
		select {
		case <-ctx.Done():
			return nil
		case trigger := <-p.triggerCh:
			switch trigger.Command {
			case "play":
				if trigger.TextChannelID != nil {
					if err := p.state.SetTextChannel(*trigger.TextChannelID); err != nil {
						return fmt.Errorf("al establecer el canal de texto: %w", err)
					}
				}
				if trigger.VoiceChannelID != nil {
					if err := p.state.SetVoiceChannel(*trigger.VoiceChannelID); err != nil {
						return fmt.Errorf("al establecer el canal de voz: %w", err)
					}
				}

				songs, err := p.state.GetSongs()
				if err != nil {
					p.logger.Error("falló al obtener canciones", zap.Error(err))
					continue
				}

				if len(songs) == 0 {
					continue
				}

				if err := p.playPlaylist(ctx); err != nil {
					p.logger.Error("falló al reproducir la lista de reproducción", zap.Error(err))
				}
			}
		}
	}
}

// playPlaylist reproduce la lista de reproducción de canciones.
func (p *GuildPlayer) playPlaylist(ctx context.Context) error {
	p.logger.Debug("playPlaylist iniciado")
	voiceChannel, err := p.state.GetVoiceChannel()
	if err != nil {
		return fmt.Errorf("al obtener el canal de voz: %w", err)
	}

	textChannel, err := p.state.GetTextChannel()
	if err != nil {
		return fmt.Errorf("al obtener el canal de texto: %w", err)
	}

	p.logger.Debug("uniéndose al canal de voz", zap.String("canal", voiceChannel))
	if err := p.session.JoinVoiceChannel(voiceChannel); err != nil {
		return fmt.Errorf("falló al unirse al canal de voz: %w", err)
	}

	defer func() {
		p.logger.Debug("saliendo del canal de voz", zap.String("canal", voiceChannel))
		if err := p.session.LeaveVoiceChannel(); err != nil {
			p.logger.Error("falló al salir del canal de voz", zap.Error(err))
		}
	}()

	for {
		song, err := p.state.PopFirstSong()
		if errors.Is(err, ErrNoSongs) {
			p.logger.Debug("la lista de reproducción está vacía")
			break
		}
		if err != nil {
			return fmt.Errorf("al obtener la primera canción: %w", err)
		}

		if err := p.state.SetCurrentSong(&voice.PlayedSong{Song: *song}); err != nil {
			return fmt.Errorf("al establecer la canción actual: %w", err)
		}

		var songCtx context.Context
		songCtx, p.songCtxCancel = context.WithCancel(ctx)

		logger := p.logger.With(zap.String("título", song.Title), zap.String("URL", song.URL))

		playMsgID, err := p.message.SendPlayMessage(textChannel, &voice.PlayMessage{
			Song: song,
		})
		if err != nil {
			return fmt.Errorf("al enviar el mensaje con el nombre de la canción: %w", err)
		}

		dcaData, err := p.dCADataGetter(songCtx, song)
		if err != nil {
			return fmt.Errorf("al obtener datos DCA de la canción %v: %w", song, err)
		}

		audioReader := bufio.NewReaderSize(dcaData, p.audioBufferSize)
		logger.Debug("enviando flujo de audio")
		if err := p.session.SendAudio(songCtx, audioReader, func(d time.Duration) {
			if err := p.state.SetCurrentSong(&voice.PlayedSong{Song: *song, Position: d}); err != nil {
				logger.Error("falló al establecer la posición actual de la canción", zap.Error(err))
			}
			if err := p.message.EditPlayMessage(textChannel, playMsgID, &voice.PlayMessage{Song: song, Position: d}); err != nil {
				logger.Error("falló al editar el mensaje", zap.Error(err))
			}

		}); err != nil {
			return fmt.Errorf("al enviar datos de audio: %w", err)
		}

		if err := p.message.EditPlayMessage(textChannel, playMsgID, &voice.PlayMessage{Song: song, Position: song.Duration}); err != nil {
			logger.Error("falló al editar el mensaje", zap.Error(err))
		}

		if err := p.state.SetCurrentSong(nil); err != nil {
			return fmt.Errorf("al establecer la canción actual: %w", err)
		}
		logger.Debug("reproducción detenida")

		time.Sleep(250 * time.Millisecond)
	}
	p.logger.Debug("playPlaylist finalizado")
	return nil
}
