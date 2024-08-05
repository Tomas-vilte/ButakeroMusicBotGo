package s3_audio

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/mock"
	"io"
)

type MockS3Uploader struct {
	mock.Mock
}

func (m *MockS3Uploader) UploadDCA(ctx context.Context, audioData io.Reader, key string) error {
	args := m.Called(ctx, audioData, key)
	return args.Error(0)
}

func (m *MockS3Uploader) UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*s3manager.UploadOutput), args.Error(1)
}

func (m *MockS3Uploader) FileExists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockS3Uploader) DownloadDCA(ctx context.Context, key string) (io.Reader, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*bytes.Reader), args.Error(1)
}

type MockS3Downloader struct {
	mock.Mock
}

func (m *MockS3Downloader) DownloadWithContext(ctx aws.Context, w io.WriterAt, input *s3.GetObjectInput, opts ...func(*s3manager.Downloader)) (int64, error) {
	args := m.Called(ctx, w, input, opts)
	return args.Get(0).(int64), args.Error(1)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) HeadObjectWithContext(ctx aws.Context, input *s3.HeadObjectInput, opts ...request.Option) (*s3.HeadObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.HeadObjectOutput), args.Error(1)
}
