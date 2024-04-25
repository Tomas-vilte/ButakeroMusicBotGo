package bot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"time"
)

var ErrNoSongs = errors.New("cancion no disponible")

// Trigger representa un disparador para comandos relacionados con la reproduccion de musica.
type Trigger struct {
	Command        string
	VoiceChannelID *string
	TextChannelID  *string
}

// PlayMessage es el mensaje que se enviara al canal de texto para mostrar la cancion que se está reproduciendo actualmente.
type PlayMessage struct {
	Song     *Song
	Position time.Duration
}

// Song representa una cancion que se puede reproducir.
type Song struct {
	Type          string
	Title         string
	URL           string
	Playable      bool
	ThumbnailURL  *string
	Duration      time.Duration
	StartPosition time.Duration
	RequestedBy   *string
}

// GetHumanName devuelve el nombre humano legible de la cancion.
func (s *Song) GetHumanName() string {
	if s.Title != "" {
		return s.Title
	}
	return s.URL
}

// VoiceChatSession define metodos para interactuar con la sesion de voz del bot de Discord.
type VoiceChatSession interface {
	Close() error
	SendMessage(channelID, message string) error
	SendPlayMessage(channelID string, message *PlayMessage) (string, error)
	EditPlayMessage(channelID, messageID string, message *PlayMessage) error
	JoinVoiceChannel(channelID string) error
	LeaveVoiceChannel() error
	SendAudio(ctx context.Context, r io.Reader, positionCallback func(time.Duration)) error
}

// DCADataGetter es una funcion para obtener datos de audio codificados en DCA para una canción especifica.
type DCADataGetter func(ctx context.Context, song *Song) (io.Reader, error)

// PlayedSong representa una cancion que ha sido reproducida.
type PlayedSong struct {
	Song     Song
	Position time.Duration
}

// PlaylistManager define metodos para gestionar la lista de reproduccion de musica.
type PlaylistManager interface {
	PrependSong(*Song) error
	AppendSong(*Song) error
	RemoveSong(int) (*Song, error)
	ClearPlaylist() error
	GetSongs() ([]*Song, error)
	PopFirstSong() (*Song, error)
	SetCurrentSong(*PlayedSong) error
	GetCurrentSong() (*PlayedSong, error)
}

// ChannelManager define metodos para gestionar los canales de voz y de texto.
type ChannelManager interface {
	SetVoiceChannel(string) error
	GetVoiceChannel() (string, error)
	SetTextChannel(string) error
	GetTextChannel() (string, error)
}

// GuildPlayer es el reproductor de musica para un servidor específico en discord.
type GuildPlayer struct {
	ctx             context.Context
	session         VoiceChatSession
	playlistManager PlaylistManager
	channelManager  ChannelManager
	dcaDataGetter   DCADataGetter
	audioBufferSize int
	logger          *zap.Logger
	triggerCh       chan Trigger
	songCtxCancel   context.CancelFunc
}

var (
	ErrRemoveInvalidPosition = errors.New("posicion invalida")
)

// NewGuildPlayer crea una nueva instancia de GuildPlayer con los parametros proporcionados.
func NewGuildPlayer(ctx context.Context, session VoiceChatSession, guildID string, playlistManager PlaylistManager, dcaDataGetter DCADataGetter) *GuildPlayer {
	return &GuildPlayer{
		ctx:             ctx,
		session:         session,
		playlistManager: playlistManager,
		dcaDataGetter:   dcaDataGetter,
		audioBufferSize: 1024 * 1024, //1mib
		logger:          zap.NewNop(),
		triggerCh:       make(chan Trigger),
	}
}

// WithLogger establece el logger para el GuildPlayer.
func (p *GuildPlayer) WithLogger(logger *zap.Logger) *GuildPlayer {
	p.logger = logger
	return p
}

