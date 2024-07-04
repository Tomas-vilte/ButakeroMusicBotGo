package queuing

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestProcessSQSEvent_ReleaseEvent(t *testing.T) {
	mockDiscordClient := new(MockDiscordGoClient)
	mockLogger := new(MockLogger)
	consumer := NewSQSConsumer(mockDiscordClient, mockLogger)

	event := map[string]interface{}{
		"action": "published",
		"release": map[string]interface{}{
			"tag_name": "v1.2.3",
			"body":     "This is the release body",
			"html_url": "https://example.com/release",
		},
	}
	eventBody, _ := json.Marshal(event)

	mockLogger.On("Error", mock.Anything, mock.Anything).Maybe()
	mockDiscordClient.On("SendMessageToServers", mock.Anything).Return(nil)

	err := consumer.ProcessSQSEvent(eventBody)
	assert.NoError(t, err)

	mockDiscordClient.AssertCalled(t, "SendMessageToServers", mock.AnythingOfType("*discordgo.MessageEmbed"))
}

//func TestProcessSQSEvent_WorkflowEvent(t *testing.T) {
//	mockDiscordClient := new(MockDiscordGoClient)
//	mockLogger := new(MockLogger)
//	consumer := NewSQSConsumer(mockDiscordClient, mockLogger)
//
//	event := map[string]interface{}{
//		"action": "completed",
//		"workflow_job": map[string]interface{}{
//			"workflow_name": "CI Pipeline",
//			"conclusion":    "success",
//			"html_url":      "https://example.com/workflow",
//			"completed_at":  "2021-01-01T12:00:00Z",
//		},
//	}
//	eventBody, _ := json.Marshal(event)
//
//	mockLogger.On("Error", mock.Anything, mock.Anything).Maybe()
//	mockDiscordClient.On("SendMessageToServers", mock.Anything).Return(nil)
//
//	err := consumer.ProcessSQSEvent(eventBody)
//	assert.NoError(t, err)
//
//	mockDiscordClient.AssertCalled(t, "SendMessageToServers", mock.AnythingOfType("*discordgo.MessageEmbed"))
//}

func TestProcessSQSEvent_InvalidJSON(t *testing.T) {
	mockDiscordClient := new(MockDiscordGoClient)
	mockLogger := new(MockLogger)
	consumer := NewSQSConsumer(mockDiscordClient, mockLogger)

	invalidJSON := []byte(`{"action":`)
	mockLogger.On("Error", "Error al analizar el cuerpo del mensaje", mock.AnythingOfType("[]zapcore.Field")).Return()
	err := consumer.ProcessSQSEvent(invalidJSON)
	assert.Error(t, err)
	mockLogger.AssertCalled(t, "Error", "Error al analizar el cuerpo del mensaje", mock.AnythingOfType("[]zapcore.Field"))

}

func TestProcessSQSEvent_UnknownAction(t *testing.T) {
	mockDiscordClient := new(MockDiscordGoClient)
	mockLogger := new(MockLogger)
	consumer := NewSQSConsumer(mockDiscordClient, mockLogger)

	event := map[string]interface{}{
		"action": "unknown",
	}
	eventBody, _ := json.Marshal(event)

	mockLogger.On("Error", "Error acción desconocida", mock.AnythingOfType("[]zapcore.Field")).Return()

	err := consumer.ProcessSQSEvent(eventBody)
	assert.Error(t, err)
	mockLogger.AssertCalled(t, "Error", "Error acción desconocida", mock.AnythingOfType("[]zapcore.Field"))

}

func TestProcessSQSEvent_SuccessfulFormattingEvent(t *testing.T) {
	mockDiscordMessenger := new(MockDiscordGoClient)
	mockLogger := new(MockLogger)
	consumer := NewSQSConsumer(mockDiscordMessenger, mockLogger)

	eventBody := []byte(`{"action": "published", "release": {"tag_name": "v1.0.0", "body": "This is a release", "html_url": "https://example.com"}}`)

	mockDiscordMessenger.On("SendMessageToServers", mock.AnythingOfType("*discordgo.MessageEmbed")).Return(nil)

	err := consumer.ProcessSQSEvent(eventBody)
	assert.NoError(t, err)
}

func TestProcessSQSEvent_ErrorSendingMessage(t *testing.T) {
	mockDiscordClient := new(MockDiscordGoClient)
	mockLogger := new(MockLogger)
	consumer := NewSQSConsumer(mockDiscordClient, mockLogger)

	event := map[string]interface{}{
		"action": "published",
		"release": map[string]interface{}{
			"tag_name": "v1.2.3",
			"body":     "This is the release body",
			"html_url": "https://example.com/release",
		},
	}
	eventBody, _ := json.Marshal(event)

	mockLogger.On("Error", "Error al enviar el mensaje a Discord", mock.AnythingOfType("[]zapcore.Field")).Maybe()
	mockDiscordClient.On("SendMessageToServers", mock.Anything).Return(errors.New("send message error"))

	err := consumer.ProcessSQSEvent(eventBody)
	assert.Error(t, err)

	mockLogger.AssertCalled(t, "Error", "Error al enviar el mensaje a Discord", mock.AnythingOfType("[]zapcore.Field"))
}

func TestProcessSQSEvent_NotAction(t *testing.T) {
	mockDiscordClient := new(MockDiscordGoClient)
	mockLogger := new(MockLogger)
	consumer := NewSQSConsumer(mockDiscordClient, mockLogger)

	event := map[string]interface{}{
		"unknown_action": "unknown_action",
	}
	eventBody, _ := json.Marshal(event)

	mockLogger.On("Error", "Error el campo 'action', no encontrado o no es una cadena", mock.AnythingOfType("[]zapcore.Field")).Return()

	err := consumer.ProcessSQSEvent(eventBody)
	assert.NoError(t, err)

	mockLogger.AssertCalled(t, "Error", "Error el campo 'action', no encontrado o no es una cadena", mock.AnythingOfType("[]zapcore.Field"))
}

func TestProcessSQSEvent_ErrorFormattingEvent(t *testing.T) {
	mockDiscordClient := new(MockDiscordGoClient)
	mockLogger := new(MockLogger)
	consumer := NewSQSConsumer(mockDiscordClient, mockLogger)

	event := map[string]interface{}{
		"action": "published",
	}
	eventBody, _ := json.Marshal(event)
	expectedError := errors.New("campo 'release' no encontrado o no es un mapa")
	mockLogger.On("Error", "Error al formatear el evento", mock.AnythingOfType("[]zapcore.Field")).Return()

	err := consumer.ProcessSQSEvent(eventBody)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	mockLogger.AssertCalled(t, "Error", "Error al formatear el evento", mock.AnythingOfType("[]zapcore.Field"))

}
