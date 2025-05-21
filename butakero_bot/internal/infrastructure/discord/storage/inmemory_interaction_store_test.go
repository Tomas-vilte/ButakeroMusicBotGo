//go:build !integration

package storage

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestInMemoryStorage_SaveSongList(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	storage := NewInMemoryInteractionStorage(mockLogger)
	channelID := "test_channel"
	songs := []*entity.DiscordEntity{
		{
			TitleTrack: "Test Song 1",
			URL:        "https://example.com/test1.mp3",
			DurationMs: 2324,
		},
		{
			TitleTrack: "Test Song 2",
			URL:        "https://example.com/test2.mp3",
			DurationMs: 12345,
		},
	}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	storage.SaveSongList(channelID, songs)

	savedSongs := storage.GetSongList(channelID)
	if len(savedSongs) != len(songs) {
		t.Errorf("Expected %d songs, but got %d", len(songs), len(savedSongs))
	}
	for i := range songs {
		if savedSongs[i].TitleTrack != songs[i].TitleTrack {
			t.Errorf("Expected title %s, but got %s", songs[i].TitleTrack, savedSongs[i].TitleTrack)
		}
		if savedSongs[i].URL != songs[i].URL {
			t.Errorf("Expected URL %s, but got %s", songs[i].URL, savedSongs[i].URL)
		}
	}
}

func TestInMemoryStorage_DeleteSongList(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	storage := NewInMemoryInteractionStorage(mockLogger)
	channelID := "test_channel"
	songs := []*entity.DiscordEntity{
		{
			TitleTrack: "Test Song 1",
			URL:        "https://example.com/test1.mp3",
			DurationMs: 124456,
		},
	}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	storage.SaveSongList(channelID, songs)
	storage.DeleteSongList(channelID)

	savedSongs := storage.GetSongList(channelID)
	if len(savedSongs) != 0 {
		t.Errorf("Expected no songs, but got %d", len(savedSongs))
	}
}

func TestInMemoryStorage_GetSongList(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	storage := NewInMemoryInteractionStorage(mockLogger)
	channelID := "test_channel"
	songs := []*entity.DiscordEntity{
		{
			TitleTrack: "Test Song 1",
			URL:        "https://example.com/test1.mp3",
			DurationMs: 2456,
		},
		{
			TitleTrack: "Test Song 2",
			URL:        "https://example.com/test2.mp3",
			DurationMs: 12345,
		},
	}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	storage.SaveSongList(channelID, songs)
	savedSongs := storage.GetSongList(channelID)

	if len(savedSongs) != len(songs) {
		t.Errorf("Expected %d songs, but got %d", len(songs), len(savedSongs))
	}
	for i := range songs {
		if savedSongs[i].TitleTrack != songs[i].TitleTrack {
			t.Errorf("Expected title %s, but got %s", songs[i].TitleTrack, savedSongs[i].TitleTrack)
		}
		if savedSongs[i].URL != songs[i].URL {
			t.Errorf("Expected URL %s, but got %s", songs[i].URL, savedSongs[i].URL)
		}
	}
}
