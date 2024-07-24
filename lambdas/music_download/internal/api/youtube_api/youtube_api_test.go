package youtube_api

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/youtube/v3"
	"testing"
	"time"
)

func TestYouTubeFetcher_LookupSongs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockYouTubeService := new(MockYouTubeService)

		fetcher := NewYoutubeFetcher(mockLogger, mockYouTubeService)

		ctx := context.Background()
		videoID := "testVideoID"
		videoURL := "https://www.youtube.com/watch?v=testVideoID"

		videoDetails := &youtube.Video{
			Snippet: &youtube.VideoSnippet{
				Title:                "Test Video",
				Thumbnails:           &youtube.ThumbnailDetails{Default: &youtube.Thumbnail{Url: "https://example.com/thumbnail.jpg"}},
				LiveBroadcastContent: "none",
			},
			ContentDetails: &youtube.VideoContentDetails{
				Duration: "PT3M21S",
			},
		}

		mockYouTubeService.On("GetVideoDetails", ctx, videoID).Return(videoDetails, nil)

		songs, err := fetcher.LookupSongs(ctx, videoID)

		assert.NoError(t, err)
		assert.Len(t, songs, 1)
		assert.Equal(t, "Test Video", songs[0].Title)
		assert.Equal(t, videoURL, songs[0].URL)
		assert.True(t, songs[0].Playable)
		assert.Equal(t, "https://example.com/thumbnail.jpg", *songs[0].ThumbnailURL)
		assert.Equal(t, 3*time.Minute+21*time.Second, songs[0].Duration)

		mockYouTubeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Invalid Duration", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockYouTubeService := new(MockYouTubeService)

		fetcher := NewYoutubeFetcher(mockLogger, mockYouTubeService)

		ctx := context.Background()
		videoID := "testVideoID"

		videoDetails := &youtube.Video{
			Snippet: &youtube.VideoSnippet{
				Title:                "Test Video",
				Thumbnails:           &youtube.ThumbnailDetails{Default: &youtube.Thumbnail{Url: "https://example.com/thumbnail.jpg"}},
				LiveBroadcastContent: "none",
			},
			ContentDetails: &youtube.VideoContentDetails{
				Duration: "PTInvalidDuration",
			},
		}

		mockYouTubeService.On("GetVideoDetails", ctx, videoID).Return(videoDetails, nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		songs, err := fetcher.LookupSongs(ctx, videoID)

		assert.Error(t, err)
		assert.Nil(t, songs)
		assert.Contains(t, err.Error(), "error al analizar la duracion")

		mockYouTubeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestYouTubeFetcher_SearchYouTubeVideoID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockYouTubeService := new(MockYouTubeService)

		fetcher := NewYoutubeFetcher(mockLogger, mockYouTubeService)

		ctx := context.Background()
		searchTerm := "test song"
		expectedVideoID := "testVideoID"

		mockYouTubeService.On("SearchVideoID", ctx, searchTerm).Return(expectedVideoID, nil)

		videoID, err := fetcher.SearchYouTubeVideoID(ctx, searchTerm)

		assert.NoError(t, err)
		assert.Equal(t, expectedVideoID, videoID)

		mockYouTubeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Search Error", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockYouTubeService := new(MockYouTubeService)

		fetcher := NewYoutubeFetcher(mockLogger, mockYouTubeService)

		ctx := context.Background()
		searchTerm := "test song"

		mockYouTubeService.On("SearchVideoID", ctx, searchTerm).Return("", errors.New("search error"))
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		videoID, err := fetcher.SearchYouTubeVideoID(ctx, searchTerm)

		assert.Error(t, err)
		assert.Empty(t, videoID)
		assert.Contains(t, err.Error(), "error al buscar el video en YouTube")

		mockYouTubeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
