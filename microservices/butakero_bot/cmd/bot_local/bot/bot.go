package bot

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/application/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/command"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/events"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/storage"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/messaging/kafka"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/repository/mongodb"
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
		panic(err)
	}

	logger, err := logging.NewDevelopmentLogger()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			panic(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageConsumer, err := kafka.NewKafkaConsumer(kafka.KafkaConfig{
		Brokers: cfg.QueueConfig.KafkaConfig.Brokers,
		Topic:   cfg.QueueConfig.KafkaConfig.Topic,
		TLS: shared.TLSConfig{
			Enabled:  cfg.QueueConfig.KafkaConfig.TLS.Enabled,
			CAFile:   cfg.QueueConfig.KafkaConfig.TLS.CAFile,
			CertFile: cfg.QueueConfig.KafkaConfig.TLS.CertFile,
			KeyFile:  cfg.QueueConfig.KafkaConfig.TLS.KeyFile,
		},
	}, logger)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := messageConsumer.ConsumeMessages(ctx, -1); err != nil {
			logger.Error("Error al consumir mensajes de kafka")
		}
	}()
	audioClient, err := adapters.NewAudioAPIClient(adapters.AudioAPIClientConfig{
		BaseURL:         cfg.ExternalService.BaseURL,
		Timeout:         10 * time.Second,
		MaxIdleConns:    10,
		MaxConnsPerHost: 5,
	}, logger)
	if err != nil {
		panic(err)
	}
	connManager := mongodb.NewConnectionManager(mongodb.MongoConfig{
		Hosts: []mongodb.Host{
			{
				Address: cfg.DatabaseConfig.MongoDB.Hosts[0].Address,
				Port:    cfg.DatabaseConfig.MongoDB.Hosts[0].Port,
			},
		},
		ReplicaSet:    cfg.DatabaseConfig.MongoDB.ReplicaSet,
		DirectConnect: cfg.DatabaseConfig.MongoDB.DirectConnect,
		Username:      cfg.DatabaseConfig.MongoDB.Username,
		Password:      cfg.DatabaseConfig.MongoDB.Password,
		Database:      cfg.DatabaseConfig.MongoDB.Database,
		RetryWrites:   cfg.DatabaseConfig.MongoDB.RetryWrites,
		AuthSource:    cfg.DatabaseConfig.MongoDB.AuthSource,
		TLS: shared.TLSConfig{
			Enabled:  cfg.DatabaseConfig.MongoDB.TLS.Enabled,
			CAFile:   cfg.DatabaseConfig.MongoDB.TLS.CAFile,
			CertFile: cfg.DatabaseConfig.MongoDB.TLS.CertFile,
			KeyFile:  cfg.DatabaseConfig.MongoDB.TLS.KeyFile,
		},
		Timeout: cfg.DatabaseConfig.MongoDB.Timeout,
	}, logger)
	err = connManager.Connect(ctx)
	if err != nil {
		panic(err)
	}
	collection := connManager.GetDatabase().Collection(cfg.DatabaseConfig.MongoDB.Collection)
	externalService := service.NewExternalAudioService(audioClient, logger)
	songRepo, err := mongodb.NewMongoDBSongRepository(mongodb.Options{
		Collection: collection,
		Logger:     logger,
	})
	if err != nil {
		panic(err)
	}
	discordClient, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		panic(err)
	}

	discordMessenger := discord.NewDiscordMessengerAdapter(discordClient, logger)

	storageAudio := local_storage.NewLocalStorage(logger)

	interactionStorage := storage.NewInMemoryInteractionStorage(logger)

	songService := service.NewSongService(songRepo, externalService, messageConsumer, logger)
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

	err = discordClient.Open()
	if err != nil {
		logger.Error("Error al abrir conexión con Discord", zap.Error(err))
		return err
	}
	defer func() {
		if err := discordClient.Close(); err != nil {
			logger.Error("Error al cerrar conexión con Discord", zap.Error(err))
		}
	}()

	commandRegistry.Register(command.NewRootCommand(cfg.CommandPrefix, commands, logger))
	_, err = discordClient.ApplicationCommandBulkOverwrite(
		discordClient.State.User.ID,
		"",
		commandRegistry.GetCommands(),
	)
	if err != nil {
		logger.Error("No se pudieron registrar los comandos", zap.Error(err))
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Cerrando bot...")

	return err
}
