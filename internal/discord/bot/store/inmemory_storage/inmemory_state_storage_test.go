package inmemory_storage

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestInmemoryStateStorage_GetCurrentSong verifica que el método GetCurrentSong devuelva la canción actual correctamente.
func TestInmemoryStateStorage_GetCurrentSong(t *testing.T) {
	mockLogger := &MockLogger{}
	storage := NewInmemoryStateStorage(mockLogger)

	currentSong := &voice.PlayedSong{Song: voice.Song{Title: "Test Song"}}
	err := storage.SetCurrentSong(currentSong)
	if err != nil {
		t.Errorf("Error al establecer la canción actual: %v", err)
	}

	song, err := storage.GetCurrentSong()
	if err != nil {
		t.Errorf("Error al obtener la canción actual: %v", err)
	}

	if song.Title != currentSong.Title {
		t.Errorf("La canción obtenida no coincide con la canción establecida. Esperado: %s, Obtenido: %s", currentSong.Title, song.Title)
	}

	mockLogger.AssertExpectations(t)
}

// TestInmemoryStateStorage_GetVoiceChannel verifica que el método GetVoiceChannel devuelva el canal de voz correctamente.
func TestInmemoryStateStorage_GetVoiceChannel(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Info", "Obteniendo el canal de voz", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Canal de voz establecido", mock.AnythingOfType("[]zapcore.Field")).Return()

	storage := NewInmemoryStateStorage(mockLogger)

	// Set voice channel
	voiceChannelID := "123456789"
	err := storage.SetVoiceChannel(voiceChannelID)
	if err != nil {
		t.Errorf("Error al establecer el canal de voz: %v", err)
	}

	// Get voice channel
	channelID, err := storage.GetVoiceChannel()
	if err != nil {
		t.Errorf("Error al obtener el canal de voz: %v", err)
	}

	// Check if the retrieved voice channel ID matches the set voice channel ID
	if channelID != voiceChannelID {
		t.Errorf("El canal de voz obtenido no coincide con el canal de voz establecido. Esperado: %s, Obtenido: %s", voiceChannelID, channelID)
	}

	mockLogger.AssertExpectations(t)
}

// TestInmemoryStateStorage_GetTextChannel verifica que el método GetTextChannel devuelva el canal de texto correctamente.
func TestInmemoryStateStorage_GetTextChannel(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Info", "Obteniendo el canal de texto", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Canal de texto establecido", mock.AnythingOfType("[]zapcore.Field")).Return()

	storage := NewInmemoryStateStorage(mockLogger)

	// Set text channel
	textChannelID := "987654321"
	err := storage.SetTextChannel(textChannelID)
	if err != nil {
		t.Errorf("Error al establecer el canal de texto: %v", err)
	}

	// Get text channel
	channelID, err := storage.GetTextChannel()
	if err != nil {
		t.Errorf("Error al obtener el canal de texto: %v", err)
	}

	// Check if the retrieved text channel ID matches the set text channel ID
	if channelID != textChannelID {
		t.Errorf("El canal de texto obtenido no coincide con el canal de texto establecido. Esperado: %s, Obtenido: %s", textChannelID, channelID)
	}

	mockLogger.AssertExpectations(t)
}
