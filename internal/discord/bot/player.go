package bot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/discordmessenger"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"io"
	"sync"
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

// GuildPlayer es el reproductor de música para un servidor específico en Discord.
type GuildPlayer struct {
	ctx             context.Context                    // Contexto para la gestión de la vida útil del reproductor.
	triggerCh       chan Trigger                       // Canal para recibir disparadores de comandos relacionados con la reproducción de música.
	session         voice.VoiceChatSession             // Sesión de chat de voz que define métodos para interactuar con la sesión de voz del bot de Discord.
	songCtxCancel   context.CancelFunc                 // Función de cancelación del contexto de la canción actual.
	songStorage     store.SongStorage                  // Almacenamiento de canciones para la lista de reproducción.
	stateStorage    store.StateStorage                 // Almacenamiento de estado para el reproductor de música.
	dCADataGetter   DCADataGetter                      // Función para obtener datos de audio codificados en DCA para una canción específica.
	audioBufferSize int                                // Tamaño del búfer de audio para la transmisión de música.
	logger          logging.Logger                     // Registro de eventos y errores.
	voiceChannelMap map[string]VoiceChannelInfo        // Mapa que contiene información sobre los canales de voz y su estado.
	message         discordmessenger.ChatMessageSender // Interfaz para enviar mensajes de chat a Discord.
	mu              sync.Mutex
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
func NewGuildPlayer(ctx context.Context, session voice.VoiceChatSession, songStorage store.SongStorage, stateStorage store.StateStorage, dCADataGetter DCADataGetter, message discordmessenger.ChatMessageSender, logger logging.Logger) *GuildPlayer {
	return &GuildPlayer{
		ctx:             ctx,
		songStorage:     songStorage,
		stateStorage:    stateStorage,
		triggerCh:       make(chan Trigger),
		session:         session,
		logger:          logger,
		dCADataGetter:   dCADataGetter,
		audioBufferSize: 1024 * 1024, // 1 MiB
		voiceChannelMap: make(map[string]VoiceChannelInfo),
		message:         message,
	}
}

// WithLogger establece el logger para el GuildPlayer y devuelve el mismo GuildPlayer.
func (p *GuildPlayer) WithLogger(l logging.Logger) *GuildPlayer {
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
		if currentSong, err := p.stateStorage.GetCurrentSong(); err != nil {
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
	p.mu.Lock()
	defer p.mu.Unlock()
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

// updateSongPosition actualiza la posición de la canción actual.
func (p *GuildPlayer) updateSongPosition(song *voice.Song, position time.Duration, textChannel, playMsgID string) {
	if err := p.stateStorage.SetCurrentSong(&voice.PlayedSong{Song: *song, Position: position}); err != nil {
		p.logger.Error("Error fallo al establecer la posicion actual de la cancion", zap.Error(err))
	}
	if err := p.message.EditPlayMessage(textChannel, playMsgID, &voice.PlayMessage{Song: song, Position: position}); err != nil {
		p.logger.Error("Error fallo al editar el mensaje")
	}
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
	p.logger.Info("Comenzando a escuchar eventos relevantes")
}

// AddSong agrega una o más canciones a la lista de reproducción.
func (p *GuildPlayer) AddSong(textChannelID, voiceChannelID *string, songs ...*voice.Song) error {
	for _, song := range songs {
		if err := p.songStorage.AppendSong(song); err != nil {
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

	p.logger.Info("Canciones agregadas a la lista de reproducción", zap.Int("cantidad", len(songs)))
	return nil
}

// SkipSong salta la canción actual.
func (p *GuildPlayer) SkipSong() {
	if p.songCtxCancel != nil {
		p.songCtxCancel()
		p.logger.Info("Canción actual saltada")
	}
}

// Stop detiene la reproducción y limpia la lista de reproducción.
func (p *GuildPlayer) Stop() error {
	if err := p.songStorage.ClearPlaylist(); err != nil {
		p.logger.Error("Error al limpiar la lista de reproducción", zap.Error(err))
		return fmt.Errorf("al limpiar la lista de reproducción: %w", err)
	}

	if p.songCtxCancel != nil {
		p.songCtxCancel()
		p.logger.Info("Reproducción detenida y lista de reproducción limpia")
	}

	return nil
}

// RemoveSong elimina una canción de la lista de reproducción por posición.
func (p *GuildPlayer) RemoveSong(position int) (*voice.Song, error) {
	song, err := p.songStorage.RemoveSong(position)
	if err != nil {
		p.logger.Error("Error al eliminar canción de la lista de reproducción", zap.Error(err))
		return nil, fmt.Errorf("al eliminar canción: %w", err)
	}

	p.logger.Info("Canción eliminada de la lista de reproducción", zap.String("título", song.Title))
	return song, nil
}

// GetPlaylist obtiene la lista de reproducción actual.
func (p *GuildPlayer) GetPlaylist() ([]string, error) {
	songs, err := p.songStorage.GetSongs()
	if err != nil {
		p.logger.Error("Error al obtener la lista de reproducción", zap.Error(err))
		return nil, fmt.Errorf("al obtener canciones: %w", err)
	}

	playlist := make([]string, len(songs))
	for i, song := range songs {
		playlist[i] = song.GetHumanName()
	}

	p.logger.Info("Lista de reproducción obtenida", zap.Int("cantidad", len(playlist)))
	return playlist, nil
}

// GetPlayedSong obtiene la canción que se está reproduciendo actualmente.
func (p *GuildPlayer) GetPlayedSong() (*voice.PlayedSong, error) {
	currentSong, err := p.stateStorage.GetCurrentSong()
	if err != nil {
		p.logger.Error("Error al obtener la canción que se está reproduciendo actualmente", zap.Error(err))
		return nil, err
	}
	p.logger.Info("Canción que se está reproduciendo actualmente obtenida")
	return currentSong, nil
}

// Run inicia el bucle principal del reproductor de música.
func (p *GuildPlayer) Run(ctx context.Context) error {
	currentSong, err := p.stateStorage.GetCurrentSong()
	if err != nil {
		p.logger.Info("falló al obtener la canción actual", zap.Error(err))
		return err
	} else if currentSong != nil {
		currentSong.StartPosition += currentSong.Position
		if err := p.songStorage.PrependSong(&currentSong.Song); err != nil {
			p.logger.Info("falló al agregar la canción actual en la lista de reproducción", zap.Error(err))
			return err
		}
	}

	songs, err := p.songStorage.GetSongs()
	if err != nil {
		p.logger.Error("Error al obtener canciones", zap.Error(err))
		return err
	}

	if len(songs) > 0 {
		voiceChannel, err := p.stateStorage.GetVoiceChannel()
		if err != nil {
			p.logger.Error("Error al obtener el canal de voz", zap.Error(err))
			return err
		}
		textChannel, err := p.stateStorage.GetTextChannel()
		if err != nil {
			p.logger.Error("Error al obtener el canal de texto", zap.Error(err))
			return err
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
		p.logger.Info("Esperando triggers")
		select {
		case <-ctx.Done():
			return nil
		case trigger := <-p.triggerCh:
			switch trigger.Command {
			case "play":
				if trigger.TextChannelID != nil {
					if err := p.stateStorage.SetTextChannel(*trigger.TextChannelID); err != nil {
						p.logger.Error("Error al establecer el canal de texto", zap.Error(err))
						return err
					}
				}
				if trigger.VoiceChannelID != nil {
					if err := p.stateStorage.SetVoiceChannel(*trigger.VoiceChannelID); err != nil {
						p.logger.Error("Error al establecer el canal de voz", zap.Error(err))
						return err
					}
				}

				songs, err := p.songStorage.GetSongs()
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
	p.logger.Info("playPlaylist iniciado")
	voiceChannel, err := p.stateStorage.GetVoiceChannel()
	if err != nil {
		p.logger.Error("Erorr al obtener el canal ve voz", zap.Error(err))
		return err
	}

	textChannel, err := p.stateStorage.GetTextChannel()
	if err != nil {
		p.logger.Error("Erorr al obtener el canal de texto", zap.Error(err))
		return err
	}

	p.logger.Info("uniéndose al canal de voz", zap.String("canal", voiceChannel))
	if err := p.session.JoinVoiceChannel(voiceChannel); err != nil {
		p.logger.Error("Error fallo al unirse al canal de voz", zap.Error(err))
		return err
	}

	defer func() {
		p.logger.Info("saliendo del canal de voz", zap.String("canal", voiceChannel))
		if err := p.session.LeaveVoiceChannel(); err != nil {
			p.logger.Error("Error falló al salir del canal de voz", zap.Error(err))
		}
	}()

	for {
		song, err := p.songStorage.PopFirstSong()
		if errors.Is(err, ErrNoSongs) {
			p.logger.Info("la lista de reproducción está vacía")
			break
		}
		if err != nil {
			p.logger.Error("Error al obtener la primera cancion", zap.Error(err))
			return err
		}

		if err := p.stateStorage.SetCurrentSong(&voice.PlayedSong{Song: *song}); err != nil {
			p.logger.Error("Error al establecer la cancion actual", zap.Error(err))
			return err
		}

		songCtx, cancel := context.WithCancel(ctx)
		p.mu.Lock()
		p.songCtxCancel = cancel
		p.mu.Unlock()

		p.logger.With(zap.String("título", song.Title), zap.String("URL", song.URL))

		playMsgID, err := p.message.SendPlayMessage(textChannel, &voice.PlayMessage{Song: song})
		if err != nil {
			p.logger.Error("Error al enviar el mensaje con el nombre de la cancion", zap.Error(err))
			return err
		}

		dcaData, err := p.dCADataGetter(songCtx, song)
		if err != nil {
			p.logger.Error("Error al obtener datos DCA de la cancion", zap.Any("Cancion", song), zap.Error(err))
			return err
		}

		audioReader := bufio.NewReaderSize(dcaData, p.audioBufferSize)
		p.logger.Info("enviando flujo de audio")
		if err := p.session.SendAudio(songCtx, audioReader, func(d time.Duration) {
			p.updateSongPosition(song, d, textChannel, playMsgID)
		}); err != nil {
			p.logger.Error("Error al enviar datos de audio", zap.Error(err))
			return err
		}
		p.logger.Info("Reproduccion detenida")
		p.updateSongPosition(song, song.Duration, textChannel, playMsgID)
		if err := p.stateStorage.SetCurrentSong(nil); err != nil {
			p.logger.Error("Error al establecer la cancion actual", zap.Error(err))
			return err
		}
		time.Sleep(250 * time.Millisecond)
	}
	p.logger.Info("playPlaylist finalizado")
	return nil
}
