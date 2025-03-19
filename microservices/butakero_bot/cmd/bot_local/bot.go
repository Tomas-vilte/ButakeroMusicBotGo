package bot_local

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/application/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/interactions"
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
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
		TLS: shared.TLSConfig{
			Enabled:  cfg.Kafka.TLS.Enabled,
			CAFile:   cfg.Kafka.TLS.CAFile,
			CertFile: cfg.Kafka.TLS.CertFile,
			KeyFile:  cfg.Kafka.TLS.KeyFile,
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
				Address: cfg.MongoDB.Hosts[0].Address,
				Port:    cfg.MongoDB.Hosts[0].Port,
			},
		},
		ReplicaSet:    cfg.MongoDB.ReplicaSet,
		DirectConnect: cfg.MongoDB.DirectConnect,
		Username:      cfg.MongoDB.Username,
		Password:      cfg.MongoDB.Password,
		Database:      cfg.MongoDB.Database,
		RetryWrites:   cfg.MongoDB.RetryWrites,
		AuthSource:    cfg.MongoDB.AuthSource,
		TLS: shared.TLSConfig{
			Enabled:  cfg.MongoDB.TLS.Enabled,
			CAFile:   cfg.MongoDB.TLS.CAFile,
			CertFile: cfg.MongoDB.TLS.CertFile,
			KeyFile:  cfg.MongoDB.TLS.KeyFile,
		},
		Timeout: cfg.MongoDB.Timeout,
	}, logger)
	err = connManager.Connect(ctx)
	if err != nil {
		panic(err)
	}
	collection := connManager.GetDatabase().Collection(cfg.MongoDB.Collection)
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

	discordMessenger := discord.NewDiscordMessengerService(discordClient, logger)

	storageAudio, err := local_storage.NewLocalStorage(&config.Config{
		Storage: config.StorageConfig{
			LocalConfig: config.LocalConfig{
				Directory: cfg.Storage.LocalConfig.Directory,
			},
		},
	}, logger)
	if err != nil {
		logger.Error("Error al crear storage de audio", zap.Error(err))
	}

	interactionStorage := storage.NewInMemoryInteractionStorage(logger)

	presenceNotifier := discord.NewDiscordPresenceNotifier()

	songService := service.NewSongService(songRepo, externalService, messageConsumer, logger)
	handler := interactions.NewInteractionHandler(
		interactionStorage,
		cfg,
		logger,
		discordMessenger,
		storageAudio,
		presenceNotifier,
		songService,
	)

	commandHandler := discord.NewSlashCommandRouter(cfg.CommandPrefix).
		PlayHandler(handler.PlaySong).
		SkipHandler(handler.SkipSong).
		StopHandler(handler.StopPlaying).
		ListHandler(handler.ListPlaylist).
		RemoveHandler(handler.RemoveSong).
		PlayingNowHandler(handler.GetPlayingSong).
		AddSongOrPlaylistHandler(handler.AddSong)

	handler.RegisterEventHandlers(discordClient, ctx)
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
		handler.StartPresenceCheck(s)
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
