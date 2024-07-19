package main

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/notification"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/processor"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/uploader"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
	"os"
)

func main() {
	cfg := &config.Config{
		BucketName:      os.Getenv("BUCKET_NAME"),
		AccessKey:       os.Getenv("ACCESS_KEY"),
		SecretKey:       os.Getenv("SECRET_KEY"),
		Region:          os.Getenv("REGION"),
		Key:             os.Getenv("KEY"),
		InputFileFromS3: os.Getenv("INPUT_FILE_FROM_S3"),
		SQSQueueURL:     os.Getenv("SQS_QUEUE_URL"),
	}

	logger, err := logging.NewZapLogger(false)
	if err != nil {
		panic("Error creando el logger: " + err.Error())
	}

	// Crear session ocn aws
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
	})
	if err != nil {
		logger.Error("Error al crear la session de aws", zap.Error(err))
		return
	}
	sqsClient := sqs.New(sess)
	notifier := notification.NewSQSNotifier(sqsClient, cfg.SQSQueueURL, logger)
	s3Client := s3.New(sess)
	s3Uploader := s3manager.NewUploader(sess)

	uploaderS3 := uploader.NewS3Uploader(s3Client, s3Uploader, logger, *cfg)
	commandExecutor := processor.NewCommandExecutor()
	audioProcessor := processor.NewAudioProcessor(logger, commandExecutor, uploaderS3, "dca", "ffmpeg")

	err = audioProcessor.ProcessToDCA(context.Background(), cfg.Key, cfg.InputFileFromS3)
	if err != nil {
		logger.Error("Error al procesar archivo de audio", zap.Error(err))
		notifier.NotifyProcessingResult(cfg.Key, cfg.InputFileFromS3, "", false, err.Error())
	} else {
		processedFile := cfg.Key
		logger.Info("Procesamiento de archivo de audio completado exitosamente", zap.String("key", cfg.Key))
		notifier.NotifyProcessingResult(cfg.Key, cfg.InputFileFromS3, processedFile, true, "")
	}
}
