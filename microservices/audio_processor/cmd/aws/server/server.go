package server

import (
	"os"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/factory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	infrastructure "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/factory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/router"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func StartServer() error {
	cfg := config.LoadConfig(os.Getenv("ENVIRONMENT"))

	var envFactory factory.EnvironmentFactory
	if cfg.Environment == "prod" {
		envFactory = &infrastructure.AWSFactory{}
	}

	log, err := logger.NewZapLogger()
	if err != nil {
		return err
	}
	defer log.Close()

	log.Info("Corriendo en un entorono", zap.String("ENV", cfg.Environment))

	storageService, err := envFactory.CreateStorage(cfg)
	if err != nil {
		log.Error("Error al crear el storage", zap.Error(err))
		return err
	}

	messaging, err := envFactory.CreateQueue(cfg, log)
	if err != nil {
		log.Error("Error al crear queue", zap.Error(err))
		return err
	}

	metadataRepo, err := envFactory.CreateMetadataRepository(cfg, log)
	if err != nil {
		log.Error("Error al crear metadata repository", zap.Error(err))
		return err
	}

	operationRepo, err := envFactory.CreateOperationRepository(cfg, log)
	if err != nil {
		log.Error("Error al crear operation repository", zap.Error(err))
		return err
	}

	tempFile, err := os.Create("cookies.txt")
	if err != nil {
		log.Error("Error al crear archivo temporal", zap.Error(err))
		return err
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(cfg.API.YouTube.Cookies); err != nil {
		log.Error("Error al escribir en archivo temporal", zap.Error(err))
		tempFile.Close()
		return err
	}

	if err := tempFile.Close(); err != nil {
		log.Error("Error al cerrar archivo temporal", zap.Error(err))
		return err
	}
	log.Info("Archivo temporal creado", zap.String("path", tempFile.Name()))

	downloaderMusic := downloader.NewYTDLPDownloader(log, downloader.YTDLPOptions{
		UseOAuth2: cfg.API.OAuth2.ParseBool(),
		Cookies:   tempFile.Name(),
	})
	youtubeAPI := api.NewYouTubeClient(cfg.API.YouTube.ApiKey)

	audioProcessingService := service.NewAudioProcessingService(
		log,
		storageService,
		downloaderMusic,
		operationRepo,
		metadataRepo,
		messaging,
		cfg,
	)

	getOperationStatus := usecase.NewGetOperationStatusUseCase(operationRepo)
	initiateDownloadUC := usecase.NewInitiateDownloadUseCase(audioProcessingService, youtubeAPI)
	audioHandler := handler.NewAudioHandler(initiateDownloadUC, getOperationStatus)
	healthCheck := handler.NewHealthHandler(cfg)

	gin.SetMode(cfg.GinConfig.Mode)
	r := gin.New()
	router.SetupRoutes(r, audioHandler, healthCheck, log)

	return r.Run(":8080")
}