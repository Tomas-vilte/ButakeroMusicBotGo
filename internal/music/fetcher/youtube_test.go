package fetcher

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"io"
	"reflect"
	"testing"
)

func TestYoutubeFetcher_LookupSongs(t *testing.T) {
	// Arrange
	ctx := context.Background()
	f := NewYoutubeFetcher()
	query := "hello"

	// Act
	songs, err := f.LookupSongs(ctx, query)

	// Assert
	if err != nil {
		t.Errorf("LookupSongs returned an unexpected error: %v", err)
	}
	if len(songs) == 0 {
		t.Error("LookupSongs did not return any songs")
	}
	for _, song := range songs {
		if song.Title == "" {
			t.Error("LookupSongs returned a song with an empty title")
		}
		if song.URL == "" {
			t.Error("LookupSongs returned a song with an empty URL")
		}
	}
}

func TestYoutubeFetcher_GetDCAData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	f := NewYoutubeFetcher()
	song := &voice.Song{
		Type: "yt-dlp",
		URL:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	}

	// Act
	reader, err := f.GetDCAData(ctx, song)

	// Assert
	if err != nil {
		t.Errorf("GetDCAData returned an unexpected error: %v", err)
	}
	data := make([]byte, 1024)
	n, err := reader.Read(data)
	if err != nil && err != io.EOF {
		t.Errorf("Error reading from DCA data reader: %v", err)
	}
	if n == 0 {
		t.Error("GetDCAData did not return any DCA data")
	}
}

func TestGetThumbnail(t *testing.T) {
	tests := []struct {
		name          string
		thumbnailsStr string
		want          *thumnail
		wantErr       bool
	}{
		{
			name:          "valid thumbnails",
			thumbnailsStr: `[{"url":"https://example.com/thumbnail1.jpg","preference":1},{"url":"https://example.com/thumbnail2.jpg","preference":2}]`,
			want:          &thumnail{URL: "https://example.com/thumbnail2.jpg", Preference: 2},
		},
		{
			name:          "invalid JSON",
			thumbnailsStr: `invalid JSON`,
			wantErr:       true,
		},
		{
			name:          "empty thumbnails",
			thumbnailsStr: `[]`,
			want:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getThumbnail(tt.thumbnailsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetThumbnail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetThumbnail() = %v, want %v", got, tt.want)
			}
		})
	}
}
