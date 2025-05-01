//go:build !integration

package discord

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewGuildManager(t *testing.T) {
	// Arrange
	mockFactory := new(MockPlayerFactory)
	mockLogger := new(logging.MockLogger)

	// Act
	manager := NewGuildManager(mockFactory, mockLogger)

	// Assert
	assert.NotNil(t, manager)
	assert.IsType(t, &GuildManager{}, manager)
}

func TestCreateGuildPlayer_Success(t *testing.T) {
	// Arrange
	mockFactory := new(MockPlayerFactory)
	mockLogger := new(logging.MockLogger)
	manager := NewGuildManager(mockFactory, mockLogger)

	guildID := "test-guild"
	mockPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockFactory.On("CreatePlayer", guildID).Return(mockPlayer, nil)

	// Act
	player, err := manager.CreateGuildPlayer(guildID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, player)
	mockFactory.AssertExpectations(t)
	mockPlayer.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCreateGuildPlayer_EmptyGuildID(t *testing.T) {
	// Arrange
	mockFactory := new(MockPlayerFactory)
	mockLogger := new(logging.MockLogger)
	manager := NewGuildManager(mockFactory, mockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	// Act
	player, err := manager.CreateGuildPlayer("")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, player)
	assert.Equal(t, errors_app.ErrCodeInvalidGuildID, err.(*errors_app.AppError).Code)
}

func TestGuildManager_CreateGuildPlayer_AlreadyExists(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	mockGuildPlayer := new(MockGuildPlayer)
	mockPlayerFactory := new(MockPlayerFactory)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	mockPlayerFactory.On("CreatePlayer", "guild123").Return(mockGuildPlayer, nil)

	guildManager := NewGuildManager(mockPlayerFactory, mockLogger)

	firstPlayer, err := guildManager.CreateGuildPlayer("guild123")
	assert.NoError(t, err)
	assert.NotNil(t, firstPlayer)

	mockPlayerFactory.Calls = nil

	secondPlayer, err := guildManager.CreateGuildPlayer("guild123")

	assert.Error(t, err)
	assert.True(t, errors_app.IsAppError(err), "El error debería ser un AppError")
	assert.Equal(t, errors_app.ErrCodeGuildPlayerAlreadyExists, err.(*errors_app.AppError).Code)
	assert.Equal(t, firstPlayer, secondPlayer)
	mockPlayerFactory.AssertNotCalled(t, "CreatePlayer")
}

func TestGuildManager_CreateGuildPlayer_FactoryError(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	factoryError := errors_app.NewAppError(errors_app.ErrCodeInternalError, "Factory error", nil)
	mockPlayerFactory := &MockPlayerFactory{}
	mockPlayerFactory.On("CreatePlayer", "guild123").Return((*MockGuildPlayer)(nil), factoryError)

	guildManager := NewGuildManager(mockPlayerFactory, mockLogger)

	// Act
	player, err := guildManager.CreateGuildPlayer("guild123")

	// Assert
	assert.Nil(t, player)
	assert.Error(t, err)
	assert.True(t, errors_app.IsAppError(err), "El error debería ser un AppError")
	assert.Equal(t, errors_app.ErrCodeGuildPlayerCreateFailed, err.(*errors_app.AppError).Code)
	mockPlayerFactory.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestGuildManager_RemoveGuildPlayer(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	mockGuildPlayer := new(MockGuildPlayer)
	mockPlayerFactory := new(MockPlayerFactory)
	mockPlayerFactory.On("CreatePlayer", "guild123").Return(mockGuildPlayer, nil)
	mockPlayerFactory.On("CreatePlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("Close").Return(nil)

	guildManager := NewGuildManager(mockPlayerFactory, mockLogger)

	_, err := guildManager.CreateGuildPlayer("guild123")
	assert.NoError(t, err)

	// Act
	err = guildManager.RemoveGuildPlayer("guild123")

	// Assert
	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)

	_, err = guildManager.GetGuildPlayer("guild123")
	assert.NoError(t, err)
	mockPlayerFactory.AssertNumberOfCalls(t, "CreatePlayer", 2)
}

func TestGuildManager_RemoveGuildPlayer_EmptyGuildID(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockPlayerFactory := new(MockPlayerFactory)
	guildManager := NewGuildManager(mockPlayerFactory, mockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	// Act
	err := guildManager.RemoveGuildPlayer("")

	// Assert
	assert.Error(t, err)
	assert.True(t, errors_app.IsAppError(err), "El error debería ser un AppError")
	assert.Equal(t, errors_app.ErrCodeInvalidGuildID, err.(*errors_app.AppError).Code)
}

func TestGuildManager_RemoveGuildPlayer_NotFound(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	mockPlayerFactory := new(MockPlayerFactory)
	guildManager := NewGuildManager(mockPlayerFactory, mockLogger)

	// Act
	err := guildManager.RemoveGuildPlayer("nonexistent")

	// Assert
	assert.Error(t, err)
	assert.True(t, errors_app.IsAppError(err), "El error debería ser un AppError")
	assert.Equal(t, errors_app.ErrCodeGuildPlayerNotFound, err.(*errors_app.AppError).Code)
	mockLogger.AssertExpectations(t)
}

func TestGuildManager_GetGuildPlayer_Existing(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	mockGuildPlayer := new(MockGuildPlayer)
	mockPlayerFactory := new(MockPlayerFactory)
	mockPlayerFactory.On("CreatePlayer", "guild123").Return(mockGuildPlayer, nil)

	guildManager := NewGuildManager(mockPlayerFactory, mockLogger)

	_, err := guildManager.CreateGuildPlayer("guild123")
	assert.NoError(t, err)

	// Act
	player, err := guildManager.GetGuildPlayer("guild123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, mockGuildPlayer, player)
	mockPlayerFactory.AssertNumberOfCalls(t, "CreatePlayer", 1)
}

func TestGuildManager_GetGuildPlayer_NonExisting(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	mockGuildPlayer := new(MockGuildPlayer)
	mockPlayerFactory := new(MockPlayerFactory)
	mockPlayerFactory.On("CreatePlayer", "guild123").Return(mockGuildPlayer, nil)

	guildManager := NewGuildManager(mockPlayerFactory, mockLogger)

	// Act
	player, err := guildManager.GetGuildPlayer("guild123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, mockGuildPlayer, player)
	mockPlayerFactory.AssertCalled(t, "CreatePlayer", "guild123")
}

func TestGuildManager_GetGuildPlayer_EmptyGuildID(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockPlayerFactory := new(MockPlayerFactory)
	guildManager := NewGuildManager(mockPlayerFactory, mockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	// Act
	player, err := guildManager.GetGuildPlayer("")

	// Assert
	assert.Nil(t, player)
	assert.Error(t, err)
	assert.True(t, errors_app.IsAppError(err), "El error debería ser un AppError")
	assert.Equal(t, errors_app.ErrCodeInvalidGuildID, err.(*errors_app.AppError).Code)
	mockPlayerFactory.AssertNotCalled(t, "CreatePlayer")
}
