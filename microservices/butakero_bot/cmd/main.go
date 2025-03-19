package main

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
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
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
		Brokers: []string{"localhost:9092"},
		Topic:   "notifications",
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
		BaseURL:         "http://localhost:8080",
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
				Address: "localhost",
				Port:    27017,
			},
		},
		ReplicaSet:    "rs0",
		DirectConnect: true,
		Username:      "root",
		Password:      "root",
		Database:      "audio_service_db",
	}, logger)
	err = connManager.Connect(ctx)
	if err != nil {
		panic(err)
	}
	collection := connManager.GetDatabase().Collection("Songs")
	externalService := service.NewExternalAudioService(audioClient, logger)
	songRepo, err := mongodb.NewMongoDBSongRepository(mongodb.Options{
		Collection: collection,
		Logger:     logger,
	})
	if err != nil {
		panic(err)
	}
	discordClient, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	discordMessenger := discord.NewDiscordMessengerService(discordClient, logger)
	cfg := &config.Config{
		CommandPrefix: "test",
	}

	storageAudio, err := local_storage.NewLocalStorage(&config.Config{
		Storage: config.StorageConfig{
			LocalConfig: config.LocalConfig{
				Directory: "/",
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
	defer discordClient.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Cerrando bot...")
}
