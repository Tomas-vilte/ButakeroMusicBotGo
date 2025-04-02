package bot

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/application/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/command"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/events"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/storage"
	sqsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/messaging/sqs"
	dynamodbApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/repository/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/storage/s3_storage"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	cfgAws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartBot() error {
	cfg, err := config.LoadConfigAws()
	if err != nil {
		panic(err)
	}

	logger, err := logging.NewProductionLogger()
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

	cfgAppAws, err := cfgAws.LoadDefaultConfig(ctx, cfgAws.WithRegion(cfg.AWS.Region))
	if err != nil {
		logger.Error("Error al cargar la configuración de AWS", zap.Error(err))
		return err
	}

	sqsClient := sqs.NewFromConfig(cfgAppAws)

	messageConsumer := sqsApp.NewSQSConsumer(sqsClient, cfg, logger)
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
		logger.Error("Error al crear cliente de audio", zap.Error(err))
		return err
	}

	externalService := service.NewExternalAudioService(audioClient, logger)

	discordClient, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		panic(err)
	}

	discordMessenger := discord.NewDiscordMessengerAdapter(discordClient, logger)

	storageAudio, err := s3_storage.NewS3Storage(cfg, logger)
	if err != nil {
		logger.Error("Error al crear cliente S3", zap.Error(err))
		return err
	}

	interactionStorage := storage.NewInMemoryInteractionStorage(logger)

	dynamoClient := dynamodb.NewFromConfig(cfgAppAws)

	songRepo, err := dynamodbApp.NewDynamoSongRepository(dynamodbApp.Options{
		Logger:    logger,
		TableName: cfg.DatabaseConfig.DynamoDB.SongsTable,
		Client:    dynamoClient,
	})
	if err != nil {
		logger.Error("Error al crear repositorio de canciones", zap.Error(err))
		return err
	}

	songService := service.NewSongService(songRepo, externalService, messageConsumer, logger)
	voiceStateService := discord.NewVoiceStateService(logger)

	guildManager := discord.NewGuildManager(discordClient, storageAudio, discordMessenger, logger)
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
		panic(err)
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
