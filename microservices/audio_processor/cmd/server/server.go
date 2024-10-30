package server

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/cloud"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/router"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
)

func StartServer() error {
	cfg, err := config.LoadConfig("./configurations/config.yaml")
	if err != nil {
		return err
	}

	log, err := logger.NewZapLogger()
	if err != nil {
		return err
	}
	defer log.Close()

	storageService, err := cloud.NewS3Storage(cfg)
	if err != nil {
		return err
	}

	downloaderMusic := downloader.NewYTDLPDownloader(log, downloader.YTDLPOptions{UseOAuth2: cfg.GinConfig.ParseBool()})
	operationRepo, err := dynamodb.NewOperationStore(cfg)
	if err != nil {
		return err
	}

	metadataRepo, err := dynamodb.NewMetadataStore(cfg)
	if err != nil {
		return err
	}

	//options := mongodb.MongoOptions{
	//	Config: cfg,
	//	Log:    log,
	//}
	//
	//client, err := mongodb.NewMongoDB(options)
	//if err != nil {
	//	return err
	//}
	//
	//operationRepo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
	//	Collection: client.GetCollection(cfg.Mongo.OperationResultsCollection),
	//	Log:        log,
	//})
	//
	//metadataRepo, err := mongodb.NewMongoMetadataRepository(mongodb.MongoMetadataOptions{
	//	Collection: client.GetCollection(cfg.Mongo.SongsCollection),
	//	Log:        log,
	//})
	//defer client.Close(context.Background())

	messaging, err := sqs.NewSQSService(cfg, log)
	if err != nil {
		return err
	}

	//messaging, err := kafka.NewKafkaService(cfg, log)
	//if err != nil {
	//	return err
	//}

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
	gin.SetMode(cfg.Environment)
	r := gin.New()
	router.SetupRoutes(r, audioHandler, healthCheck, log)
	return r.Run(":8080")
}
