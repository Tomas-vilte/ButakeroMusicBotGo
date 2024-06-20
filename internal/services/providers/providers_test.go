package providers

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/internal/services/providers/youtube_provider"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/youtube/v3"
	"testing"
)

func TestNewRealYouTubeClient(t *testing.T) {
	apiKey := "valid_api_key"

	client, err := youtube_provider.NewRealYouTubeClient(apiKey)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Service)
}

func TestRealYouTubeClient_VideosListCall(t *testing.T) {
	// Arrange
	client, _ := youtube_provider.NewRealYouTubeClient("valid_api_key")
	ctx := context.Background()
	part := []string{"snippet"}

	// Act
	wrapper := client.VideosListCall(ctx, part)

	// Assert
	assert.NotNil(t, wrapper)
}

func TestRealYouTubeClient_SearchListCall(t *testing.T) {
	client, _ := youtube_provider.NewRealYouTubeClient("valid_api_key")
	ctx := context.Background()
	part := []string{"snippet"}

	wrapper := client.SearchListCall(ctx, part)

	assert.NotNil(t, wrapper)
}

func TestRealVideosListCallWrapper_Do(t *testing.T) {
	mockWrapper := &youtube_provider.VideosListCallWrapperMock{}
	expectedResponse := &youtube.VideoListResponse{}

	mockWrapper.On("Do").Return(expectedResponse, nil)

	response, err := mockWrapper.Do()

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	mockWrapper.AssertExpectations(t)
}
