package youtube_provider

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/youtube/v3"
	"testing"
)

func TestSearchVideoID(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		clientMock := new(MockYouTubeClient)
		loggerMock := new(logging.MockLogger)
		searchCallMock := new(SearchListCallWrapperMock)

		clientMock.On("SearchListCall", mock.Anything, []string{"id"}).Return(searchCallMock)
		searchCallMock.On("Q", "test").Return(searchCallMock)
		searchCallMock.On("MaxResults", int64(1)).Return(searchCallMock)
		searchCallMock.On("Type", "video").Return(searchCallMock)
		loggerMock.On("Info", "Buscando video en YouTube", mock.Anything).Return()
		loggerMock.On("Info", "Video encontrado", mock.Anything).Return()

		response := &youtube.SearchListResponse{
			Items: []*youtube.SearchResult{
				{Id: &youtube.ResourceId{VideoId: "12345"}},
			},
		}
		searchCallMock.On("Do").Return(response, nil)

		provider := NewYouTubeProvider("dummyApiKey", loggerMock, clientMock)
		ctx := context.Background()
		videoID, err := provider.SearchVideoID(ctx, "test")
		assert.NoError(t, err)
		assert.Equal(t, "12345", videoID)

		clientMock.AssertExpectations(t)
		searchCallMock.AssertExpectations(t)
		loggerMock.AssertExpectations(t)
	})

	t.Run("search error", func(t *testing.T) {
		clientMock := new(MockYouTubeClient)
		loggerMock := new(logging.MockLogger)
		searchCallMock := new(SearchListCallWrapperMock)

		clientMock.On("SearchListCall", mock.Anything, []string{"id"}).Return(searchCallMock)
		searchCallMock.On("Q", "test").Return(searchCallMock)
		searchCallMock.On("MaxResults", int64(1)).Return(searchCallMock)
		searchCallMock.On("Type", "video").Return(searchCallMock)
		loggerMock.On("Info", "Buscando video en YouTube", mock.Anything).Return()
		loggerMock.On("Error", "Error al buscar vídeo en YouTube", mock.Anything).Return()

		searchCallMock.On("Do").Return(&youtube.SearchListResponse{}, errors.New("search error"))

		provider := NewYouTubeProvider("dummyApiKey", loggerMock, clientMock)
		ctx := context.Background()
		videoID, err := provider.SearchVideoID(ctx, "test")
		assert.Error(t, err)
		assert.Equal(t, "", videoID)
		assert.Contains(t, err.Error(), "search error")

		clientMock.AssertExpectations(t)
		searchCallMock.AssertExpectations(t)
		loggerMock.AssertExpectations(t)
	})

	t.Run("no video found", func(t *testing.T) {
		clientMock := new(MockYouTubeClient)
		loggerMock := new(logging.MockLogger)
		searchCallMock := new(SearchListCallWrapperMock)

		clientMock.On("SearchListCall", mock.Anything, []string{"id"}).Return(searchCallMock)
		searchCallMock.On("Q", "test").Return(searchCallMock)
		searchCallMock.On("MaxResults", int64(1)).Return(searchCallMock)
		searchCallMock.On("Type", "video").Return(searchCallMock)
		loggerMock.On("Info", "Buscando video en YouTube", mock.Anything).Return()
		loggerMock.On("Info", "No se encontró ningún vídeo para el término de búsqueda", mock.Anything).Return()

		response := &youtube.SearchListResponse{
			Items: []*youtube.SearchResult{},
		}
		searchCallMock.On("Do").Return(response, nil)

		provider := NewYouTubeProvider("dummyApiKey", loggerMock, clientMock)
		ctx := context.Background()
		videoID, err := provider.SearchVideoID(ctx, "test")
		assert.Error(t, err)
		assert.Equal(t, "", videoID)
		assert.Contains(t, err.Error(), "no se encontró ningún vídeo para el término de búsqueda")

		clientMock.AssertExpectations(t)
		searchCallMock.AssertExpectations(t)
		loggerMock.AssertExpectations(t)
	})
}

