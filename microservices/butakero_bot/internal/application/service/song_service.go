package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/queue"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"regexp"
	"sync"
	"time"
)

type SongService struct {
	mediaClient     ports.MediaClient
	messageProducer ports.SongDownloadRequestPublisher
	messageConsumer ports.SongDownloadEventSubscriber
	logger          logging.Logger

	responseChannels map[string]chan *queue.DownloadStatusMessage
	mu               sync.Mutex
	stopCh           chan struct{}
	wg               sync.WaitGroup
}

func NewSongService(
	mediaClient ports.MediaClient,
	messageProducer ports.SongDownloadRequestPublisher,
	messageConsumer ports.SongDownloadEventSubscriber,
	logger logging.Logger,
) *SongService {
	s := &SongService{
		mediaClient:      mediaClient,
		messageProducer:  messageProducer,
		messageConsumer:  messageConsumer,
		logger:           logger,
		responseChannels: make(map[string]chan *queue.DownloadStatusMessage),
		stopCh:           make(chan struct{}),
	}
	s.wg.Add(1)
	go s.listenForDownloadEvents()
	return s
}

func (s *SongService) extractURLOrTitle(input string) (string, bool) {
	urlRegex := regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.?be)/.+$`)
	return input, urlRegex.MatchString(input)
}

func (s *SongService) Close() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *SongService) listenForDownloadEvents() {
	defer s.wg.Done()
	msgChan := s.messageConsumer.DownloadEventsChannel()
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				s.logger.Info("El canal de eventos de descarga se cerró.")
				return
			}
			s.mu.Lock()
			respCh, exists := s.responseChannels[msg.RequestID]
			s.mu.Unlock()

			if exists {
				select {
				case respCh <- msg:
				default:
					s.logger.Warn("No se pudo enviar el evento de descarga al canal de respuesta, puede estar lleno o el receptor ya no existe",
						zap.String("requestID", msg.RequestID))
				}
			} else {
				s.logger.Warn("Se recibió un evento de descarga para un requestID desconocido o expirado",
					zap.String("requestID", msg.RequestID))
			}
		case <-s.stopCh:
			s.logger.Info("Deteniendo el receptor de eventos de descarga.")
			return
		}
	}
}

func (s *SongService) GetSongFromAPI(ctx context.Context, input string) (*entity.DiscordEntity, error) {
	_, isURL := s.extractURLOrTitle(input)

	if isURL {
		videoID := extractVideoID(input)
		if videoID != "" {
			media, err := s.mediaClient.GetMediaByID(ctx, videoID)
			if err == nil && media != nil {
				s.logger.Info("Media encontrada a través de la API por videoID",
					zap.String("videoID", videoID))
				return mediaToDiscordEntity(media), nil
			}
		}
	} else {
		mediaList, err := s.mediaClient.SearchMediaByTitle(ctx, input)
		if err == nil && len(mediaList) > 0 {
			s.logger.Info("Media encontrada a través de la API por título",
				zap.String("title", mediaList[0].Metadata.Title))
			return mediaToDiscordEntity(mediaList[0]), nil
		}
	}

	return nil, fmt.Errorf("canción no encontrada en la API")
}

func (s *SongService) DownloadSongViaQueue(ctx context.Context, userID, input, providerType string) (*entity.DiscordEntity, error) {
	requestID := uuid.New().String()

	s.logger.Info("Enviando solicitud a través de la queue",
		zap.String("input", input),
		zap.String("requestID", requestID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("userID", userID))

	responseChan := make(chan *queue.DownloadStatusMessage, 1)

	s.mu.Lock()
	s.responseChannels[requestID] = responseChan
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.responseChannels, requestID)
		s.mu.Unlock()
		close(responseChan)
	}()

	message := &queue.DownloadRequestMessage{
		RequestID:    requestID,
		UserID:       userID,
		Song:         input,
		ProviderType: providerType,
		Timestamp:    time.Now(),
	}

	if err := s.messageProducer.PublishDownloadRequest(ctx, message); err != nil {
		s.logger.Error("Error al publicar mensaje en la queue",
			zap.String("input", input),
			zap.Error(err))
		return nil, fmt.Errorf("error al solicitar la descarga: %w", err)
	}

	downloadCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	return s.waitForDownloadResponse(downloadCtx, requestID, responseChan)
}

func (s *SongService) waitForDownloadResponse(ctx context.Context, requestID string, msgChan <-chan *queue.DownloadStatusMessage) (*entity.DiscordEntity, error) {
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				s.logger.Warn("El canal de respuesta se cerró inesperadamente", zap.String("requestID", requestID))
				return nil, fmt.Errorf("se cerró el canal de respuesta para el requestID %s", requestID)
			}

			s.logger.Info("Mensaje recibido para solicitud actual",
				zap.String("requestID", requestID),
				zap.String("status", msg.Status))

			if msg.Status == "success" {
				s.logger.Info("Descarga completada exitosamente",
					zap.String("requestID", requestID),
					zap.String("video_id", msg.VideoID))
				return &entity.DiscordEntity{
					ID:           msg.VideoID,
					TitleTrack:   msg.PlatformMetadata.Title,
					DurationMs:   msg.PlatformMetadata.DurationMs,
					ThumbnailURL: msg.PlatformMetadata.ThumbnailURL,
					Platform:     msg.PlatformMetadata.Platform,
					FilePath:     msg.FileData.FilePath,
					URL:          msg.PlatformMetadata.URL,
					AddedAt:      time.Now(),
				}, nil
			} else {
				s.logger.Error("Error en la descarga",
					zap.String("requestID", requestID),
					zap.String("error", msg.Message))
				return nil, fmt.Errorf("error en la descarga: %s", msg.Message)
			}
		case <-ctx.Done():
			err := ctx.Err()
			if errors.Is(err, context.DeadlineExceeded) {
				s.logger.Error("Tiempo de espera agotado para la descarga",
					zap.String("requestID", requestID))
				return nil, fmt.Errorf("tiempo de espera agotado para la descarga (requestID: %s): %w", requestID, err)
			}
			s.logger.Error("Contexto cancelado durante la espera de la descarga",
				zap.String("requestID", requestID),
				zap.Error(err))
			return nil, err
		}
	}
}

func (s *SongService) GetOrDownloadSong(ctx context.Context, userID, songInput, providerType string) (*entity.DiscordEntity, error) {
	song, err := s.GetSongFromAPI(ctx, songInput)
	if err == nil {
		return song, nil
	}

	s.logger.Info("Media no encontrada a través de la API, intentando descarga",
		zap.String("input", songInput),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("userID", userID))

	return s.DownloadSongViaQueue(ctx, userID, songInput, providerType)
}

func mediaToDiscordEntity(media *model.Media) *entity.DiscordEntity {
	return &entity.DiscordEntity{
		TitleTrack:   media.Metadata.Title,
		DurationMs:   media.Metadata.DurationMs,
		Platform:     media.Metadata.Platform,
		FilePath:     media.FileData.FilePath,
		ThumbnailURL: media.Metadata.ThumbnailURL,
		URL:          media.Metadata.URL,
	}
}

func extractVideoID(url string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`youtube\.com/watch\?v=([^&]+)`),
		regexp.MustCompile(`youtu\.be/([^?]+)`),
		regexp.MustCompile(`youtube\.com/embed/([^?]+)`),
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(url); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}
