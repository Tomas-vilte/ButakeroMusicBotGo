package main

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/application/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/client"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/decoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/storage"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/messaging/kafka"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/repository/mongodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/storage/s3_storage"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func RegisterCommands(s *discordgo.Session, guildID string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "Reproduce una canción o la agrega a la lista de reproducción",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "input",
					Description: "URL o nombre de la canción",
					Required:    true,
				},
			},
		},
		{
			Name:        "stop",
			Description: "Detiene la reproducción actual",
		},
		{
			Name:        "skip",
			Description: "Salta a la siguiente canción en la lista de reproducción",
		},
		{
			Name:        "queue",
			Description: "Muestra la lista de reproducción actual",
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate("1231502131002085439", guildID, cmd)
		if err != nil {
			fmt.Println("Error al registrar comando:", err)
		}
	}
}

func main() {
	// Inicializa el bot y las dependencias
	discordClient, _ := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	logger, _ := logging.NewZapLogger()

	discordMessenger := discord.NewDiscordMessengerService(discordClient, logger)
	messageConsumer, err := kafka.NewKafkaConsumer(kafka.KafkaConfig{
		Brokers: []string{"localhost:29092"},
		Topic:   "notification",
	}, logger)
	if err != nil {
		panic(err)
	}
	audioClient, err := client.NewAudioAPIClient(client.AudioAPIClientConfig{
		BaseURL: "http://localhost:8080",
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
	err = connManager.Connect(context.Background())
	if err != nil {
		panic(err)
	}
	collection := connManager.GetDatabase().Collection("Songs")
	externalService := service.NewExternalAudioService(audioClient, logger)
	songRepo, err := mongodb.NewMongoDBSongRepository(mongodb.Options{
		Collection:     collection,
		Logger:         logger,
		CollectionName: "Songs",
		Database:       "audio_service_db",
	})
	if err != nil {
		panic(err)
	}
	songStorage := inmemory.NewInmemorySongStorage(logger)
	stateStorage := inmemory.NewInmemoryStateStorage(logger)
	storageAudio, err := s3_storage.NewS3Storage(&config.Config{
		AWS: config.AWSConfig{
			Region: "us-east-1",
		},
		Storage: config.StorageConfig{
			S3Config: config.S3Config{
				BucketName: "butakero-test",
			},
		},
	}, logger)
	decoderFactory := func(r io.ReadCloser) ports.Decoder {
		return decoder.NewOpusDecoder(r)
	}
	interactionStore := storage.NewInMemoryInteractionStorage(logger)
	voiceSessionFactory := func(guildID string) ports.VoiceSession {
		return voice.NewDiscordVoiceSession(discordClient, guildID, logger)
	}
	handler := service.NewInteractionHandler(
		interactionStore,
		discordMessenger,
		voiceSessionFactory,
		decoderFactory,
		messageConsumer,
		externalService,
		songRepo,
		songStorage,
		stateStorage,
		storageAudio,
	)

	// Maneja interacciones
	discordClient.AddHandler(handler.HandleInteraction)

	// Inicia el bot
	discordClient.Identify.Intents = discordgo.IntentsAll
	err = discordClient.Open()
	if err != nil {
		logger.Error("error al abrir la session de discord", zap.Error(err))
	}
	defer func(dg *discordgo.Session) {
		err := dg.Close()
		if err != nil {
			logger.Error("Hubo un error al cerrar session", zap.Error(err))
		}
	}(discordClient)
	defer discordClient.Close()

	// Mantén el bot en ejecución
	logger.Info("bot esta corriendo. Apreta ctrl - alt para salir")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
