package inmemory

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"go.uber.org/zap"
	"sync"
	"time"
)

var _ ports.PlayerStateStorage = (*PlayerStateManager)(nil)

type PlayerStateManager struct {
	mu           sync.RWMutex
	currentTrack *entity.PlayedSong
	textChannel  string
	voiceChannel string
	logger       logging.Logger
}

func NewPlayerStateManager(logger logging.Logger) *PlayerStateManager {
	return &PlayerStateManager{
		mu:     sync.RWMutex{},
		logger: logger,
	}
}

func (s *PlayerStateManager) GetCurrentTrack(ctx context.Context) (*entity.PlayedSong, error) {
	logger := s.logger.With(
		zap.String("component", "player_state_storage"),
		zap.String("method", "GetCurrentTrack"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.currentTrack == nil {
		logger.Debug("No hay track en reproducción")
		return nil, nil
	}

	logger.Debug("Cancion actual obtenido",
		zap.String("song_id", s.currentTrack.DiscordSong.ID),
		zap.String("title", s.currentTrack.DiscordSong.TitleTrack),
		zap.Int64("position", s.currentTrack.Position),
		zap.String("requested_by", s.currentTrack.RequestedByName))

	return s.currentTrack, nil
}

func (s *PlayerStateManager) SetCurrentTrack(ctx context.Context, track *entity.PlayedSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := s.logger.With(
		zap.String("component", "PlayerStateManager"),
		zap.String("method", "SetCurrentTrack"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	if track != nil {
		if track.DiscordSong == nil {
			logger.Error("Track no puede tener DiscordSong nil")
			return errors_app.NewAppError(
				errors_app.ErrCodeInvalidSong,
				"La canción proporcionada no es válida",
				nil,
			)
		}

		if track.DiscordSong.ID == "" {
			logger.Warn("Track sin ID asignado, generando uno nuevo")
			track.DiscordSong.ID = generateTrackID()
		}

		if track.DiscordSong.AddedAt.IsZero() {
			track.DiscordSong.AddedAt = time.Now()
			logger.Debug("Asignada fecha de agregado automática")
		}

		logger = logger.With(
			zap.String("track_id", track.DiscordSong.ID),
			zap.String("title", track.DiscordSong.TitleTrack),
			zap.Int64("duration", track.DiscordSong.DurationMs),
			zap.Time("added_at", track.DiscordSong.AddedAt),
		)
	}

	s.currentTrack = track

	if track == nil {
		logger.Info("Cancion actual limpiado")
	} else {
		logger.Info("Cancion actual actualizado",
			zap.Int64("position", track.Position),
			zap.String("requested_by", track.RequestedByName),
		)
	}

	return nil
}

func (s *PlayerStateManager) GetVoiceChannelID(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logger := s.logger.With(
		zap.String("component", "PlayerStateManager"),
		zap.String("method", "GetVoiceChannelID"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	logger.Debug("Obteniendo ID de canal de voz")

	if s.voiceChannel == "" {
		logger.Warn("No hay canal de voz configurado")
		return "", nil
	}

	return s.voiceChannel, nil
}

func (s *PlayerStateManager) SetVoiceChannelID(ctx context.Context, channelID string) error {
	logger := s.logger.With(
		zap.String("component", "PlayerStateManager"),
		zap.String("method", "SetVoiceChannelID"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("channel_id", channelID),
	)

	if channelID == "" {
		logger.Error("ID de canal inválido")
		return errors_app.NewAppError(
			errors_app.ErrCodeInvalidGuildID,
			"El ID del canal proporcionado no es válido",
			nil,
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.voiceChannel = channelID

	logger.Info("Canal de voz actualizado exitosamente")
	return nil
}

func (s *PlayerStateManager) GetTextChannelID(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logger := s.logger.With(
		zap.String("component", "PlayerStateManager"),
		zap.String("method", "GetTextChannelID"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	logger.Debug("Obteniendo ID de canal de texto")

	if s.textChannel == "" {
		logger.Warn("No hay canal de texto configurado")
		return "", nil
	}

	return s.textChannel, nil
}

func (s *PlayerStateManager) SetTextChannelID(ctx context.Context, channelID string) error {
	logger := s.logger.With(
		zap.String("component", "PlayerStateManager"),
		zap.String("method", "SetTextChannelID"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("channel_id", channelID),
	)

	if channelID == "" {
		logger.Error("ID de canal inválido")
		return errors_app.NewAppError(
			errors_app.ErrCodeInvalidGuildID,
			"El ID del canal proporcionado no es válido",
			nil,
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.textChannel = channelID

	logger.Info("Canal de texto actualizado exitosamente")
	return nil
}

// generateTrackID genera un ID único para la canción actual
func generateTrackID() string {
	return fmt.Sprintf("track_%d", time.Now().UnixNano())
}
