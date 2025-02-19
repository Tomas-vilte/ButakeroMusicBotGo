package ports

type Decoder interface {
	OpusFrame() ([]byte, error)
	Close() error
}
