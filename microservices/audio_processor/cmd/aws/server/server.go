package server

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/queue/processor"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/router"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/adapters"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/cloud"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func StartServer() error {
	cfg := config.LoadConfigAws()

	log, err := logger.NewProductionLogger()
	if err != nil {
		return err
	}
	defer func() {
		if err := log.Close(); err != nil {
			log.Error("Error al cerrar el logger", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Info("Senal recivida cerrando", zap.String("signal", sig.String()))
		cancel()
	}()

	storage, err := cloud.NewS3Storage(cfg, log)
	if err != nil {
		log.Error("Error al crear el storage", zap.Error(err))
		return err
	}

	sqsProducer, err := sqs.NewProducerSQS(cfg, log)
	if err != nil {
		log.Error("Error al crear el producer", zap.Error(err))
		return err
	}

	sqsConsumer, err := sqs.NewConsumerSQS(cfg, log)
	if err != nil {
		log.Error("Error al crear el consumer", zap.Error(err))
		return err
	}

	mediaRepository, err := dynamodb.NewMediaRepositoryDynamoDB(cfg, log)
	if err != nil {
		log.Error("Error al crear metadata repository", zap.Error(err))
		return err
	}

	cookiesContent, err := storage.GetFileContent(context.Background(), "", "yt-cookies.txt")
	if err != nil {
		log.Error("Error al obtener contenido de archivo", zap.Error(err))
		return err
	}

	file, err := os.Create("yt-cookies.txt")
	if err != nil {
		log.Error("Error al crear archivo", zap.Error(err))
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error("Error al close archivo", zap.Error(err))
		}
	}()

	_, err = io.Copy(file, cookiesContent)
	if err != nil {
		log.Error("Error al copiar contenido a archivo", zap.Error(err))
		return err
	}
	log.Info("Archivo temporal creado", zap.String("path", file.Name()))

	downloaderMusic, err := downloader.NewYTDLPDownloader(log, downloader.YTDLPOptions{Cookies: file.Name()})
	if err != nil {
		log.Error("Error al crear downloader", zap.Error(err))
		return err
	}
	youtubeAPI := adapters.NewYouTubeClient(cfg.API.YouTube.ApiKey, log)
	providers := map[string]ports.VideoProvider{
		"youtube": youtubeAPI,
	}

	encoderAudio := encoder.NewFFmpegEncoder(log)
	mediaService := service.NewMediaService(mediaRepository, log)
	audioStorageService := service.NewAudioStorageService(storage, log)
	topicPublisherService := service.NewMediaProcessingPublisherService(sqsProducer, log)
	audioDownloadService := service.NewAudioDownloaderService(downloaderMusic, encoderAudio, log)
	coreService := service.NewCoreService(mediaService, audioStorageService, topicPublisherService, audioDownloadService, log, cfg)
	operationService := service.NewOperationService(mediaService, log)

	providerService := service.NewVideoService(providers, log)
	operationUC := usecase.NewGetOperationStatusUseCase(mediaRepository)
	operationHandler := handler.NewOperationHandler(operationUC)
	healthCheck := handler.NewHealthHandler(cfg)

	downloadService := processor.NewDownloadService(cfg.NumWorkers, sqsConsumer, mediaService, providerService, coreService, operationService, log)

	errChan := make(chan error, 1)
	go func() {
		errChan <- downloadService.Run(ctx)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			log.Error("Error al cerrar el download", zap.Error(err))
			os.Exit(1)
		}
	case <-ctx.Done():
		log.Info("señal de cancelación recibida, cerrando el servicio")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
		defer shutdownCancel()

		<-shutdownCtx.Done()
		log.Info("Servicio cerrado con éxito")
	}

	gin.SetMode(cfg.GinConfig.Mode)
	r := gin.New()
	router.SetupRoutes(r, operationHandler, healthCheck, log)

	return r.Run(":8080")
}
