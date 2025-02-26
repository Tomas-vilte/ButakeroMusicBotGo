package ports

// Decoder define una interfaz para decodificar audio descargado a un formato que Discord entiende.
type Decoder interface {
	OpusFrame() ([]byte, error)
	Close() error
}
