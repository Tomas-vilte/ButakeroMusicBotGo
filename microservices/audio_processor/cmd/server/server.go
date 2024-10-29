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
		QueueURL:              os.Getenv("SQS_QUEUE_URL"),
		Topic:                 os.Getenv("TOPIC"),
		Brokers:               []string{os.Getenv("KAFKA_BROKERS")},
		Environment:           getGinMode(),
		OAuth2:                os.Getenv("OAUTH2"),
		Mongo: config.MongoConfig{
			User:                       os.Getenv("MONGO_USER"),
			Password:                   os.Getenv("MONGO_PASSWORD"),
			Port:                       os.Getenv("MONGO_PORT"),
			Host:                       os.Getenv("MONGO_HOST"),
			Database:                   os.Getenv("MONGO_DATABASE"),
			SongsCollection:            os.Getenv("MONGO_SONGS_COLLECTION"),
			OperationResultsCollection: os.Getenv("MONGO_OPERATION_RESULTS_COLLECTION"),
		},
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

	downloaderMusic := downloader.NewYTDLPDownloader(log, downloader.YTDLPOptions{UseOAuth2: cfg.ParseBool()})
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

	youtubeAPI := api.NewYouTubeClient(cfg.YouTubeApiKey)
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

func getGinMode() string {
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		return mode
	}
	return "debug"
}
