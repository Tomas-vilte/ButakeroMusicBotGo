package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// MessageConsumer define una interfaz para consumir mensajes de una cola.
// Cuando se inicia la descarga de la canción desde el SongDownloader,
// el microservicio de descarga envía el resultado a una cola. Los estados posibles
// son "success" y "error".
type MessageConsumer interface {
	ConsumeMessages(ctx context.Context, offset int64) error
	GetMessagesChannel() <-chan *entity.StatusMessage
}
