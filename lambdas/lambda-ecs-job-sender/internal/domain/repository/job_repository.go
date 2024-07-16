package repository

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/entity"
)

// JobRepository define los métodos que cualquier repositorio de trabajos debe implementar.
type JobRepository interface {
	Update(ctx context.Context, job entity.Job) error
	GetJobByS3Key(ctx context.Context, s3Key string) (entity.Job, error)
}
