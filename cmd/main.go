package main

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/metrics"
	"github.com/Tomas-vilte/GoMusicBot/internal/music/fetcher"
	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

var (
	ctx            context.Context
	cancelCtx      context.CancelFunc
	cfg            = &config.Config{}
	youtubeFetcher *fetcher.YoutubeFetcher
	storage        *discord.InMemoryInteractionStorage
)

func main() {
	// Crear un nuevo logger usando la librería zap.
	logger, err := logging.NewZapLogger()
	if err != nil {
		panic("Error creando el logger: " + err.Error())
	}
	fmt.Println("Testtt")
	promRegistry := metrics.NewPrometheusRegistry()
	commandUsageCounter := metrics.NewCommandUsageCounter()
	promRegistry.Register(commandUsageCounter)

	promHTTPServer := metrics.NewPrometheusHTTPServer(":8080", promRegistry)

	go func() {
		if err := promHTTPServer.Start(); err != nil {
			logger.Error("Error al iniciar el servidor HTTP de métricas Prometheus: ", zap.Error(err))
		}
	}()
	defer func() {
		// Cerrar el logger cuando la función termine.
		err := logger.Close()
		if err != nil {
			logger.Error("Error cerrando el logger", zap.Error(err))
		}
	}()
	ctx, cancelCtx = context.WithCancel(context.Background())
	defer cancelCtx()
	if err := envconfig.Process("", cfg); err != nil {
		logger.Error("error al cargar las variables de entorno", zap.Error(err))
	}
	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		logger.Error("error al crear la session de messaging", zap.Error(err))
		return
	}
	storage = discord.NewInMemoryStorage()
	youtubeFetcher = fetcher.NewYoutubeFetcher(logger)
	responseHandler := discord.NewDiscordResponseHandler(logger)
	sessionService := discord.NewSessionService(dg)

	handler := discord.NewInteractionHandler(ctx, cfg.DiscordToken, responseHandler, sessionService, youtubeFetcher, storage, cfg, logger, commandUsageCounter).WithLogger(logger)
	commandHandler := discord.NewSlashCommandRouter(cfg.CommandPrefix).
		PlayHandler(handler.PlaySong).
		SkipHandler(handler.SkipSong).
		StopHandler(handler.StopPlaying).
		ListHandler(handler.ListPlaylist).
		RemoveHandler(handler.RemoveSong).
		PlayingNowHandler(handler.GetPlayingSong).
		AddSongOrPlaylistHandler(handler.AddSongOrPlaylist)

	handler.RegisterEventHandlers(dg)
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			if h, ok := commandHandler.GetComponentHandlers()[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}

		default:
			if h, ok := commandHandler.GetCommandHandlers()[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
		handler.CheckVoiceChannelsPresence()
	})
	dg.Identify.Intents = discordgo.IntentsAll
	err = dg.Open()
	if err != nil {
		logger.Error("error al abrir la session de messaging", zap.Error(err))
	}
	defer func(dg *discordgo.Session) {
		err := dg.Close()
		if err != nil {
			logger.Error("Hubo un error al cerrar session", zap.Error(err))
		}
	}(dg)
	slashCommands := commandHandler.GetSlashCommands()
	registeredCommands, err := dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, cfg.GuildID, slashCommands)
	if err != nil {
		logger.Error("no se pudo realizar el comando de sobrescritura masiva", zap.Error(err))
	}
	if cfg.GuildID != "" {
		defer func() {
			for _, cmd := range registeredCommands {
				if err := dg.ApplicationCommandDelete(dg.State.User.ID, cfg.GuildID, cmd.ID); err != nil {
					logger.Error("no se pudo realizar el comando de sobrescritura masiva", zap.String("command", cmd.Name), zap.Error(err))
				}
			}
		}()
	}
	logger.Info("bot esta corriendo. Apreta ctrl - alt para salir")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
