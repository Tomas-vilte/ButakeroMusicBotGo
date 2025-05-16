package interfaces

// Decoder define una interfaz para decodificar audio descargado a un formato que Discord entiende.
type Decoder interface {
	// OpusFrame devuelve un marco de audio en formato Opus.
	OpusFrame() ([]byte, error)
	// Close cierra el decodificador.
	Close() error
}
