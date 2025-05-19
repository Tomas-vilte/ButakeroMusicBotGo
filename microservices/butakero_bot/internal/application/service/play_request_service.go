package service

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"go.uber.org/zap"
	"sync"
)

type PlayRequestManager struct {
	guildQueues  map[string]chan model.PlayRequestData
	mu           sync.Mutex
	songService  ports.SongService
	guildManager ports.GuildManager
	logger       logging.Logger
}

func NewPlayRequestManager(service ports.SongService, gm ports.GuildManager, logger logging.Logger) *PlayRequestManager {
	return &PlayRequestManager{
		guildQueues:  make(map[string]chan model.PlayRequestData),
		songService:  service,
		guildManager: gm,
		logger:       logger,
	}
}

func (prm *PlayRequestManager) Enqueue(guildID string, data model.PlayRequestData) <-chan model.PlayResult {
	data.ResultChan = make(chan model.PlayResult, 1)

	prm.mu.Lock()
	queue, exists := prm.guildQueues[guildID]
	if !exists {
		queue = make(chan model.PlayRequestData, 100)
		prm.guildQueues[guildID] = queue
		go prm.guildWorker(guildID, queue)
	}
	prm.mu.Unlock()

	queue <- data
	return data.ResultChan
}

func (prm *PlayRequestManager) guildWorker(guildID string, queue chan model.PlayRequestData) {
	for request := range queue {
		workerCtx := trace.WithTraceID(request.Ctx)
		log := prm.logger.With(zap.String("guildID", guildID), zap.String("traceID", trace.GetTraceID(workerCtx)))

		songEntity, err := prm.songService.GetOrDownloadSong(workerCtx, request.UserID, request.SongInput, "youtube")
		if err != nil {
			request.ResultChan <- model.PlayResult{
				Err:             fmt.Errorf("no se pudo obtener/descargar la canción: %w", err),
				RequestedByID:   request.UserID,
				RequestedByName: request.RequestedByName,
			}
			close(request.ResultChan)
			continue
		}

		guildPlayer, err := prm.guildManager.GetGuildPlayer(request.GuildID)
		if err != nil {
			log.Error("Error al obtener GuildPlayer", zap.Error(err))
			request.ResultChan <- model.PlayResult{
				Err:             fmt.Errorf("error al obtener GuildPlayer: %w", err),
				RequestedByID:   request.UserID,
				RequestedByName: request.RequestedByName,
			}
			close(request.ResultChan)
			continue
		}

		playedSong := &entity.PlayedSong{
			DiscordSong:     songEntity,
			RequestedByName: request.RequestedByName,
			RequestedByID:   request.UserID,
		}

		if err := guildPlayer.AddSong(workerCtx, &request.ChannelID, &request.VoiceChannelID, playedSong); err != nil {
			log.Error("Error al agregar canción a la cola del GuildPlayer", zap.Error(err), zap.String("songTitle", songEntity.TitleTrack))
			request.ResultChan <- model.PlayResult{
				SongTitle:       songEntity.TitleTrack,
				Err:             fmt.Errorf("no se pudo agregar la canción '%s' a la cola: %w", songEntity.TitleTrack, err),
				RequestedByID:   request.UserID,
				RequestedByName: request.RequestedByName,
			}
			close(request.ResultChan)
			continue
		}

		log.Info("Canción procesada y enviada a GuildPlayer", zap.String("songTitle", songEntity.TitleTrack))
		request.ResultChan <- model.PlayResult{
			SongTitle:       songEntity.TitleTrack,
			RequestedByID:   request.UserID,
			RequestedByName: request.RequestedByName,
		}
		close(request.ResultChan)
	}
	prm.mu.Lock()
	delete(prm.guildQueues, guildID)
	prm.mu.Unlock()
	prm.logger.Info("GuildWorker finalizado", zap.String("guildID", guildID))
}
