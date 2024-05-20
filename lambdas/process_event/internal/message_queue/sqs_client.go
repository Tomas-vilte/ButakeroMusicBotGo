package message_queue

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// SQSClient define los métodos necesarios para interactuar con Amazon SQS.
type SQSClient interface {
	SendMessageWithContext(ctx aws.Context, input *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error)
}

// Client es una implementación de SQSClient que interactúa con Amazon SQS.
type Client struct {
	sqsClient *sqs.SQS
}

// NewSQSClient crea una nueva instancia de Client que se comunica con Amazon SQS.
func NewSQSClient() *Client {
	sess := session.Must(session.NewSession())
	return &Client{
		sqsClient: sqs.New(sess),
	}
}

// SendMessageWithContext envía un mensaje a una cola de mensajes de Amazon SQS utilizando el cliente SQS proporcionado.
func (c *Client) SendMessageWithContext(ctx aws.Context, input *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error) {
	return c.sqsClient.SendMessageWithContext(ctx, input, opts...)
}
