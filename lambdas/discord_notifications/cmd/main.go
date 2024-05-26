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

func handler(sqsEvent events.SQSEvent) error {
	logger, err := logging.NewZapLogger()
	if err != nil {
		panic("Error creando el logger: " + err.Error())
	}
	configEnv := config.LoadConfig()
	defer func() {
		err := logger.Close()
		if err != nil {
			logger.Error("Error cerrando el logger", zap.Error(err))
		}
	}()

	discordSession, err := messaging.NewDiscordSessionImpl(configEnv.DiscordToken)
	if err != nil {
		logger.Error("Error en creando session con discord", zap.Error(err))
		return err
	}

	discordClient := messaging.NewDiscordGoClient(discordSession, logger)

	sqsConsumer := queuing.NewSQSConsumer(discordClient, logger)

	for _, message := range sqsEvent.Records {
		if err := sqsConsumer.ProcessSQSEvent([]byte(message.Body)); err != nil {
			return fmt.Errorf("error al procesar el mensaje de la cola SQS: %v", err)
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
