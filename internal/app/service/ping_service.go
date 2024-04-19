package service

// PingService define la interfaz para el servicio de ping.
type PingService interface {
	Ping() string
}

// PingServiceImpl es una implementación concreta de PingService.
type PingServiceImpl struct{}

func (s *PingServiceImpl) Ping() string {
	return "Ping service activated!"
}
