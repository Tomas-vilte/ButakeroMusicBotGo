package server

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/router"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/adapters"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/cloud"
	"io"
	"os"

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

	storage, err := cloud.NewS3Storage(cfg, log)
	if err != nil {
		log.Error("Error al crear el storage", zap.Error(err))
		return err
	}

	messaging, err := sqs.NewSQSService(cfg, log)
	if err != nil {
		log.Error("Error al crear queue", zap.Error(err))
		return err
	}

	metadataRepo, err := dynamodb.NewMetadataStore(cfg, log)
	if err != nil {
		log.Error("Error al crear metadata repository", zap.Error(err))
		return err
	}

	operationRepo, err := dynamodb.NewOperationStore(cfg, log)
	if err != nil {
		log.Error("Error al crear operation repository", zap.Error(err))
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

	downloaderMusic, err := downloader.NewYTDLPDownloader(log, downloader.YTDLPOptions{
		UseOAuth2: cfg.API.OAuth2.ParseBool(),
		Cookies:   file.Name(),
	})
	if err != nil {
		log.Error("Error al crear downloader", zap.Error(err))
		return err
	}
	youtubeAPI := adapters.NewYouTubeClient(cfg.API.YouTube.ApiKey, log)

	encoderAudio := encoder.NewFFmpegEncoder(log)
	downloaderService := service.NewAudioDownloader(downloaderMusic, encoderAudio, log)
	storageService := service.NewAudioStorage(storage, metadataRepo, log)
	opsManager := service.NewOperationManager(operationRepo, log, cfg)
	messageService := service.NewMessagingService(messaging, log)
	errorHandler := service.NewErrorHandler(operationRepo, messaging, log, cfg)

	audioProcessingService := service.NewAudioProcessingService(downloaderService, storageService, opsManager, messageService, errorHandler, log, cfg)

	operationUC := usecase.NewGetOperationStatusUseCase(operationRepo)

	operationService := service.NewOperationService(operationRepo, log)

	providers := map[string]ports.VideoProvider{
		"youtube": youtubeAPI,
	}

	providerService := service.NewVideoService(providers, log)
	initiateDownloadUC := usecase.NewInitiateDownloadUseCase(audioProcessingService, providerService, operationService)
	audioHandler := handler.NewAudioHandler(initiateDownloadUC)
	operationHandler := handler.NewOperationHandler(operationUC)
	healthCheck := handler.NewHealthHandler(cfg)

	gin.SetMode(cfg.GinConfig.Mode)
	r := gin.New()
	router.SetupRoutes(r, audioHandler, operationHandler, healthCheck, log)

	return r.Run(":8080")
}
