package shared

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func StringPtr(s string) *string {
	return &s
}

type TLSConfig struct {
	Enabled  bool
	CAFile   string
	CertFile string
	KeyFile  string
}

func ConfigureTLS(config TLSConfig) (*tls.Config, error) {
	if !config.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if config.CAFile != "" {
		caCert, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("error al leer el archivo CA: %v", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("error al agregar el CA certificate al pool")
		}

		tlsConfig.RootCAs = caCertPool
	}

	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("error al cargar el client certificate y la clave: %v", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil

}
