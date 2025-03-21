package server

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/router"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/adapters"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/kafka"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/mongodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/local"
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
	if err != nil {
		log.Error("Error al crear mongo connection", zap.Error(err))
		return err
	}

	mediaRepository, err := mongodb.NewMediaRepository(mongodb.MediaRepositoryOptions{
		Log:        log,
		Collection: conn.GetCollection(cfg.Database.Mongo.Collections.Songs),
	})
	if err != nil {
		log.Error("Error al crear metadata repository", zap.Error(err))
		return err
	}

	downloaderMusic, err := downloader.NewYTDLPDownloader(log, downloader.YTDLPOptions{
		Cookies: cfg.API.YouTube.Cookies,
	})
	if err != nil {
		log.Error("Error al crear downloader", zap.Error(err))
		return err
	}
	youtubeAPI := adapters.NewYouTubeClient(cfg.API.YouTube.ApiKey, log)

	encoderAudio := encoder.NewFFmpegEncoder(log)

	providers := map[string]ports.VideoProvider{
		"youtube": youtubeAPI,
	}

	mediaService := service.NewMediaService(mediaRepository, log)
	audioStorageService := service.NewAudioStorageService(storage, log)
	topicPublisherService := service.NewMediaProcessingPublisherService(messaging, log)
	audioDownloadService := service.NewAudioDownloaderService(downloaderMusic, encoderAudio, log)
	coreService := service.NewCoreService(mediaService, audioStorageService, topicPublisherService, audioDownloadService, log, cfg)
	operationService := service.NewOperationService(mediaService, log)

	providerService := service.NewVideoService(providers, log)
	initiateDownloadUC := usecase.NewInitiateDownloadUseCase(coreService, providerService, operationService)
	operationUC := usecase.NewGetOperationStatusUseCase(mediaRepository)
	audioHandler := handler.NewAudioHandler(initiateDownloadUC)
	operationHandler := handler.NewOperationHandler(operationUC)
	healthCheck := handler.NewHealthHandler(cfg)

	gin.SetMode(cfg.GinConfig.Mode)
	r := gin.New()
	router.SetupRoutes(r, audioHandler, operationHandler, healthCheck, log)

	return r.Run(":8080")
}
