package server

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/adapters"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/kafka"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/mongodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/local"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/router"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func StartServer() error {
	cfg := config.LoadConfigLocal()

	log, err := logger.NewDevelopmentLogger()
	if err != nil {
		return err
	}
	defer func() {
		if err := log.Close(); err != nil {
			log.Error("Error al cerrar el logger", zap.Error(err))
		}
	}()

	storage, err := local.NewLocalStorage(cfg, log)
	if err != nil {
		log.Error("Error al crear el storage", zap.Error(err))
		return err
	}

	messaging, err := kafka.NewKafkaService(cfg, log)
	if err != nil {
		log.Error("Error al crear queue", zap.Error(err))
		return err
	}

	defer func() {
		if err := messaging.Close(); err != nil {
			log.Error("Error al cerrar el queue", zap.Error(err))
		}
	}()

	conn, err := mongodb.NewMongoDB(mongodb.MongoOptions{
		Log:    log,
		Config: cfg,
	})

	metadataRepo, err := mongodb.NewMongoMetadataRepository(mongodb.MongoMetadataOptions{
		Log:        log,
		Collection: conn.GetCollection(cfg.Database.Mongo.Collections.Songs),
	})
	if err != nil {
		log.Error("Error al crear metadata repository", zap.Error(err))
		return err
	}

	operationRepo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
		Log:        log,
		Collection: conn.GetCollection(cfg.Database.Mongo.Collections.Operations),
	})
	if err != nil {
		log.Error("Error al crear operation repository", zap.Error(err))
		return err
	}

	downloaderMusic, err := downloader.NewYTDLPDownloader(log, downloader.YTDLPOptions{
		UseOAuth2: cfg.API.OAuth2.ParseBool(),
		Cookies:   cfg.API.YouTube.Cookies,
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
