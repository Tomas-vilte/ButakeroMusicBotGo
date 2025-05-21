package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/queue"
)

type (
	// SongDownloadEventSubscriber define una interfaz para consumir mensajes de una cola.
	// Cuando se inicia la descarga de la canción desde el SongDownloader,
	// el microservicio de descarga envía el resultado a una cola. Los estados posibles
	// son "success" y "error".
	SongDownloadEventSubscriber interface {
		// SubscribeToDownloadEvents inicia la suscripción a los eventos de descarga.
		SubscribeToDownloadEvents(ctx context.Context) error
		// DownloadEventsChannel devuelve un canal para recibir mensajes de estado de descarga.
		DownloadEventsChannel() <-chan *queue.DownloadStatusMessage
		// CloseSubscription cierra la suscripción a los eventos de descarga.
		CloseSubscription() error
	}

	// SongDownloadRequestPublisher define una interfaz para producir mensajes en una cola.
	// Cuando se solicita una canción y no existe en la base de datos,
	// el bot envía un mensaje a la cola para iniciar el proceso de descarga.
	// El mensaje incluye la información necesaria como el ID de interacción,
	// el ID del usuario y los detalles de la canción solicitada.
	SongDownloadRequestPublisher interface {
		// PublishDownloadRequest envía un mensaje de solicitud de descarga a la cola.
		PublishDownloadRequest(ctx context.Context, request *queue.DownloadRequestMessage) error
		// ClosePublisher cierra la conexión del publicador.
		ClosePublisher() error
	}
)
