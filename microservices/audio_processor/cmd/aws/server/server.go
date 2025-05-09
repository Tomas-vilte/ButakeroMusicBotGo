package server

import (
	"context"
	"errors"
	"io"
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
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/cloud"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
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

	mediaRepository := dynamodb.NewMediaRepositoryDynamoDB(cfg, log)

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
			log.Error("Error al cerrar archivo", zap.Error(err))
		}
		if err := os.Remove(file.Name()); err != nil {
			log.Error("Error al eliminar archivo temporal", zap.Error(err))
		}
	}()

	if _, err = io.Copy(file, cookiesContent); err != nil {
		log.Error("Error al copiar contenido a archivo", zap.Error(err))
		return err
	}

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
	providerService := service.NewVideoService(providers, log)
	healthCheck := controller.NewHealthHandler(cfg)
	mediaController := controller.NewMediaController(mediaService)

	downloadService := processor.NewDownloadService(cfg.NumWorkers, sqsConsumer, mediaService, providerService, coreService, log)

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
