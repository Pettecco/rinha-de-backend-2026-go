package ivf

import (
	"encoding/binary"
	"io"
	"math"
	"rinha26/internal/consts"
)

func WriteIndex(w io.Writer, n uint32, centroids [][]float64, cb *ClusterBlocks) error {
	header := make([]byte, HeaderSize)
	h := &Header{
		N:   n,
		Dim: uint16(consts.Dim),
		K:   uint16(consts.K),
	}
	copy(h.Magic[:], Magic)
	h.Version = Version
	h.Scale = uint16(consts.Scale)
	h.Write(header)
	if _, err := w.Write(header); err != nil {
		return err
	}

	centBytes := make([]byte, len(centroids)*consts.Dim*4)
	off := 0
	for _, c := range centroids {
		for _, v := range c {
			binary.LittleEndian.PutUint32(centBytes[off:off+4], math.Float32bits(float32(v)))
			off += 4
		}
	}
	if _, err := w.Write(centBytes); err != nil {
		return err
	}

	offsetBytes := make([]byte, len(cb.Offsets)*4)
	for i, o := range cb.Offsets {
		binary.LittleEndian.PutUint32(offsetBytes[i*4:i*4+4], o)
	}
	if _, err := w.Write(offsetBytes); err != nil {
		return err
	}

	blockBytes := make([]byte, len(cb.Blocks)*2)
	for i, v := range cb.Blocks {
		binary.LittleEndian.PutUint16(blockBytes[i*2:i*2+2], uint16(v))
	}
	if _, err := w.Write(blockBytes); err != nil {
		return err
	}

	if _, err := w.Write(cb.Labels); err != nil {
		return err
	}

	return nil
}
