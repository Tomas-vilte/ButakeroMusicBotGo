package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type (
	// MessageConsumer define una interfaz para consumir mensajes de una cola.
	// Cuando se inicia la descarga de la canción desde el SongDownloader,
	// el microservicio de descarga envía el resultado a una cola. Los estados posibles
	// son "success" y "error".
	MessageConsumer interface {
		ConsumeMessages(ctx context.Context, offset int64) error
		GetMessagesChannel() <-chan *entity.MessageQueue
		Close() error
	}

	// MessageProducer define una interfaz para producir mensajes en una cola.
	// Cuando se solicita una canción y no existe en la base de datos,
	// el bot envía un mensaje a la cola para iniciar el proceso de descarga.
	// El mensaje incluye la información necesaria como el ID de interacción,
	// el ID del usuario y los detalles de la canción solicitada.
	MessageProducer interface {
		PublishSongRequest(ctx context.Context, message *entity.SongRequestMessage) error
		Close() error
	}
)
