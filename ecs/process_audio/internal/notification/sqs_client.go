package notification

import (
	"encoding/json"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
	"time"
)

type SQSNotifier struct {
	sqsClient *sqs.SQS
	queueURL  string
	logger    logging.Logger
}

type NotificationPayload struct {
	Key           string    `json:"key"`
	Success       bool      `json:"success"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	ProcessedAt   time.Time `json:"processed_at"`
	OriginalFile  string    `json:"original_file"`
	ProcessedFile string    `json:"processed_file"`
}

func NewSQSNotifier(sqsClient *sqs.SQS, queueURL string, logger logging.Logger) *SQSNotifier {
	return &SQSNotifier{
		sqsClient: sqsClient,
		queueURL:  queueURL,
		logger:    logger,
	}
}

func (n *SQSNotifier) NotifyProcessingResult(key, originalFile, processedFile string, success bool, errorMsg string) error {
	payload := NotificationPayload{
		Key:           key,
		Success:       success,
		ErrorMessage:  errorMsg,
		ProcessedAt:   time.Now(),
		OriginalFile:  originalFile,
		ProcessedFile: processedFile,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		n.logger.Error("Error al crear el JSON para notificacion SQS", zap.Error(err))
		return err
	}

	_, err = n.sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(jsonPayload)),
		QueueUrl:    aws.String(n.queueURL),
	})
	if err != nil {
		n.logger.Error("Error al enviar notificacion SQS", zap.Error(err))
		return err
	}

	n.logger.Info("Notificacion SQS enviada con exito",
		zap.String("key", key),
		zap.Bool("success", success),
		zap.String("original_file", originalFile),
		zap.String("processed_file", processedFile),
	)
	return nil
}