// Close cierra el GuildPlayer.
func (p *GuildPlayer) Close() error {
	p.songCtxCancel()
	return p.session.Close()
}

// SendMessage envia un mensaje al canal de texto.
func (p *GuildPlayer) SendMessage(message string) {
	channel, err := p.channelManager.GetTextChannel()
	if err != nil {
		p.logger.Error("error al obtener el canal de texto", zap.Error(err))
		return
	}
	if err := p.session.SendMessage(channel, message); err != nil {
		p.logger.Error("error al enviar mensaje", zap.Error(err))
	}
}

// AddSong agrega una cancion a la lista de reproduccion y comienza a reproducir si es necesario.
func (p *GuildPlayer) AddSong(textChannelID, voiceChannelID *string, sons ...*Song) error {
	for _, song := range sons {
		if err := p.playlistManager.AppendSong(song); err != nil {
			return fmt.Errorf("error al añadir canción: %w", err)
		}
	}
	go func() {
		p.triggerCh <- Trigger{
			Command:        "play",
			VoiceChannelID: voiceChannelID,
			TextChannelID:  textChannelID,
		}
	}()
	return nil
}

// SkipSong omite la cancion actualmente en reproduccion.
func (p *GuildPlayer) SkipSong() {
	if p.songCtxCancel != nil {
		p.songCtxCancel()
	}
}

// Stop detiene la reproduccion de todas las canciones y limpia la lista de reproduccion.
func (p *GuildPlayer) Stop() error {
	if err := p.playlistManager.ClearPlaylist(); err != nil {
		return fmt.Errorf("error al limpiar la lista de reproducción: %w", err)
	}
	if p.songCtxCancel != nil {
		p.songCtxCancel()
	}
	return nil
}

// RemoveSong elimina una cancion de la lista de reproduccion en la posicion especificada.
func (p *GuildPlayer) RemoveSong(position int) (*Song, error) {
	song, err := p.playlistManager.RemoveSong(position)
	if err != nil {
		return nil, fmt.Errorf("error al eliminar la canción: %w", err)
	}
	return song, nil
}

// GetPlaylist devuelve la lista de reproduccion actual como una lista de nombres de canciones.
func (p *GuildPlayer) GetPlaylist() ([]string, error) {
	songs, err := p.playlistManager.GetSongs()
	if err != nil {
		return nil, fmt.Errorf("error al obtener las canciones: %w", err)
	}
	playlist := make([]string, len(songs))
	for i, song := range songs {
		playlist[i] = song.GetHumanName()
	}
	return playlist, nil
}

// GetPlayedSong devuelve la cancion que se está reproduciendo actualmente.
func (p *GuildPlayer) GetPlayedSong() (*PlayedSong, error) {
	return p.playlistManager.GetCurrentSong()
}

// JoinVoiceChannel se une al canal de voz especificado.
func (p *GuildPlayer) JoinVoiceChannel(channelID, textChannelID string) {
	p.triggerCh <- Trigger{
		Command:        "join",
		VoiceChannelID: &channelID,
		TextChannelID:  &textChannelID,
	}
}

// LeaveVoiceChannel abandona el canal de voz.
func (p *GuildPlayer) LeaveVoiceChannel() {
	p.triggerCh <- Trigger{
		Command: "leave",
	}
}

// Run ejecuta el bucle principal del GuildPlayer.
func (p *GuildPlayer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case trigger := <-p.triggerCh:
			switch trigger.Command {
			case "play":
				if err := p.handlePlayCommand(ctx, trigger.VoiceChannelID, trigger.TextChannelID); err != nil {
					p.logger.Error("error al procesar el comando de reproducción", zap.Error(err))
				}
			}
		}
	}
}

