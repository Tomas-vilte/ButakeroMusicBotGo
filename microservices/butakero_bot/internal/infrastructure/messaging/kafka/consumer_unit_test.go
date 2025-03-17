//go:build !integration

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
		messageChan: make(chan *entity.MessageQueue, 1),
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	err := consumer.Close()

	// Assert
	require.NoError(t, err, "El método Close() no debe retornar error")
}

func TestHandleSuccessMessage(t *testing.T) {
	// Arrange
	successJSON := `{
        "video_id": "19f6c66f-26f3-4ccf-bfc7-967449a95ad4",
        "status": "success",
        "message": "Procesamiento exitoso",
        "platform_metadata": {
            "title": "The Emptiness Machine (Official Music Video) - Linkin Park",
            "duration_ms": 2345,
            "url": "https://youtube.com/watch?v=SRXH9AbT280",
            "thumbnail_url": "https://i.ytimg.com/vi/SRXH9AbT280/default.jpg",
            "platform": "youtube"
        },
        "file_data": {
            "file_path": "audio/The Emptiness Machine (Official Music Video) - Linkin Park.dca",
            "file_size": "1.44MB",
            "file_type": "audio/dca"
        },
        "success": true
    }`
	mockLogger := new(logging.MockLogger)
	consumer := &KafkaConsumer{
		messageChan: make(chan *entity.MessageQueue, 1),
		logger:      mockLogger,
	}

	testMsg := &sarama.ConsumerMessage{
		Offset: 1,
		Value:  []byte(successJSON),
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// Act
	consumer.handleMessage(testMsg)

	// Assert
	select {
	case msg := <-consumer.GetMessagesChannel():
		assert.Equal(t, "success", msg.Status, "El estado debe ser success")
		assert.True(t, msg.Success, "El campo success debe ser true")
	case <-time.After(1 * time.Second):
		t.Fatal("No se recibió el mensaje esperado en el canal")
	}
}

func TestHandleErrorMessage(t *testing.T) {
	// Arrange
	errorJSON := `{
        "video_id": "DFswyIQyrl8",
        "status": "error",
        "message": "Error en descarga: io: read/write on closed pipe",
        "platform_metadata": {
            "title": "Ke Personajes - My Immortal (Video Oficial)",
            "duration_ms": 259000,
            "url": "https://youtube.com/watch?v=DFswyIQyrl8",
            "thumbnail_url": "https://i.ytimg.com/vi/DFswyIQyrl8/default.jpg",
            "platform": "youtube"
        },
		"file_data": null,
        "success": false
    }`
	mockLogger := new(logging.MockLogger)
	consumer := &KafkaConsumer{
		messageChan: make(chan *entity.MessageQueue, 1),
		logger:      mockLogger,
	}

	testMsg := &sarama.ConsumerMessage{
		Offset: 2,
		Value:  []byte(errorJSON),
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
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
