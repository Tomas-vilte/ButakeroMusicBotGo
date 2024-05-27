package discordmessenger

import (
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
)

// MockMessageSender es una implementación de MessageSenderWrapper para pruebas.
type MockMessageSender struct {
	mock.Mock
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zap.Field) {
	m.Called(fields)
}

func (m *MockMessageSender) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	args := m.Called(channelID, data, options)
	return args.Get(0).(*discordgo.Message), args.Error(1)
}

func (m *MockMessageSender) ChannelMessageEditComplex(message *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	args := m.Called(m, options)
	return args.Get(0).(*discordgo.Message), args.Error(1)
}

func TestSendMessage(t *testing.T) {
	mockSender := new(MockMessageSender)
	mockLogger := new(MockLogger)
	messageSender := NewMessageSenderImpl(mockSender, mockLogger)
	channelID := "123"
	message := "Test Message"
	mockSender.On("ChannelMessageSendComplex", channelID, mock.Anything, mock.Anything).Return(&discordgo.Message{}, nil)
	mockLogger.On("Info", "Enviando mensaje al canal", mock.AnythingOfType("[]zapcore.Field")).Return()

	err := messageSender.SendMessage(channelID, message)

	assert.NoError(t, err)
	mockSender.AssertCalled(t, "ChannelMessageSendComplex", channelID, mock.Anything, mock.Anything)

}

func TestSendMessage_Error(t *testing.T) {
	// Configuración
	mockSender := new(MockMessageSender)
	mockLogger := new(MockLogger)
	messageSender := NewMessageSenderImpl(mockSender, mockLogger)
	channelID := "123"
	message := "Test Message"
	expectedError := errors.New("error sending message")
	mockMessage := &discordgo.Message{} // Creamos un mensaje simulado
	mockSender.On("ChannelMessageSendComplex", channelID, mock.Anything, mock.Anything).Return(mockMessage, expectedError)
	mockLogger.On("Info", "Enviando mensaje al canal", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Error", "Error al enviar el mensaje: ", mock.AnythingOfType("[]zapcore.Field")).Return()

	// Ejecución
	err := messageSender.SendMessage(channelID, message)

	// Verificación
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockSender.AssertCalled(t, "ChannelMessageSendComplex", channelID, mock.Anything, mock.Anything)
}

func TestSendPlayMessage(t *testing.T) {
	// Configuración
	mockSender := new(MockMessageSender)
	mockLogger := new(MockLogger)
	messageSender := NewMessageSenderImpl(mockSender, mockLogger)
	channelID := "123"
	mockMessage := &voice.PlayMessage{}
	mockSender.On("ChannelMessageSendComplex", channelID, mock.Anything, mock.Anything).Return(&discordgo.Message{}, nil)
	mockLogger.On("Info", "Enviando mensaje de reproducción...", mock.AnythingOfType("[]zapcore.Field")).Return()

	// Ejecución
	_, err := messageSender.SendPlayMessage(channelID, mockMessage)

	// Verificación
	assert.NoError(t, err)
	mockSender.AssertCalled(t, "ChannelMessageSendComplex", channelID, mock.Anything, mock.Anything)
}

func TestSendPlayMessage_Error(t *testing.T) {
	// Configuración
	mockSender := new(MockMessageSender)
	mockLogger := new(MockLogger)
	messageSender := NewMessageSenderImpl(mockSender, mockLogger)
	channelID := "123"
	mockMessage := &voice.PlayMessage{}
	expectedErr := errors.New("error al enviar mensaje de reproducción")

	// Configura el mock para devolver un error al enviar el mensaje de reproducción
	mockSender.On("ChannelMessageSendComplex", channelID, mock.Anything, mock.Anything).Return(&discordgo.Message{}, expectedErr)
	mockLogger.On("Info", "Enviando mensaje de reproducción...", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Error", "Error al enviar mensaje de reproducción: ", mock.AnythingOfType("[]zapcore.Field")).Return()

	// Ejecución
	_, err := messageSender.SendPlayMessage(channelID, mockMessage)

	// Verificación
	assert.Error(t, err)
	assert.EqualError(t, err, expectedErr.Error())
	mockSender.AssertCalled(t, "ChannelMessageSendComplex", channelID, mock.Anything, mock.Anything)
}

func TestEditPlayMessage(t *testing.T) {
	// Configuración
	mockSender := new(MockMessageSender)
	mockLogger := new(MockLogger)
	messageSender := NewMessageSenderImpl(mockSender, mockLogger)
	channelID := "123"
	messageID := "456"
	mockMessage := &voice.PlayMessage{}
	mockSender.On("ChannelMessageEditComplex", mock.Anything, mock.Anything).Return(&discordgo.Message{}, nil)

	// Ejecución
	err := messageSender.EditPlayMessage(channelID, messageID, mockMessage)

	// Verificación
	assert.NoError(t, err)
	mockSender.AssertCalled(t, "ChannelMessageEditComplex", mock.Anything, mock.Anything)
}

func TestEditPlayMessage_Error(t *testing.T) {
	// Configuración
	mockSender := new(MockMessageSender)
	mockLogger := new(MockLogger)
	messageSender := NewMessageSenderImpl(mockSender, mockLogger)
	channelID := "123"
	messageID := "456"
	mockMessage := &voice.PlayMessage{}
	expectedErr := errors.New("error al editar mensaje de reproducción")

	// Configura el mock para que devuelva un error al llamar a ChannelMessageEditComplex
	mockSender.On("ChannelMessageEditComplex", mock.Anything, mock.Anything, mock.Anything).Return(&discordgo.Message{}, expectedErr)
	mockLogger.On("Error", "Error al editar el mensaje de reproducción: ", mock.AnythingOfType("[]zapcore.Field")).Return()

	// Ejecución
	err := messageSender.EditPlayMessage(channelID, messageID, mockMessage)

	// Verificación
	assert.Error(t, err)
	assert.EqualError(t, err, expectedErr.Error())
	mockSender.AssertCalled(t, "ChannelMessageEditComplex", mock.Anything, mock.Anything, mock.Anything)
}
