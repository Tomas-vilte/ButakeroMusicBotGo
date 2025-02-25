package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"io"
	"sync"
)

type MusicPlayer struct {
	discordMessenger    ports.DiscordMessenger
	messageConsumer     ports.MessageConsumer
	externalSongService ports.ExternalSongService
	songRepo            ports.SongRepository
	songStorage         ports.SongStorage
	stateStorage        ports.StateStorage
	storageAudio        ports.StorageAudio
	voiceSession        ports.VoiceSession
	decoderFactory      func(io.ReadCloser) ports.Decoder

	// Estado específico del servidor
	playlist    []*entity.Song
	currentSong *entity.PlayedSong
	voiceChanID string
	textChanID  string
	isPlaying   bool
	cancelFunc  context.CancelFunc
	mu          sync.Mutex
}

func NewMusicPlayer(
	discordMessenger ports.DiscordMessenger,
	messageConsumer ports.MessageConsumer,
	externalSongService ports.ExternalSongService,
	songRepo ports.SongRepository,
	songStorage ports.SongStorage,
	stateStorage ports.StateStorage,
	storageAudio ports.StorageAudio,
	voiceSession ports.VoiceSession,
	decoderFactory func(io.ReadCloser) ports.Decoder,
) *MusicPlayer {
	return &MusicPlayer{
		discordMessenger:    discordMessenger,
		messageConsumer:     messageConsumer,
		externalSongService: externalSongService,
		songRepo:            songRepo,
		songStorage:         songStorage,
		stateStorage:        stateStorage,
		storageAudio:        storageAudio,
		voiceSession:        voiceSession,
		decoderFactory:      decoderFactory,
		playlist:            []*entity.Song{},
	}
}

// Play inicia la reproducción de una canción
func (m *MusicPlayer) Play(ctx context.Context, channelID, input string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Guardar el canal de voz y texto
	m.voiceChanID = channelID
	m.textChanID = channelID

	// Buscar la canción en la base de datos
	song, err := m.songRepo.SearchSongsByTitle(ctx, input)
	if err != nil && !errors.Is(err, errors.New("no canciones")) {
		return err
	}

	// Si no existe, solicitar descarga al microservicio externo
	if song == nil {
		resp, err := m.externalSongService.RequestDownload(ctx, input)
		if err != nil {
			return err
		}

		// Escuchar eventos de la cola de mensajes
		go m.waitForDownloadCompletion(ctx, resp.OperationID)
		return nil
	}

	// Agregar la canción a la lista de reproducción
	m.playlist = append(m.playlist, song[0])

	// Si no hay una canción reproduciéndose, iniciar la reproducción
	if !m.isPlaying {
		go m.startPlayback(ctx)
	}

	return nil
}

// waitForDownloadCompletion espera a que la descarga se complete
func (m *MusicPlayer) waitForDownloadCompletion(ctx context.Context, operationID string) {
	for {
		select {
		case msg := <-m.messageConsumer.GetMessagesChannel():
			if msg.Status.ID == operationID {
				if msg.Status.Status == "success" {
					// Obtener la canción desde la base de datos
					song, err := m.songRepo.GetSongByVideoID(ctx, msg.Status.Metadata.VideoID)
					if err != nil {
						// Manejar error
						return
					}

					// Agregar la canción a la lista de reproducción
					m.mu.Lock()
					m.playlist = append(m.playlist, song)
					m.mu.Unlock()

					// Iniciar reproducción si no hay una canción reproduciéndose
					if !m.isPlaying {
						go m.startPlayback(ctx)
					}
				} else {
					// Manejar error de descarga
				}
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// startPlayback inicia la reproducción de la lista de reproducción
func (m *MusicPlayer) startPlayback(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for len(m.playlist) > 0 {
		m.isPlaying = true

		// Obtener la primera canción de la lista
		song := m.playlist[0]
		m.playlist = m.playlist[1:]

		// Reproducir la canción
		err := m.playSong(ctx, song)
		if err != nil {
			// Manejar error
		}
	}

	m.isPlaying = false
}

// playSong reproduce una canción
func (m *MusicPlayer) playSong(ctx context.Context, song *entity.Song) error {
	// Unirse al canal de voz
	err := m.voiceSession.JoinVoiceChannel("1231503103745069077")
	if err != nil {
		return err
	}

	// Obtener el archivo de audio
	reader, err := m.storageAudio.GetAudio(ctx, "audio/Twenty One Pilots - “The Line” (from Arcane Season 2) [Official Music Video].dca")
	if err != nil {
		return err
	}
	defer reader.Close()

	// Crear un nuevo decodificador
	decoder := m.decoderFactory(reader)
	defer decoder.Close()

	// Decodificar y enviar el audio
	for {
		frame, err := decoder.OpusFrame()
		if err != nil {
			if err != io.EOF {
				fmt.Println("Archivo DCA terminado.")
			}
			fmt.Printf("Error al leer el frame de Opus: %v\n", err)
			break
		}

		if err := m.voiceSession.SendAudio(ctx, frame); err != nil {
			return err
		}
	}

	return nil
}

// Stop detiene la reproducción actual
func (m *MusicPlayer) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isPlaying {
		return errors.New("no hay reproducción en curso")
	}

	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	m.isPlaying = false
	return nil
}

// Skip salta a la siguiente canción
func (m *MusicPlayer) Skip() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isPlaying {
		return errors.New("no hay reproducción en curso")
	}

	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	return nil
}

// Close libera los recursos del reproductor
func (m *MusicPlayer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.voiceSession != nil {
		return m.voiceSession.Close()
	}
	return nil
}
