package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type TLSConfig struct {
	CertFile string
	CaFile   string
	KeyFile  string
}

func NewTLSConfig(params *TLSConfig) (*tls.Config, error) {
	caCert, err := os.ReadFile(params.CaFile)
	if err != nil {
		return nil, errors.Wrap(err, "Error al leer el archivo de certificado CA")
	}

	cert, err := tls.LoadX509KeyPair(params.CertFile, params.KeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "Error al cargar el certificado y la clave")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}

func BuildMongoURI(cfg *config.Config) string {
	hostList := strings.Join(cfg.Database.Mongo.Host, ",")
	return fmt.Sprintf("mongodb://%s:%s@%s/?replicaSet=%s&tls=%v",
		cfg.Database.Mongo.User,
		cfg.Database.Mongo.Password,
		hostList,
		cfg.Database.Mongo.ReplicaSetName,
		cfg.Database.Mongo.EnableTLS,
	)
}