// playPlaylist reproduce la lista de reproduccion actual.
func (p *GuildPlayer) playPlaylist(ctx context.Context) error {
	voiceChannel, err := p.channelManager.GetVoiceChannel()
	if err != nil {
		return fmt.Errorf("error al obtener el canal de voz: %w", err)
	}
	textChannel, err := p.channelManager.GetTextChannel()
	if err != nil {
		return fmt.Errorf("error al obtener el canal de texto: %w", err)
	}
	p.logger.Debug("uniéndose al canal de voz", zap.String("channel", voiceChannel))
	if err := p.session.JoinVoiceChannel(voiceChannel); err != nil {
		return fmt.Errorf("error al unirse al canal de voz: %w", err)
	}
	defer func() {
		p.logger.Debug("saliendo del canal de voz", zap.String("channel", voiceChannel))
		if err := p.session.LeaveVoiceChannel(); err != nil {
			p.logger.Error("error al salir del canal de voz", zap.Error(err))
		}
	}()

	for {
		song, err := p.playlistManager.PopFirstSong()
		if errors.Is(err, ErrNoSongs) {
			p.logger.Debug("la lista de reproduccion esta vacia")
			break
		}
		if err != nil {
			return fmt.Errorf("error al extraer la primera canción: %w", err)
		}
		if err := p.playSong(ctx, song, textChannel); err != nil {
		}
	}
	return nil
}

// handlePlayCommand maneja el comando de reproduccion de la lista de reproduccion.
func (p *GuildPlayer) handlePlayCommand(ctx context.Context, voiceChannelID, textChannelID *string) error {
	if textChannelID != nil {
		if err := p.channelManager.SetTextChannel(*textChannelID); err != nil {
			return fmt.Errorf("error al establecer el canal de texto: %w", err)
		}
	}
	if voiceChannelID != nil {
		if err := p.channelManager.SetVoiceChannel(*voiceChannelID); err != nil {
			return fmt.Errorf("error al establecer el canal de voz: %w", err)
		}
	}
	songs, err := p.playlistManager.GetSongs()
	if err != nil {
		return fmt.Errorf("error al obtener las canciones: %w", err)
	}
	if len(songs) == 0 {
		return nil
	}
	return p.playPlaylist(ctx)
}

// playSong reproduce una cancion.
func (p *GuildPlayer) playSong(ctx context.Context, song *Song, textChannelID string) error {
	logger := p.logger.With(zap.String("title", song.Title), zap.String("url", song.URL))

	playMsgID, err := p.session.SendPlayMessage(textChannelID, &PlayMessage{Song: song})
	if err != nil {
		return fmt.Errorf("error al enviar el mensaje con el nombre de la canción: %w", err)
	}
	dcaData, err := p.dcaDataGetter(ctx, song)
	if err != nil {
		return fmt.Errorf("error al obtener los datos DCA de la canción %v: %w", song, err)
	}
	audioreader := bufio.NewReaderSize(dcaData, p.audioBufferSize)
	logger.Debug("enviando el flujo de audio")
	if err := p.session.SendAudio(ctx, audioreader, func(d time.Duration) {
		if err := p.playlistManager.SetCurrentSong(&PlayedSong{Song: *song, Position: d}); err != nil {
			logger.Error("error al establecer la posición actual de la canción", zap.Error(err))
		}
		if err := p.session.EditPlayMessage(textChannelID, playMsgID, &PlayMessage{Song: song, Position: d}); err != nil {
			logger.Error("error al editar el mensaje", zap.Error(err))
		}
	}); err != nil {
		return fmt.Errorf("error al enviar los datos de audio: %w", err)
	}

	if err := p.session.EditPlayMessage(textChannelID, playMsgID, &PlayMessage{Song: song, Position: song.Duration}); err != nil {
		logger.Error("error al editar el mensaje", zap.Error(err))
	}

	if err := p.playlistManager.SetCurrentSong(nil); err != nil {
		return fmt.Errorf("error al establecer la canción actual: %w", err)
	}
	logger.Debug("Reproduccion detenida")
	time.Sleep(200 * time.Millisecond)

	return nil
}
