package kafka

import (
	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewTLSConfig_Failure(t *testing.T) {
	certFile := "nonexistent_cert.pem"
	keyFile := "nonexistent_key.pem"
	caCertFile := "nonexistent_ca.pem"

	// Act
	tlsConfig, err := NewTLSConfig(certFile, keyFile, caCertFile)

	// Assert
	require.Error(t, err, "Se esperaba error al cargar archivos inexistentes")
	assert.Nil(t, tlsConfig, "El TLS config debe ser nil en caso de error")
}

func TestKafkaConsumer_Close(t *testing.T) {
	config := sarama.NewConfig()
	mockConsumer := mocks.NewConsumer(t, config)
	mockLogger := new(logging.MockLogger)
	consumer := &KafkaConsumer{
		consumer:    mockConsumer,
		brokers:     []string{"dummy"},
		topic:       "test-topic",
		logger:      mockLogger,
		messageChan: make(chan *entity.StatusMessage, 1),
	}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	err := consumer.Close()

	// Assert
	require.NoError(t, err, "El método Close() no debe retornar error")
}

func TestHandleSuccessMessage(t *testing.T) {
	// Arrange
	successJSON := `{
		"status": {
			"id": "19f6c66f-26f3-4ccf-bfc7-967449a95ad4",
			"sk": "SRXH9AbT280",
			"status": "success",
			"message": "Procesamiento exitoso",
			"metadata": {
				"id": "63f48016-78cd-4387-99b9-c38af46e8e90",
				"video_id": "SRXH9AbT280",
				"title": "The Emptiness Machine (Official Music Video) - Linkin Park",
				"duration": "PT3M21S",
				"url_youtube": "https://youtube.com/watch?v=SRXH9AbT280",
				"thumbnail": "https://i.ytimg.com/vi/SRXH9AbT280/default.jpg",
				"platform": "Youtube"
			},
			"file_data": {
				"file_path": "audio/The Emptiness Machine (Official Music Video) - Linkin Park.dca",
				"file_size": "1.44MB",
				"file_type": "audio/dca",
				"public_url": "file://data/audio-files/audio/The Emptiness Machine (Official Music Video) - Linkin Park.dca"
			},
			"processing_date": "2024-12-24T05:39:58Z",
			"success": true,
			"attempts": 1,
			"failures": 0
		}
	}`
	mockLogger := new(logging.MockLogger)
	consumer := &KafkaConsumer{
		messageChan: make(chan *entity.StatusMessage, 1),
		logger:      mockLogger,
	}

	testMsg := &sarama.ConsumerMessage{
		Offset: 1,
		Value:  []byte(successJSON),
	}

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// Act
	consumer.handleMessage(testMsg)

	// Assert
	select {
	case msg := <-consumer.GetMessagesChannel():
		assert.Equal(t, "success", msg.Status.Status, "El estado debe ser success")
		assert.True(t, msg.Status.Success, "El campo success debe ser true")
	case <-time.After(1 * time.Second):
		t.Fatal("No se recibió el mensaje esperado en el canal")
	}
}

func TestHandleErrorMessage(t *testing.T) {
	// Arrange
	errorJSON := `{
		"status": {
			"id": "959326a1-53db-4810-9fc8-b17275122158",
			"sk": "DFswyIQyrl8",
			"status": "error",
			"message": "Error en descarga: io: read/write on closed pipe",
			"metadata": {
				"id": "873f0521-f808-4721-b2c1-5e63a782b7cf",
				"video_id": "DFswyIQyrl8",
				"title": "Ke Personajes - My Immortal (Video Oficial)",
				"duration": "PT4M19S",
				"url_youtube": "https://youtube.com/watch?v=DFswyIQyrl8",
				"thumbnail": "https://i.ytimg.com/vi/DFswyIQyrl8/default.jpg",
				"platform": "Youtube"
			},
			"file_data": null,
			"processing_date": "2025-02-10T14:23:14Z",
			"success": false,
			"attempts": 8,
			"failures": 8
		}
	}`
	mockLogger := new(logging.MockLogger)
	consumer := &KafkaConsumer{
		messageChan: make(chan *entity.StatusMessage, 1),
		logger:      mockLogger,
	}

	testMsg := &sarama.ConsumerMessage{
		Offset: 2,
		Value:  []byte(errorJSON),
	}

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	// Act
	consumer.handleMessage(testMsg)

	select {
	case msg := <-consumer.GetMessagesChannel():
		t.Fatalf("No se esperaba recibir mensaje, pero se obtuvo: %+v", msg)
	case <-time.After(500 * time.Millisecond):
	}
}
