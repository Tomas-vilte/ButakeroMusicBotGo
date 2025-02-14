package kafka

type KafkaConfig struct {
	Brokers    []string
	Topic      string
	TLS        bool
	CertFile   string
	KeyFile    string
	CACertFile string
}
