package api

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/utils"
	"github.com/pkg/errors"
)

func CheckKafka(cfgApplication *config.Config) (*KafkaMetadata, error) {
	var tlsConfig *tls.Config
	var err error

	if cfgApplication.Messaging.Kafka.EnableTLS {
		tlsConfig, err = utils.NewTLSConfig(&utils.TLSConfig{
			CaFile:   cfgApplication.Messaging.Kafka.CaFile,
			CertFile: cfgApplication.Messaging.Kafka.CertFile,
			KeyFile:  cfgApplication.Messaging.Kafka.KeyFile,
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error configurando conexion de TLS de Kafka")
		}
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version = sarama.V3_5_1_0

	kafkaConfig.Net.DialTimeout = 30 * time.Second
	kafkaConfig.Net.ReadTimeout = 30 * time.Second
	kafkaConfig.Net.WriteTimeout = 30 * time.Second

	if cfgApplication.Messaging.Kafka.EnableTLS {
		kafkaConfig.Net.TLS.Enable = true
		kafkaConfig.Net.TLS.Config = tlsConfig
	} else {
		kafkaConfig.Net.TLS.Enable = false
	}

	client, err := sarama.NewClient(cfgApplication.Messaging.Kafka.Brokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("error conectando a kafka: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("Error cerrando cliente Kafka: %v\n", err)
		}
	}()

	metadata := &KafkaMetadata{}

	for _, broker := range client.Brokers() {
		if err := broker.Open(client.Config()); err != nil && !errors.Is(err, sarama.ErrAlreadyConnected) {
			continue
		}

		request := &sarama.MetadataRequest{}
		response, err := broker.GetMetadata(request)
		if err != nil {
			continue
		}

		brokerMetadata := BrokersMetadata{
			Address:  broker.Addr(),
			IsLeader: response.ControllerID == broker.ID(),
		}
		metadata.Brokers = append(metadata.Brokers, brokerMetadata)
	}
	return metadata, nil
}
