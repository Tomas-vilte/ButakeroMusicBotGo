package s3uploader

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
)

// MockS3API es un mock del cliente S3API para pruebas.
type MockS3API struct {
	mock.Mock
	s3iface.S3API
}

// PutObjectWithContext implementa el método PutObjectWithContext del cliente S3API mock.
func (m *MockS3API) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}
