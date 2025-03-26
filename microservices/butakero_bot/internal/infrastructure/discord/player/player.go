package player

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"sync"
	"time"
)

var _ ports.GuildPlayer = (*GuildPlayer)(nil)

// Trigger representa un disparador para comandos relacionados con la reproducción de música.
type Trigger struct {
	Command        string
	VoiceChannelID *string
	TextChannelID  *string
}

// GuildPlayer es el reproductor de música para un servidor específico en Discord.
type GuildPlayer struct {
	triggerCh     chan Trigger
	session       ports.VoiceSession
	songCtxCancel context.CancelFunc
	songStorage   ports.SongStorage
	stateStorage  ports.StateStorage
	logger        logging.Logger
	message       ports.DiscordMessenger
	storageAudio  ports.StorageAudio
	mu            sync.Mutex
}

// NewGuildPlayer crea una nueva instancia de GuildPlayer con los parámetros proporcionados.
func NewGuildPlayer(session ports.VoiceSession, songStorage ports.SongStorage, stateStorage ports.StateStorage,
	message ports.DiscordMessenger, storageAudio ports.StorageAudio, logger logging.Logger) *GuildPlayer {
	return &GuildPlayer{
		songStorage:  songStorage,
		stateStorage: stateStorage,
		triggerCh:    make(chan Trigger),
		session:      session,
		logger:       logger,
		message:      message,
		storageAudio: storageAudio,
	}
}

func (p *GuildPlayer) getVoiceAndTextChannels() (voiceChannel string, textChannel string, err error) {
	voiceChannel, err = p.stateStorage.GetVoiceChannel()
	if err != nil {
		return "", "", fmt.Errorf("error al obtener el canal de voz: %w", err)
	}
	textChannel, err = p.stateStorage.GetTextChannel()
	if err != nil {
		return "", "", fmt.Errorf("error al obtener el canal de texto: %w", err)
	}
	return voiceChannel, textChannel, nil
}

