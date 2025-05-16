package inmemory

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"sync"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

var _ ports.PlaylistStorage = (*MemoryPlaylistStore)(nil)

type MemoryPlaylistStore struct {
	mu     sync.RWMutex
	songs  []*entity.PlayedSong
	logger logging.Logger
}

func NewMemoryPlaylistStore(logger logging.Logger) *MemoryPlaylistStore {
	return &MemoryPlaylistStore{
		mu:     sync.RWMutex{},
		songs:  make([]*entity.PlayedSong, 0),
		logger: logger,
	}
}

func (s *MemoryPlaylistStore) AppendTrack(ctx context.Context, song *entity.PlayedSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := s.logger.With(
		zap.String("component", "MemoryPlaylistStore"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "AppendTrack"),
	)

	if song == nil || song.DiscordSong == nil {
		logger.Error("Intento de agregar canción inválida")
		return errors_app.NewAppError(errors_app.ErrCodeInvalidSong, "La canción proporcionada no es válida", nil)
	}

	if song.DiscordSong.ID == "" {
		song.DiscordSong.ID = generateSongID()
	}
	if song.DiscordSong.AddedAt.IsZero() {
		song.DiscordSong.AddedAt = time.Now()
	}

	s.songs = append(s.songs, song)

	logger.Info("Canción agregada al final",
		zap.String("song_id", song.DiscordSong.ID),
		zap.String("title", song.DiscordSong.TitleTrack),
		zap.Time("added_at", song.DiscordSong.AddedAt),
		zap.Int("new_length", len(s.songs)))
	return nil
}

func (s *MemoryPlaylistStore) RemoveTrack(ctx context.Context, position int) (*entity.PlayedSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := position - 1

	logger := s.logger.With(
		zap.String("component", "MemoryPlaylistStore"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "RemoveTrack"),
		zap.Int("position", position),
	)

	if index >= len(s.songs) || index < 0 {
		logger.Error("Posición inválida",
			zap.Int("playlist_length", len(s.songs)))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInvalidTrackPosition, "Posición de la canción inválida", nil)
	}

	song := s.songs[index]

	copy(s.songs[index:], s.songs[index+1:])
	s.songs[len(s.songs)-1] = nil
	s.songs = s.songs[:len(s.songs)-1]

	logger.Info("Canción removida",
		zap.String("song_id", song.DiscordSong.ID),
		zap.String("title", song.DiscordSong.TitleTrack),
		zap.Int("new_length", len(s.songs)))

	return song, nil
}

func (s *MemoryPlaylistStore) ClearPlaylist(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := s.logger.With(
		zap.String("component", "MemoryPlaylistStore"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "ClearPlaylist"),
	)

	count := len(s.songs)
	s.songs = make([]*entity.PlayedSong, 0)

	logger.Info("Playlist limpiada",
		zap.Int("songs_removed", count))
	return nil
}

func (s *MemoryPlaylistStore) GetAllTracks(ctx context.Context) ([]*entity.PlayedSong, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logger := s.logger.With(
		zap.String("component", "MemoryPlaylistStore"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "GetAllTracks"),
	)

	songs := make([]*entity.PlayedSong, len(s.songs))
	copy(songs, s.songs)

	logger.Debug("Playlist obtenida",
		zap.Int("count", len(songs)))
	return songs, nil
}

func (s *MemoryPlaylistStore) PopNextTrack(ctx context.Context) (*entity.PlayedSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := s.logger.With(
		zap.String("component", "MemoryPlaylistStore"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "PopNextTrack"),
	)

	if len(s.songs) == 0 {
		logger.Debug("Intento de obtener canción de playlist vacía")
		return nil, errors_app.NewAppError(errors_app.ErrCodePlaylistEmpty, "No hay canciones disponibles en la playlist", nil)
	}

	song := s.songs[0]
	s.songs = s.songs[1:]

	logger.Info("Primera canción obtenida",
		zap.String("song_id", song.DiscordSong.ID),
		zap.String("title", song.DiscordSong.TitleTrack),
		zap.Time("added_at", song.DiscordSong.AddedAt),
		zap.Int("remaining", len(s.songs)))

	return song, nil
}

// generateSongID genera un ID único para canciones que no lo tengan
func generateSongID() string {
	return fmt.Sprintf("song_%d", time.Now().UnixNano())
}
