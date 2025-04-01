package player

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
)

// PlayerState representa el estado actual del reproductor
type PlayerState string

const (
	StateIdle    PlayerState = "idle"
	StatePlaying PlayerState = "playing"
	StatePaused  PlayerState = "paused"
)

// Config contiene todas las dependencias necesarias para el reproductor
type Config struct {
	VoiceSession ports.VoiceSession
	SongStorage  ports.SongStorage
	StateStorage ports.StateStorage
	Messenger    ports.DiscordMessenger
	StorageAudio ports.StorageAudio
	Logger       logging.Logger
}
