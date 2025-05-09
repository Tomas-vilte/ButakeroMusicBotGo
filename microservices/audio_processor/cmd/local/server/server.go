package server

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/http/controller"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/queue/processor"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	storage, err := local.NewLocalStorage(cfg, log)
	if err != nil {
		log.Error("Error al crear el storage", zap.Error(err))
		return err
	}

	kafkaProducer, err := kafka.NewProducerKafka(cfg, log)
	if err != nil {
		log.Error("Error al crear el producer", zap.Error(err))
		return err
	}

	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Error("Error al cerrar el producidor", zap.Error(err))
		}
	}()

	kafkaConsumer, err := kafka.NewConsumerKafka(cfg, log)
	if err != nil {
		log.Error("Error al crear el consumer", zap.Error(err))
		return err
	}
	defer func() {
		if err := kafkaConsumer.Close(); err != nil {
			log.Error("Error al cerrar el consumidor", zap.Error(err))
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

	audioStorageService := service.NewAudioStorageService(storage, log)
	audioDownloadService := service.NewAudioDownloaderService(downloaderMusic, encoderAudio, log, model.StdEncodeOptions)
	coreService := service.NewCoreService(mediaRepository, audioStorageService, kafkaProducer, audioDownloadService, log, cfg)
	providerService := service.NewVideoService(providers, log)
	healthCheck := controller.NewHealthHandler(cfg)
	mediaController := controller.NewMediaController(mediaRepository)

	downloadService := processor.NewDownloadService(cfg.NumWorkers, kafkaConsumer, mediaRepository, providerService, coreService, log)

	gin.SetMode(cfg.GinConfig.Mode)
	r := gin.New()
	router.SetupRoutes(r, healthCheck, mediaController, log)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	errChan := make(chan error, 2)

	go func() {
		log.Info("Iniciando servidor HTTP en :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	go func() {
		if err := downloadService.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		log.Error("Error en el servicio", zap.Error(err))
		return err
	case sig := <-sigChan:
		log.Info("Señal recibida, iniciando shutdown", zap.String("signal", sig.String()))

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("Error durante el shutdown del servidor HTTP", zap.Error(err))
			return err
		}

		cancel()
		<-shutdownCtx.Done()
		log.Info("Servicio cerrado con éxito")
	}

	return nil
}
