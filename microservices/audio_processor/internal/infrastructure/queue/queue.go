package queue

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type (
	// MessageQueue es la interfaz que debe implementar cualquier servicio de mensajeria
	MessageQueue interface {
		SendMessage(ctx context.Context, message Message) error
		ReceiveMessage(ctx context.Context) ([]Message, error)
		DeleteMessage(ctx context.Context, receiptHandle string) error
	}

	SQSClientInterface interface {
		SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
		ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
		DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	}

	// Message seria la estructura de un mensaje
	Message struct {
		ID            string
		Content       string
		ReceiptHandle string
	}

	MessageBody struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}
)
