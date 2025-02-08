package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type SongRepository interface {
	GetSongByID(ctx context.Context, id string) (*entity.Song, error)
}
