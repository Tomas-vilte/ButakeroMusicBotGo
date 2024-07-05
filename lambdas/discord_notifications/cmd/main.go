package main

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/message_processing/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/message_processing/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/message_processing/internal/messaging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/message_processing/internal/queuing"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

// handler es la función que maneja los eventos de SQS.
// Recibe un events.SQSEvent y devuelve un error si ocurre.
func handler(sqsEvent events.SQSEvent) error {
	// Crear un nuevo logger usando la librería zap.
	logger, err := logging.NewZapLogger()
	if err != nil {
		panic("Error creando el logger: " + err.Error())
	}
	// Cargar la configuración del entorno.
	configEnv := config.LoadConfig()
	defer func() {
		// Cerrar el logger cuando la función termine.
		err := logger.Close()
		if err != nil {
			logger.Error("Error cerrando el logger", zap.Error(err))
		}
	}()

	// Crear una nueva sesión de Discord.
	discordSession, err := messaging.NewDiscordSessionImpl(configEnv.DiscordToken)
	if err != nil {
		logger.Error("Error en creando session con discord", zap.Error(err))
		return err
	}

	// Crear un cliente DiscordGo utilizando la sesión de Discord.
	discordClient := messaging.NewDiscordGoClient(discordSession, logger)

	// Crear un consumidor SQS para procesar los mensajes de la cola.
	sqsConsumer := queuing.NewSQSConsumer(discordClient, logger)

	// Iterar sobre cada mensaje en el evento SQS.
	for _, message := range sqsEvent.Records {
		// Procesar el mensaje utilizando el consumidor SQS.
		if err := sqsConsumer.ProcessSQSEvent([]byte(message.Body)); err != nil {
			return fmt.Errorf("error al procesar el mensaje de la cola SQS: %v", err)
		}
	}
	return nil
}

func main() {
	// Iniciar la función de lambda pasando la función processor como argumento.
	lambda.Start(handler)
}
