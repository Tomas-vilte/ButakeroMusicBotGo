//go:build !integration

package inmemory

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestInmemoryStateStorage_GetCurrentSong(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	storage := NewPlayerStateManager(mockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	ctx := context.Background()

	currentSong := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "Test Song"}}
	err := storage.SetCurrentTrack(ctx, currentSong)
	if err != nil {
		t.Errorf("Error al establecer la canción actual: %v", err)
	}

	song, err := storage.GetCurrentTrack(ctx)
	if err != nil {
		t.Errorf("Error al obtener la canción actual: %v", err)
	}

	if song.DiscordSong.TitleTrack != currentSong.DiscordSong.TitleTrack {
		t.Errorf("La canción obtenida no coincide con la canción establecida. Esperado: %s, Obtenido: %s", currentSong.DiscordSong.TitleTrack, song.DiscordSong.TitleTrack)
	}

	mockLogger.AssertExpectations(t)
}

func TestInmemoryStateStorage_GetVoiceChannel(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	storage := NewPlayerStateManager(mockLogger)
	ctx := context.Background()

	voiceChannelID := "123456789"
	err := storage.SetVoiceChannelID(ctx, voiceChannelID)
	if err != nil {
		t.Errorf("Error al establecer el canal de voz: %v", err)
	}

	channelID, err := storage.GetVoiceChannelID(ctx)
	if err != nil {
		t.Errorf("Error al obtener el canal de voz: %v", err)
	}

	if channelID != voiceChannelID {
		t.Errorf("El canal de voz obtenido no coincide con el canal de voz establecido. Esperado: %s, Obtenido: %s", voiceChannelID, channelID)
	}

	mockLogger.AssertExpectations(t)
}

func TestInmemoryStateStorage_GetTextChannel(t *testing.T) {
	mockLogger := new(logging.MockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	storage := NewPlayerStateManager(mockLogger)

	ctx := context.Background()

	textChannelID := "987654321"
	err := storage.SetTextChannelID(ctx, textChannelID)
	if err != nil {
		t.Errorf("Error al establecer el canal de texto: %v", err)
	}

	channelID, err := storage.GetTextChannelID(ctx)
	if err != nil {
		t.Errorf("Error al obtener el canal de texto: %v", err)
	}

	if channelID != textChannelID {
		t.Errorf("El canal de texto obtenido no coincide con el canal de texto establecido. Esperado: %s, Obtenido: %s", textChannelID, channelID)
	}

	mockLogger.AssertExpectations(t)
}
