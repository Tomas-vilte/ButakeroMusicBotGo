package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type SongDownloader interface {
	DownloadSong(ctx context.Context, songName string) (*entity.DownloadResponse, error)
}
