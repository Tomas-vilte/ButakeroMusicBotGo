package main

import (
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/infrastructure/aws"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/infrastructure/repository"
	lambdaHandler "github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/interfaces/lambda"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/usecase"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/pkg/logger"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

func main() {
	logger, err := logging.NewZapLogger()
	if err != nil {
		panic("Error creando el logger: " + err.Error())
	}

	ecsClient, err := aws.NewECSClient()
	if err != nil {
		logger.Error("Error al crear el cliente de ECS:", zap.Error(err))
	}

	s3Client, err := aws.NewS3Client()
	if err != nil {
		logger.Error("Error al crear el cliente de S3:", zap.Error(err))
	}

	jobRepo, err := repository.NewDynamoDBJobRepository(logger)
	if err != nil {
		logger.Error("Error al crear el repositorio de DynamoDB:", zap.Error(err))
	}

	sendJobUseCase := usecase.NewSendJobsToECS(ecsClient, s3Client, jobRepo, logger)

	handler := lambdaHandler.NewHandler(sendJobUseCase, logger)

	lambda.Start(handler.Handle)
}
