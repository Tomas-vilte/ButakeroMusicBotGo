package inmemory

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"sort"
	"sync"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

var (
	ErrNoSongs               = errors.New("no hay canciones disponibles")
	ErrRemoveInvalidPosition = errors.New("posición inválida")
)

type InmemoryPlaylistStorage struct {
	mu     sync.RWMutex
	songs  []*entity.PlayedSong
	logger logging.Logger
}

func NewInmemoryPlaylistStorage(logger logging.Logger) *InmemoryPlaylistStorage {
	return &InmemoryPlaylistStorage{
		mu:     sync.RWMutex{},
		songs:  make([]*entity.PlayedSong, 0),
		logger: logger,
	}
}

func (s *InmemoryPlaylistStorage) AppendTrack(ctx context.Context, song *entity.PlayedSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := s.logger.With(
		zap.String("component", "InmemoryPlaylistStorage"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "AppendTrack"),
	)

	if song == nil || song.DiscordSong == nil {
		logger.Error("Intento de agregar canción inválida")
		return errors.New("canción inválida")
	}

	if song.DiscordSong.ID == "" {
		song.DiscordSong.ID = generateSongID()
	}
	if song.DiscordSong.AddedAt.IsZero() {
		song.DiscordSong.AddedAt = time.Now()
	}

	s.songs = append(s.songs, song)

	sort.Slice(s.songs, func(i, j int) bool {
		return s.songs[i].DiscordSong.AddedAt.Before(s.songs[j].DiscordSong.AddedAt)
	})

	logger.Info("Canción agregada al final",
		zap.String("song_id", song.DiscordSong.ID),
		zap.String("title", song.DiscordSong.TitleTrack),
		zap.Time("added_at", song.DiscordSong.AddedAt),
		zap.Int("new_length", len(s.songs)))
	return nil
}

func (s *InmemoryPlaylistStorage) RemoveTrack(ctx context.Context, position int) (*entity.PlayedSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := position - 1

	logger := s.logger.With(
		zap.String("component", "InmemoryPlaylistStorage"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "RemoveTrack"),
		zap.Int("position", position),
	)

	if index >= len(s.songs) || index < 0 {
		logger.Error("Posición inválida",
			zap.Int("playlist_length", len(s.songs)))
		return nil, ErrRemoveInvalidPosition
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

func (s *InmemoryPlaylistStorage) ClearPlaylist(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := s.logger.With(
		zap.String("component", "InmemoryPlaylistStorage"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "ClearPlaylist"),
	)

	count := len(s.songs)
	s.songs = make([]*entity.PlayedSong, 0)

	logger.Info("Playlist limpiada",
		zap.Int("songs_removed", count))
	return nil
}

func (s *InmemoryPlaylistStorage) GetAllTracks(ctx context.Context) ([]*entity.PlayedSong, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logger := s.logger.With(
		zap.String("component", "InmemoryPlaylistStorage"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "GetAllTracks"),
	)

	songs := make([]*entity.PlayedSong, len(s.songs))
	copy(songs, s.songs)

	logger.Debug("Playlist obtenida",
		zap.Int("count", len(songs)))
	return songs, nil
}

func (s *InmemoryPlaylistStorage) PopNextTrack(ctx context.Context) (*entity.PlayedSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := s.logger.With(
		zap.String("component", "InmemoryPlaylistStorage"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "PopNextTrack"),
	)

	if len(s.songs) == 0 {
		logger.Debug("Intento de obtener canción de playlist vacía")
		return nil, ErrNoSongs
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
