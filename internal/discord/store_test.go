package discord

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"testing"
	"time"
)

func TestInMemoryStorage_SaveSongList(t *testing.T) {
	storage := NewInMemoryStorage()
	channelID := "test_channel"
	songs := []*voice.Song{
		{
			Type:          "audio",
			Title:         "Test Song 1",
			URL:           "https://example.com/test1.mp3",
			Playable:      true,
			Duration:      180 * time.Second,
			StartPosition: 0 * time.Second,
			RequestedBy:   nil,
		},
		{
			Type:          "audio",
			Title:         "Test Song 2",
			URL:           "https://example.com/test2.mp3",
			Playable:      true,
			Duration:      240 * time.Second,
			StartPosition: 0 * time.Second,
			RequestedBy:   nil,
		},
	}

	storage.SaveSongList(channelID, songs)

	savedSongs := storage.GetSongList(channelID)
	if len(savedSongs) != len(songs) {
		t.Errorf("Expected %d songs, but got %d", len(songs), len(savedSongs))
	}
	for i := range songs {
		if savedSongs[i].Title != songs[i].Title {
			t.Errorf("Expected title %s, but got %s", songs[i].Title, savedSongs[i].Title)
		}
		if savedSongs[i].URL != songs[i].URL {
			t.Errorf("Expected URL %s, but got %s", songs[i].URL, savedSongs[i].URL)
		}
	}
}

func TestInMemoryStorage_DeleteSongList(t *testing.T) {
	storage := NewInMemoryStorage()
	channelID := "test_channel"
	songs := []*voice.Song{
		{
			Type:          "audio",
			Title:         "Test Song 1",
			URL:           "https://example.com/test1.mp3",
			Playable:      true,
			Duration:      180 * time.Second,
			StartPosition: 0 * time.Second,
			RequestedBy:   nil,
		},
	}

	storage.SaveSongList(channelID, songs)
	storage.DeleteSongList(channelID)

	savedSongs := storage.GetSongList(channelID)
	if len(savedSongs) != 0 {
		t.Errorf("Expected no songs, but got %d", len(savedSongs))
	}
}

func TestInMemoryStorage_GetSongList(t *testing.T) {
	storage := NewInMemoryStorage()
	channelID := "test_channel"
	songs := []*voice.Song{
		{
			Type:          "audio",
			Title:         "Test Song 1",
			URL:           "https://example.com/test1.mp3",
			Playable:      true,
			Duration:      180 * time.Second,
			StartPosition: 0 * time.Second,
			RequestedBy:   nil,
		},
		{
			Type:          "audio",
			Title:         "Test Song 2",
			URL:           "https://example.com/test2.mp3",
			Playable:      true,
			Duration:      240 * time.Second,
			StartPosition: 0 * time.Second,
			RequestedBy:   nil,
		},
	}

	storage.SaveSongList(channelID, songs)
	savedSongs := storage.GetSongList(channelID)

	if len(savedSongs) != len(songs) {
		t.Errorf("Expected %d songs, but got %d", len(songs), len(savedSongs))
	}
	for i := range songs {
		if savedSongs[i].Title != songs[i].Title {
			t.Errorf("Expected title %s, but got %s", songs[i].Title, savedSongs[i].Title)
		}
		if savedSongs[i].URL != songs[i].URL {
			t.Errorf("Expected URL %s, but got %s", songs[i].URL, savedSongs[i].URL)
		}
	}
}
