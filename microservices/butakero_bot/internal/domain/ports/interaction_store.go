package ports

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// InteractionStorage define la interfaz para el almacenamiento de interacciones.
type InteractionStorage interface {
	SaveSongList(channelID string, list []*entity.Song)
	GetSongList(channelID string) []*entity.Song
	DeleteSongList(channelID string)
}
