package queuing

import (
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/message_processing/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/message_processing/internal/messaging"
	"go.uber.org/zap"
)

// EventProcessor define la interfaz para procesar eventos provenientes de una cola.
type EventProcessor interface {
	ProcessSQSEvent(body []byte) error // ProcessSQSEvent procesa un evento proveniente de una cola SQS.
}

// SQSConsumer es una implementación de EventProcessor que consume eventos desde una cola SQS.
type SQSConsumer struct {
	discordClient messaging.DiscordMessenger // Cliente Discord para enviar mensajes.
	logger        logging.Logger             // Logger para registrar eventos.
}

// NewSQSConsumer crea una nueva instancia de SQSConsumer.
func NewSQSConsumer(discordClient messaging.DiscordMessenger, logger logging.Logger) *SQSConsumer {
	return &SQSConsumer{
		discordClient: discordClient,
		logger:        logger,
	}
}

// ProcessSQSEvent procesa un evento proveniente de una cola SQS.
func (s *SQSConsumer) ProcessSQSEvent(body []byte) error {
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		s.logger.Error("Error al analizar el cuerpo del mensaje", zap.Error(err))
		return errors.New("error al analizar el cuerpo del mensaje")
	}

	action, ok := event["action"].(string)
	if !ok {
		s.logger.Error("Error el campo 'action', no encontrado o no es una cadena")
		return nil
	}
	var formatter EventFormatter
	switch action {
	case "published":
		formatter = &ReleaseEventFormatter{}
	//case "completed":
	//	formatter = &WorkflowActionEventFormatter{}
	default:
		s.logger.Error("Error acción desconocida", zap.String("action", action))
		return errors.New("acción desconocida: " + action)
	}

	embed, err := formatter.FormatEvent(event)
	if err != nil {
		s.logger.Error("Error al formatear el evento", zap.Error(err))
		return err
	}

	if err := s.discordClient.SendMessageToServers(embed); err != nil {
		s.logger.Error("Error al enviar el mensaje a Discord", zap.Error(err))
		return err
	}
	return nil
}
