package server

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/persistence/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/router"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
	"os"
	"time"
)

func StartServer() error {
	cfg := config.Config{
		MaxAttempts:           3,
		Timeout:               4 * time.Minute,
		BucketName:            os.Getenv("BUCKET_NAME"),
		Region:                os.Getenv("REGION"),
		YouTubeApiKey:         os.Getenv("YOUTUBE_API_KEY"),
		SongsTable:            os.Getenv("DYNAMODB_TABLE_NAME_SONGS"),
		OperationResultsTable: os.Getenv("DYNAMODB_TABLE_NAME_OPERATION"),
		AccessKey:             os.Getenv("ACCESS_KEY"),
		SecretKey:             os.Getenv("SECRET_KEY"),
	}

	log, err := logger.NewZapLogger()
	if err != nil {
		return err
	}
	defer log.Close()

	storageService, err := storage.NewS3Storage(cfg.BucketName, cfg.Region)
	if err != nil {
		return err
	}

	downloaderMusic := downloader.NewYTDLPDownloader(log, downloader.YTDLPOptions{UseOAuth2: false})
	operationRepo, err := dynamodb.NewOperationStore(cfg.OperationResultsTable, cfg.Region)
	if err != nil {
		return err
	}

	metadataRepo, err := dynamodb.NewMetadataStore(cfg.SongsTable, cfg.Region)
	if err != nil {
		return err
	}

	youtubeAPI := api.NewYouTubeClient(cfg.YouTubeApiKey)
	audioProcessingService := service.NewAudioProcessingService(
		log,
		storageService,
		downloaderMusic,
		operationRepo,
		metadataRepo,
		cfg,
	)
	getOperationStatus := usecase.NewGetOperationStatusUseCase(operationRepo)
	initiateDownloadUC := usecase.NewInitiateDownloadUseCase(audioProcessingService, youtubeAPI)
	audioHandler := handler.NewAudioHandler(initiateDownloadUC, getOperationStatus)
	healthCheck := handler.NewHealthHandler(cfg)
	r := gin.Default()
	router.SetupRoutes(r, audioHandler, healthCheck)

	return r.Run(":8080")
}