func TestGetVideoDetails(t *testing.T) {
	t.Run("successful get video details", func(t *testing.T) {
		clientMock := new(MockYouTubeClient)
		loggerMock := new(logging.MockLogger)
		videosCallMock := new(VideosListCallWrapperMock)

		clientMock.On("VideosListCall", mock.Anything, []string{"snippet", "contentDetails", "liveStreamingDetails"}).Return(videosCallMock)
		videosCallMock.On("Id", "12345").Return(videosCallMock)

		video := &youtube.Video{Snippet: &youtube.VideoSnippet{Title: "Test Video"}}
		response := &youtube.VideoListResponse{
			Items: []*youtube.Video{video},
		}
		videosCallMock.On("Do").Return(response, nil)
		loggerMock.On("Info", "Obteniendo detalles del video desde Youtube", mock.Anything).Return()
		loggerMock.On("Info", "Detalles del video recuperados exitosamente", mock.Anything).Return()

		provider := NewYouTubeProvider("dummyApiKey", loggerMock, clientMock)
		ctx := context.Background()
		videoDetails, err := provider.GetVideoDetails(ctx, "12345")
		assert.NoError(t, err)
		assert.Equal(t, "Test Video", videoDetails.Snippet.Title)

		clientMock.AssertExpectations(t)
		videosCallMock.AssertExpectations(t)
		loggerMock.AssertExpectations(t)
	})

	t.Run("get video details error", func(t *testing.T) {
		clientMock := new(MockYouTubeClient)
		loggerMock := new(logging.MockLogger)
		videosCallMock := new(VideosListCallWrapperMock)

		clientMock.On("VideosListCall", mock.Anything, []string{"snippet", "contentDetails", "liveStreamingDetails"}).Return(videosCallMock)
		videosCallMock.On("Id", "12345").Return(videosCallMock)

		videosCallMock.On("Do").Return(&youtube.VideoListResponse{}, errors.New("details error"))
		loggerMock.On("Info", "Obteniendo detalles del video desde Youtube", mock.Anything).Return()
		loggerMock.On("Error", "Error al obtener detalles del video desde Youtube", mock.Anything).Return()

		provider := NewYouTubeProvider("dummyApiKey", loggerMock, clientMock)
		ctx := context.Background()
		videoDetails, err := provider.GetVideoDetails(ctx, "12345")
		assert.Error(t, err)
		assert.Nil(t, videoDetails)
		assert.Contains(t, err.Error(), "details error")

		clientMock.AssertExpectations(t)
		videosCallMock.AssertExpectations(t)
		loggerMock.AssertExpectations(t)
	})

	t.Run("video not found", func(t *testing.T) {
		clientMock := new(MockYouTubeClient)
		loggerMock := new(logging.MockLogger)
		videosCallMock := new(VideosListCallWrapperMock)

		clientMock.On("VideosListCall", mock.Anything, []string{"snippet", "contentDetails", "liveStreamingDetails"}).Return(videosCallMock)
		videosCallMock.On("Id", "12345").Return(videosCallMock)

		response := &youtube.VideoListResponse{
			Items: []*youtube.Video{},
		}
		videosCallMock.On("Do").Return(response, nil)
		loggerMock.On("Info", "Obteniendo detalles del video desde Youtube", mock.Anything).Return()
		loggerMock.On("Info", "Video no encontrado", mock.Anything).Return()

		provider := NewYouTubeProvider("dummyApiKey", loggerMock, clientMock)
		ctx := context.Background()
		videoDetails, err := provider.GetVideoDetails(ctx, "12345")
		assert.Error(t, err)
		assert.Nil(t, videoDetails)
		assert.Contains(t, err.Error(), "video no encontrado con el ID")

		clientMock.AssertExpectations(t)
		videosCallMock.AssertExpectations(t)
		loggerMock.AssertExpectations(t)
	})
}
