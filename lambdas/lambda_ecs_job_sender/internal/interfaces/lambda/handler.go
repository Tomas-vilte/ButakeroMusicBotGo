package lambda

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/entity"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/usecase"
	logging "github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/pkg/logger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
	"path"
	"strings"
	"time"
)

type Handler struct {
	sendJobsUseCase usecase.JobSender
	logger          logging.Logger
}

func NewHandler(useCase usecase.JobSender, logger logging.Logger) *Handler {
	return &Handler{
		sendJobsUseCase: useCase,
		logger:          logger,
	}
}

func (h *Handler) Handle(ctx context.Context, event events.S3Event) error {
	for _, record := range event.Records {
		originalKey := record.S3.Object.Key
		h.logger.Info("Procesando evento de audio", zap.String("originalKey", originalKey), zap.String("decodedKey", record.S3.Object.URLDecodedKey))

		fileName := path.Base(record.S3.Object.URLDecodedKey)
		dcaFileName := strings.TrimSuffix(fileName, path.Ext(fileName)) + ".dca"
		job := entity.Job{
			ID:                   uuid.New().String(),
			S3Key:                record.S3.Object.URLDecodedKey,
			KEY:                  dcaFileName,
			Status:               "PENDING",
			BucketName:           record.S3.Bucket.Name,
			Region:               record.AWSRegion,
			TaskDefinition:       os.Getenv("TASK_DEFINITION"),
			ClusterName:          os.Getenv("CLUSTER_NAME"),
			SecurityGroup:        os.Getenv("SECURITY_GROUP"),
			TaskRoleArn:          os.Getenv("TASK_ROLE_ARN"),
			TaskExecutionRoleArn: os.Getenv("TASK_EXECUTION_ARN"),
			QueueURL:             os.Getenv("SQS_QUEUE_URL"),
			Subnets:              parseSubnets(os.Getenv("SUBNETS")),
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		if err := h.sendJobsUseCase.Execute(ctx, job); err != nil {
			h.logger.Error("Error al enviar el job a ECS", zap.Error(err))
			return err
		}
	}
	return nil
}

func parseSubnets(subnetString string) []string {
	return strings.Split(strings.TrimSpace(subnetString), ",")
}
