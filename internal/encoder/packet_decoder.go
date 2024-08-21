package encoder

import (
	"mccoy.space/g/ogg"
)

// PacketDecoder is a struct that wraps an ogg.Decoder to easily read individual ogg packets.
type PacketDecoder struct {
	D                *ogg.Decoder
	currentPage      ogg.Page
	currentPacketIdx int
}

func NewPacketDecoder(d *ogg.Decoder) *PacketDecoder {
	return &PacketDecoder{
		D: d,
	}
}

func (p *PacketDecoder) Decode() (packet []byte, newPage ogg.Page, err error) {
	if len(p.currentPage.Packets) == 0 {
		p.currentPage, err = p.D.Decode()
		if err != nil {
			return
		}
	}

	if p.currentPacketIdx < len(p.currentPage.Packets) {
		packet = p.currentPage.Packets[p.currentPacketIdx]
		p.currentPacketIdx++
		return
	}

	// If we've exhausted the current page's packets, get the next page
	p.currentPage, err = p.D.Decode()
	if err != nil {
		return
	}

	// Reset the current packet index and return the first packet from the new page
	p.currentPacketIdx = 1
	packet = p.currentPage.Packets[0]
	return
}
