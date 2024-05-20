package main

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/github_event"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/message_queue"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/service"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
	"os"
)

func main() {
	// Configuración del logger
	logger, err := logging.NewZapLogger()
	if err != nil {
		panic("Error creando el logger: " + err.Error())
	}
	defer func() {
		err := logger.Close()
		if err != nil {
			logger.Error("Error cerrando el logger", zap.Error(err))
		}
	}()

	// Crear el cliente SQS
	sqsClient := message_queue.NewSQSClient()

	queueURL := os.Getenv("QUEUE_URL_SQS")

	// Crear el publicador SQS
	sqsPublisher := message_queue.NewSQSPublisher(sqsClient, queueURL, logger)

	// Crear el procesador de eventos
	eventProcessor := service.NewEventProcessor(sqsPublisher, logger)

	// Crear el manejador de eventos de GitHub
	eventHandler := github_event.NewEventHandler(eventProcessor, logger)

	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return eventHandler.HandleGitHubEvent(ctx, request)
	})
}