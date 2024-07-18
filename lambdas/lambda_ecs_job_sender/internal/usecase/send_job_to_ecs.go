package usecase

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/entity"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/repository"
	logging "github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"os"
)

type (
	ECSClient interface {
		RunTask(ctx context.Context, input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error)
	}

	S3Client interface {
		GetObject(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	}

	SendJobToECS struct {
		ecsClient ECSClient
		s3Client  S3Client
		jobRepo   repository.JobRepository
		logger    logging.Logger
	}

	JobSender interface {
		Execute(ctx context.Context, job entity.Job) error
	}
)

func NewSendJobsToECS(ecs ECSClient, s3 S3Client, repo repository.JobRepository, log logging.Logger) *SendJobToECS {
	return &SendJobToECS{
		ecsClient: ecs,
		s3Client:  s3,
		jobRepo:   repo,
		logger:    log,
	}
}

func (s *SendJobToECS) Execute(ctx context.Context, job entity.Job) error {
	s.logger.Info("Comenzando a procesar el trabajo", zap.String("jobID", job.ID))

	// Verificar si el trabajo ya existe en la base de datos
	existingJob, err := s.jobRepo.GetJobByS3Key(ctx, job.KEY)
	if err != nil {
		s.logger.Error("Error al verificar la existencia del job", zap.Error(err))
		return err
	}

	if existingJob.ID != "" {
		s.logger.Info("Job ya existe y ha sido procesado previamente", zap.String("jobID", existingJob.ID))
		return nil
	}

	// Verificamos si existe el objeto en S3
	_, err = s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(job.BucketName),
		Key:    aws.String(job.S3Key),
	})
	if err != nil {
		s.logger.Error("Error al obtener el objeto desde S3", zap.Error(err))
		return err
	}

	// Enviamos tarea a ECS
	runTaskOutput, err := s.ecsClient.RunTask(ctx, &ecs.RunTaskInput{
		TaskDefinition: aws.String(job.TaskDefinition),
		Cluster:        aws.String(job.ClusterName),
		LaunchType:     types.LaunchTypeFargate,
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				AssignPublicIp: types.AssignPublicIpEnabled,
				Subnets:        job.Subnets,
				SecurityGroups: []string{job.SecurityGroup},
			},
		},
		Overrides: &types.TaskOverride{
			TaskRoleArn:      aws.String(job.TaskRoleArn),
			ExecutionRoleArn: aws.String(job.TaskExecutionRoleArn),
			ContainerOverrides: []types.ContainerOverride{
				{
					Name: aws.String("audio-processor"),
					Environment: []types.KeyValuePair{
						{
							Name:  aws.String("INPUT_FILE_FROM_S3"),
							Value: aws.String(job.S3Key),
						},
						{
							Name:  aws.String("BUCKET_NAME"),
							Value: aws.String(job.BucketName),
						},
						{
							Name:  aws.String("ACCESS_KEY"),
							Value: aws.String(os.Getenv("ACCESS_KEY")),
						},
						{
							Name:  aws.String("SECRET_KEY"),
							Value: aws.String(os.Getenv("SECRET_KEY")),
						},
						{
							Name:  aws.String("REGION"),
							Value: aws.String(job.Region),
						},
						{
							Name:  aws.String("KEY"),
							Value: aws.String(job.KEY),
						},
					},
				},
			},
		},
	})
	if err != nil {
		s.logger.Error("Error al enviar la tarea a ECS", zap.Error(err))
		return err
	}

	s.logger.Info("Tarea enviada a ECS", zap.Any("output", runTaskOutput))

	// Actualizar estado del job
	job.Status = "PROCESSING"
	err = s.jobRepo.Update(ctx, job)
	if err != nil {
		s.logger.Error("Error al actualizar estado del job", zap.Error(err))
		return err
	}

	s.logger.Info("Job enviado exitosamente a ECS")
	return nil
}
