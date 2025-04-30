package bot

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/application/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters/health"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/command"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/events"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/messenger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/storage"
	sqsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/messaging/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/storage/s3_storage"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	cfgAws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartBot() error {
	cfg, err := config.LoadConfigAws()
	if err != nil {
		return fmt.Errorf("error al cargar la configuración de AWS: %v", err)
	}

	logger, err := logging.NewProductionLogger()
	if err != nil {
		return fmt.Errorf("error al inicializar el logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			logger.Error("Error al cerrar el logger", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfgAppAws, err := cfgAws.LoadDefaultConfig(ctx, cfgAws.WithRegion(cfg.AWS.Region))
	if err != nil {
		logger.Error("Error al cargar la configuración de AWS", zap.Error(err))
		return fmt.Errorf("error en la configuración de AWS: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfgAppAws)
	messageConsumer := sqsApp.NewSQSConsumer(sqsClient, cfg, logger)

	go func() {
		if err := messageConsumer.SubscribeToDownloadEvents(ctx); err != nil {
			logger.Error("Error al consumir mensajes de SQS", zap.Error(err))
		}
	}()

	messageProducer, err := sqsApp.NewProducerSQS(cfg, logger)
	if err != nil {
		return fmt.Errorf("error al crear el productor SQS: %v", err)
	}
	defer func() {
		if err := messageProducer.ClosePublisher(); err != nil {
			logger.Error("Error al cerrar el productor SQS", zap.Error(err))
		}
	}()

	discordClient, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return fmt.Errorf("error al crear el cliente de Discord: %v", err)
	}

	storageAudio, err := s3_storage.NewS3Storage(cfg, logger)
	if err != nil {
		logger.Error("Error al crear el cliente de S3", zap.Error(err))
		return fmt.Errorf("error en el almacenamiento S3: %v", err)
	}

	discordMessenger := messenger.NewDiscordMessengerAdapter(discordClient, logger)
	interactionStorage := storage.NewInMemoryInteractionStorage(logger)

	mediaClient, err := api.NewMediaAPIClient(api.AudioAPIClientConfig{
		BaseURL: cfg.ExternalService.BaseURL,
		Timeout: 1 * time.Minute,
	}, logger)
	if err != nil {
		return fmt.Errorf("error al crear el cliente de la API de medios: %v", err)
	}

	songService := service.NewSongService(mediaClient, messageProducer, messageConsumer, logger)
	tracker := discord.NewBotChannelTracker(logger)
	mover := discord.NewBotMover(logger)
	playback := discord.NewPlaybackController(logger)
	voiceStateService := discord.NewVoiceStateService(tracker, mover, playback, logger)
	playerFactory := discord.NewGuildPlayerFactory(discordClient, storageAudio, discordMessenger, logger)
	guildManager := discord.NewGuildManager(playerFactory, logger)
	eventsHandler := events.NewEventHandler(guildManager, voiceStateService, logger, cfg)
	handler := command.NewCommandHandler(
		interactionStorage,
		logger,
		songService,
		guildManager,
		discordMessenger,
		eventsHandler,
	)

	commandRegistry := command.NewCommandRegistry()
	commands := []command.Command{
		command.NewPlayCommand(handler, logger),
		command.NewStopCommand(handler, logger),
		command.NewSkipCommand(handler, logger),
		command.NewListCommand(handler, logger),
		command.NewRemoveCommand(handler, logger),
		command.NewPlayingCommand(handler, logger),
		command.NewPauseCommand(handler, logger),
		command.NewResumeCommand(handler, logger),
	}

	eventsHandler.RegisterEventHandlers(discordClient)
	discordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			if h, ok := commandRegistry.GetCommandHandlers()[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
	})

	if err := discordClient.Open(); err != nil {
		logger.Error("Error al abrir la conexión con Discord", zap.Error(err))
		return fmt.Errorf("error al conectar con Discord: %v", err)
	}
	defer func() {
		if err := discordClient.Close(); err != nil {
			logger.Error("Error al cerrar la conexión con Discord", zap.Error(err))
		}
	}()

	discordChecker := health.NewDiscordChecker(discordClient, logger)
	serviceBChecker, err := health.NewServiceBChecker(&health.ServiceConfig{
		BaseURL: cfg.ExternalService.BaseURL,
		Timeout: 5 * time.Second,
	}, logger)
	if err != nil {
		logger.Error("Error al crear el verificador de salud de Service B", zap.Error(err))
		return fmt.Errorf("error al crear el verificador de salud de Service B: %v", err)
	}

	healthHandler := api.NewHealthHandler(discordChecker, serviceBChecker, logger, cfg)

	router := http.NewServeMux()
	router.HandleFunc("/api/v1/health", healthHandler.Handle)

	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	go func() {
		logger.Info("Iniciando servidor HTTP en el puerto 8081")
		if err := server.ListenAndServe(); err != nil {
			logger.Error("Error al iniciar el servidor HTTP", zap.Error(err))
		}
	}()

	commandRegistry.Register(command.NewRootCommand(cfg.CommandPrefix, commands, logger))
	if _, err := discordClient.ApplicationCommandBulkOverwrite(
		discordClient.State.User.ID,
		"",
		commandRegistry.GetCommands(),
	); err != nil {
		logger.Error("Error al registrar los comandos en Discord", zap.Error(err))
		return fmt.Errorf("error al registrar comandos en Discord: %v", err)
	}

	logger.Info("Bot iniciado correctamente y listo para recibir comandos")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Cerrando bot...")
	return nil
}
