package encoder

import "mccoy.space/g/ogg"

// PacketDecoder es una estructura que envuelve un ogg.Decoder para leer fácilmente
// paquetes individuales de Ogg.
type PacketDecoder struct {
	D                *ogg.Decoder // Decodificador de Ogg que procesa los datos de entrada.
	currentPage      ogg.Page     // Página actual de Ogg que contiene los paquetes.
	currentPacketIdx int          // Índice del paquete actual dentro de la página.
}

// NewPacketDecoder crea una nueva instancia de PacketDecoder.
// Recibe un puntero a ogg.Decoder y devuelve un puntero a PacketDecoder.
func NewPacketDecoder(d *ogg.Decoder) *PacketDecoder {
	return &PacketDecoder{
		D: d,
	}
}

// Decode lee y devuelve el siguiente paquete Ogg del decodificador.
// Si se agotan los paquetes de la página actual, intenta leer la siguiente página.
// Devuelve un paquete, la nueva página de Ogg y un error si ocurre.
func (p *PacketDecoder) Decode() (packet []byte, newPage ogg.Page, err error) {
	// Si no hay paquetes en la página actual, se decodifica una nueva página.
	if len(p.currentPage.Packets) == 0 {
		p.currentPage, err = p.D.Decode() // Decodifica la siguiente página.
		if err != nil {
			return // Retorna si hay un error durante la decodificación.
		}
	}

	// Si hay paquetes en la página actual, se devuelve el siguiente paquete.
	if p.currentPacketIdx < len(p.currentPage.Packets) {
		packet = p.currentPage.Packets[p.currentPacketIdx] // Obtiene el paquete actual.
		p.currentPacketIdx++                               // Incrementa el índice del paquete.
		return
	}

	// Si se agotaron los paquetes de la página actual, se decodifica una nueva página.
	p.currentPage, err = p.D.Decode() // Decodifica la siguiente página.
	if err != nil {
		return // Retorna si hay un error durante la decodificación.
	}

	// Reinicia el índice del paquete y devuelve el primer paquete de la nueva página.
	p.currentPacketIdx = 1
	packet = p.currentPage.Packets[0] // Obtiene el primer paquete de la nueva página.
	return
}
