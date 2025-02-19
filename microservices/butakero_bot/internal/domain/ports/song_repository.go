package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type SongRepository interface {
	GetSongByVideoID(ctx context.Context, videoID string) (*entity.Song, error)
	SearchSongsByTitle(ctx context.Context, title string) ([]*entity.Song, error)
}
