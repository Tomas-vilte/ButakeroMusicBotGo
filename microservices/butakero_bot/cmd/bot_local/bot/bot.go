package bot

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/application/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/commands"
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
	voiceStateService := discord.NewVoiceStateService(logger)
	eventsHandler := events.NewEventHandler(cfg, logger, discordMessenger, storageAudio, voiceStateService)
	handler := commands.NewCommandHandler(
		interactionStorage,
		logger,
		songService,
		discordMessenger,
		eventsHandler,
	)

	commandHandler := discord.NewSlashCommandRouter(cfg.CommandPrefix).
		PlayHandler(handler.PlaySong).
		SkipHandler(handler.SkipSong).
		StopHandler(handler.StopPlaying).
		ListHandler(handler.ListPlaylist).
		RemoveHandler(handler.RemoveSong).
		PlayingNowHandler(handler.GetPlayingSong)

	eventsHandler.RegisterEventHandlers(discordClient, ctx)
	discordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			if h, ok := commandHandler.GetComponentHandlers()[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}

		default:
			if h, ok := commandHandler.GetCommandHandlers()[i.ApplicationCommandData().Name]; ok {
				h(ctx, s, i)
			}
		}
	})
	discordClient.Identify.Intents = discordgo.IntentsAll

	err = discordClient.Open()
	if err != nil {
		logger.Error("Error al abrir conexión con Discord", zap.Error(err))
	}
	defer func() {
		if err := discordClient.Close(); err != nil {
			logger.Error("Error al cerrar conexión con Discord", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Cerrando bot...")

	return err
}
