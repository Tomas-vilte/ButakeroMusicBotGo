package bot

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/application/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/command"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/events"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/messenger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/storage"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/messaging/kafka"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/storage/local"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartBot() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error al cargar la configuración: %v", err)
	}

	logger, err := logging.NewDevelopmentLogger()
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

	messageConsumer, err := kafka.NewKafkaConsumer(kafka.ConfigKafka{
		Brokers: cfg.QueueConfig.KafkaConfig.Brokers,
		Topic:   cfg.QueueConfig.KafkaConfig.Topics.BotDownloadStatus,
		Offset:  -1,
		TLS: shared.TLSConfig{
			Enabled:  cfg.QueueConfig.KafkaConfig.TLS.Enabled,
			CAFile:   cfg.QueueConfig.KafkaConfig.TLS.CAFile,
			CertFile: cfg.QueueConfig.KafkaConfig.TLS.CertFile,
			KeyFile:  cfg.QueueConfig.KafkaConfig.TLS.KeyFile,
		},
	}, logger)
	if err != nil {
		return fmt.Errorf("error al crear el consumidor de Kafka: %v", err)
	}

	go func() {
		if err := messageConsumer.SubscribeToDownloadEvents(ctx); err != nil {
			logger.Error("Error al suscribirse a eventos de descarga de Kafka", zap.Error(err))
		}
	}()
	defer func() {
		if err := messageConsumer.CloseSubscription(); err != nil {
			logger.Error("Error al cerrar la suscripción a Kafka", zap.Error(err))
		}
	}()

	messageProducer, err := kafka.NewProducerKafka(cfg, logger)
	if err != nil {
		return fmt.Errorf("error al crear el productor de Kafka: %v", err)
	}
	defer func() {
		if err := messageProducer.ClosePublisher(); err != nil {
			logger.Error("Error al cerrar el productor de Kafka", zap.Error(err))
		}
	}()

	discordClient, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return fmt.Errorf("error al crear el cliente de Discord: %v", err)
	}

	discordMessenger := messenger.NewDiscordMessengerAdapter(discordClient, logger)
	storageAudio := local_storage.NewLocalStorage(logger)
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
	voiceStateService := discord.NewVoiceStateService(tracker, mover, playback)
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

	eventsHandler.RegisterEventHandlers(discordClient, ctx)
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

	commandRegistry.Register(command.NewRootCommand(cfg.CommandPrefix, commands, logger))
	if _, err := discordClient.ApplicationCommandBulkOverwrite(
		discordClient.State.User.ID,
		"",
		commandRegistry.GetCommands(),
	); err != nil {
		logger.Error("Error al registrar los comandos en Discord", zap.Error(err))
		return fmt.Errorf("error al registrar comandos en Discord: %v", err)
	}

	logger.Info("Bot iniciado correctamente")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Cerrando bot...")
	return nil
}