// AddSong agrega una o más canciones a la lista de reproducción.
func (p *GuildPlayer) AddSong(textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error {
	if err := p.songStorage.AppendSong(playedSong); err != nil {
		p.logger.Error("Error al agregar canción a la lista de reproducción", zap.Error(err))
		return fmt.Errorf("al agregar canción: %w", err)
	}

	go func() {
		p.triggerCh <- Trigger{
			Command:        "play",
			VoiceChannelID: voiceChannelID,
			TextChannelID:  textChannelID,
		}
	}()

	p.logger.Info("Canción agregada a la lista de reproducción", zap.String("título", playedSong.DiscordSong.TitleTrack))
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
func (p *GuildPlayer) RemoveSong(position int) (*entity.DiscordEntity, error) {
	song, err := p.songStorage.RemoveSong(position)
	if err != nil {
		p.logger.Error("Error al eliminar canción de la lista de reproducción", zap.Error(err))
		return nil, fmt.Errorf("al eliminar canción: %w", err)
	}

	p.logger.Info("Canción eliminada de la lista de reproducción", zap.String("título", song.DiscordSong.TitleTrack))
	return song.DiscordSong, nil
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
		playlist[i] = song.DiscordSong.TitleTrack
	}

	p.logger.Info("Lista de reproducción obtenida", zap.Int("cantidad", len(playlist)))
	return playlist, nil
}

// GetPlayedSong obtiene la canción que se está reproduciendo actualmente.
func (p *GuildPlayer) GetPlayedSong() (*entity.PlayedSong, error) {
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
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Manejo de la canción actual al iniciar
	currentSong, err := p.stateStorage.GetCurrentSong()
	if err != nil {
		p.logger.Info("falló al obtener la canción actual", zap.Error(err))
		return err
	}
	if currentSong != nil {
		currentSong.StartPosition += currentSong.Position
		if err := p.songStorage.PrependSong(currentSong); err != nil {
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
		voiceChannel, textChannel, err := p.getVoiceAndTextChannels()
		if err != nil {
			p.logger.Error("Error al obtener canales", zap.Error(err))
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
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
			p.logger.Info("Contexto cancelado, saliendo del bucle principal")
			wg.Wait()
			return nil
		case trigger := <-p.triggerCh:
			if err := p.handleTrigger(ctx, trigger); err != nil {
				p.logger.Error("Error al manejar trigger", zap.Error(err))
			}
		}
	}
}

// playPlaylist reproduce la lista de reproducción de canciones.
func (p *GuildPlayer) playPlaylist(ctx context.Context) error {
	p.logger.Info("playPlaylist iniciado")
	voiceChannel, textChannel, err := p.getVoiceAndTextChannels()
	if err != nil {
		p.logger.Error("Error al obtener canales", zap.Error(err))
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
		if errors.Is(err, inmemory.ErrNoSongs) {
			p.logger.Info("la lista de reproducción está vacía")
			break
		}
		if err != nil {
			p.logger.Error("Error al obtener la primera cancion", zap.Error(err))
			return err
		}

		if err := p.playSingleSong(ctx, song, textChannel); err != nil {
			p.logger.Error("Error reproduciendo canción", zap.Error(err))
			continue
		}

		time.Sleep(250 * time.Millisecond)
	}

	p.logger.Info("playPlaylist finalizado")
	return nil
}

// playSingleSong reproduce una única canción
func (p *GuildPlayer) playSingleSong(ctx context.Context, song *entity.PlayedSong, textChannel string) error {
	if err := p.stateStorage.SetCurrentSong(song); err != nil {
		p.logger.Error("Error al establecer la canción actual", zap.Error(err))
		return err
	}

	songCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	p.mu.Lock()
	p.songCtxCancel = cancel
	p.mu.Unlock()

	playMsgID, err := p.message.SendPlayStatus(textChannel, song)
	if err != nil {
		p.logger.Error("Error al enviar el mensaje con el nombre de la canción", zap.Error(err))
		return err
	}

	audioData, err := p.storageAudio.GetAudio(songCtx, song.DiscordSong.FilePath)
	if err != nil {
		p.logger.Error("Error al obtener datos de audio", zap.Any("Cancion", song), zap.Error(err))
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	startTime := time.Now()

	go func() {
		for {
			select {
			case <-ticker.C:
				elapsedMs := time.Since(startTime).Milliseconds()
				song.Position = elapsedMs

				if err := p.message.UpdatePlayStatus(textChannel, playMsgID, song); err != nil {
					p.logger.Error("Error al actualizar el estado de reproducción", zap.Error(err))
				}
			case <-done:
				return
			}
		}
	}()

	if err := p.session.SendAudio(songCtx, audioData); err != nil {
		p.logger.Error("Error al enviar datos de audio", zap.Error(err))
		return err
	}

	close(done)

	return nil
}

func (p *GuildPlayer) handleTrigger(ctx context.Context, trigger Trigger) error {
	switch trigger.Command {
	case "play":
		if trigger.TextChannelID != nil {
			if err := p.stateStorage.SetTextChannel(*trigger.TextChannelID); err != nil {
				return fmt.Errorf("error al establecer el canal de texto: %w", err)
			}
		}
		if trigger.VoiceChannelID != nil {
			if err := p.stateStorage.SetVoiceChannel(*trigger.VoiceChannelID); err != nil {
				return fmt.Errorf("error al establecer el canal de voz: %w", err)
			}
		}
		songs, err := p.songStorage.GetSongs()
		if err != nil {
			return fmt.Errorf("error al obtener canciones: %w", err)
		}

		if len(songs) == 0 {
			return nil
		}

		if err := p.playPlaylist(ctx); err != nil {
			return fmt.Errorf("error al reproducir la lista de reproducción: %w", err)
		}
	}
	return nil
}
