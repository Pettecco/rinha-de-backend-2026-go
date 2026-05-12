package ivf

import (
	"encoding/binary"
)

const Magic = "RIVF0004"
const Version = 4
const HeaderSize = 64

type Header struct {
	Magic   [8]byte
	Version uint32
	N       uint32
	Dim     uint16
	K       uint16
	Scale   uint16
	_       [28]byte
}

func (h *Header) Write(buf []byte) {
	copy(buf[0:8], Magic)
	binary.LittleEndian.PutUint32(buf[8:12], Version)
	binary.LittleEndian.PutUint32(buf[12:16], h.N)
	binary.LittleEndian.PutUint16(buf[16:18], h.Dim)
	binary.LittleEndian.PutUint16(buf[18:20], h.K)
	binary.LittleEndian.PutUint16(buf[20:22], h.Scale)
}

func ReadHeader(buf []byte) (*Header, error) {
	if string(buf[0:8]) != Magic {
		return nil, ErrBadMagic
	}
	h := &Header{
		Version: binary.LittleEndian.Uint32(buf[8:12]),
		N:       binary.LittleEndian.Uint32(buf[12:16]),
		Dim:     binary.LittleEndian.Uint16(buf[16:18]),
		K:       binary.LittleEndian.Uint16(buf[18:20]),
		Scale:   binary.LittleEndian.Uint16(buf[20:22]),
	}
	return h, nil
}
