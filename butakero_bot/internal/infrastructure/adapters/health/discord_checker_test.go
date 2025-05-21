//go:build !integration

package health_test

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters/health"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestDiscordChecker_Check_SessionNil(t *testing.T) {
	// Arrange
	logger := new(logging.MockLogger)

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	checker := health.NewDiscordChecker(nil, logger)

	// Act
	h, err := checker.Check(context.Background())

	// Assert
	assert.False(t, h.Connected)
	assert.Equal(t, "Sesión de Discord no inicializada", h.Error)
	assert.NotNil(t, err)
	logger.AssertExpectations(t)
}

func TestDiscordChecker_Check_Success(t *testing.T) {
	// Arrange
	logger := new(logging.MockLogger)
	session := &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{
				SessionID: "test_session_id",
				User: &discordgo.User{
					ID: "bot_user_id",
				},
				Guilds: []*discordgo.Guild{
					{
						ID: "guild1",
						VoiceStates: []*discordgo.VoiceState{
							{
								UserID:    "bot_user_id",
								ChannelID: "channel1",
							},
							{
								UserID:    "other_user",
								ChannelID: "channel2",
							},
						},
					},
					{
						ID: "guild2",
						VoiceStates: []*discordgo.VoiceState{
							{
								UserID:    "bot_user_id",
								ChannelID: "channel3",
							},
						},
					},
					{
						ID:          "guild3",
						VoiceStates: []*discordgo.VoiceState{},
					},
				},
			},
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session.DataReady = true
	now := time.Now()
	session.LastHeartbeatSent = now.Add(-500 * time.Millisecond)
	session.LastHeartbeatAck = now

	checker := health.NewDiscordChecker(session, logger)

	// Act
	h, err := checker.Check(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, h.Connected)
	assert.InDelta(t, 500.0, h.HeartbeatLatencyMS, 10.0)
	assert.Equal(t, 3, h.Guilds)
	assert.Equal(t, 2, h.VoiceConnections)
	fmt.Println(h.CheckDurationMS)
	assert.Equal(t, "test_session_id", h.SessionID)
	assert.GreaterOrEqual(t, h.CheckDurationMS, 0.0)
	if h.CheckDurationMS == 0 {
		t.Log("CheckDurationMS es 0, puede ser demasiado rápido para medir")
	}
	assert.Empty(t, h.Error)
}

func TestDiscordChecker_Check_NotConnected(t *testing.T) {
	// Arrange
	logger := new(logging.MockLogger)
	session := &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{
				SessionID: "test_session_id",
				User: &discordgo.User{
					ID: "bot_user_id",
				},
				Guilds: []*discordgo.Guild{},
			},
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	session.DataReady = false

	checker := health.NewDiscordChecker(session, logger)

	// Act
	h, err := checker.Check(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.False(t, h.Connected)
	assert.Equal(t, "WebSocket no conectado", h.Error)
	assert.GreaterOrEqual(t, h.CheckDurationMS, 0.0, "CheckDurationMS debería ser mayor o igual que cero")

}

func TestDiscordChecker_Check_HighLatency(t *testing.T) {
	// Arrange
	logger := new(logging.MockLogger)
	session := &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{
				SessionID: "test_session_id",
				User: &discordgo.User{
					ID: "bot_user_id",
				},
				Guilds: []*discordgo.Guild{},
			},
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	session.DataReady = true
	now := time.Now()
	session.LastHeartbeatSent = now.Add(-1500 * time.Millisecond)
	session.LastHeartbeatAck = now

	checker := health.NewDiscordChecker(session, logger)

	// Act
	h, err := checker.Check(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, h.Connected)
	assert.InDelta(t, 1500.0, h.HeartbeatLatencyMS, 10.0)
	assert.Contains(t, h.Error, "Alta latencia")
	assert.GreaterOrEqual(t, h.CheckDurationMS, 0.0, "CheckDurationMS debería ser mayor o igual que cero")

}

func TestDiscordChecker_Check_NoVoiceConnections(t *testing.T) {
	// Arrange
	logger := new(logging.MockLogger)
	session := &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{
				SessionID: "test_session_id",
				User: &discordgo.User{
					ID: "bot_user_id",
				},
				Guilds: []*discordgo.Guild{
					{
						ID: "guild1",
						VoiceStates: []*discordgo.VoiceState{
							{
								UserID:    "other_user1",
								ChannelID: "channel1",
							},
							{
								UserID:    "other_user2",
								ChannelID: "channel2",
							},
						},
					},
					{
						ID:          "guild2",
						VoiceStates: []*discordgo.VoiceState{},
					},
				},
			},
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	session.DataReady = true
	now := time.Now()
	session.LastHeartbeatSent = now.Add(-100 * time.Millisecond)
	session.LastHeartbeatAck = now

	checker := health.NewDiscordChecker(session, logger)

	// Act
	h, err := checker.Check(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, h.Connected)
	assert.InDelta(t, 100.0, h.HeartbeatLatencyMS, 10.0)
	assert.Equal(t, 2, h.Guilds)
	assert.Equal(t, 0, h.VoiceConnections)
	assert.Empty(t, h.Error)
	assert.GreaterOrEqual(t, h.CheckDurationMS, 0.0, "CheckDurationMS debería ser mayor o igual que cero")

}
