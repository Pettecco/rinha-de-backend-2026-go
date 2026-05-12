package ivf

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"rinha26/internal/consts"
)

type Index struct {
	mmap          []byte
	N             uint32
	Centroids     []float32
	Offsets       []uint32
	Blocks        []int16
	Labels        []byte
	CentroidsF64  []float64
	CentroidNorms []float64
}

func Open(path string) (*Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	size := info.Size()

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap: %w", err)
	}

	h, err := ReadHeader(data[0:HeaderSize])
	if err != nil {
		syscall.Munmap(data)
		return nil, fmt.Errorf("header: %w", err)
	}

	if h.Dim != uint16(consts.Dim) {
		syscall.Munmap(data)
		return nil, fmt.Errorf("dim mismatch: got %d want %d", h.Dim, consts.Dim)
	}

	idx := &Index{
		mmap: data,
		N:    h.N,
	}

	centOff := HeaderSize
	centBytes := int(h.K) * consts.Dim * 4
	idx.Centroids = unsafe.Slice((*float32)(unsafe.Pointer(&data[centOff])), int(h.K)*consts.Dim)

	offOff := centOff + centBytes
	offCount := int(h.K) + 1
	idx.Offsets = unsafe.Slice((*uint32)(unsafe.Pointer(&data[offOff])), offCount)

	blockOff := offOff + offCount*4
	totalBlocks := int(idx.Offsets[h.K])
	blockCount := totalBlocks * consts.BlockBytes / 2
	idx.Blocks = unsafe.Slice((*int16)(unsafe.Pointer(&data[blockOff])), blockCount)

	labelOff := blockOff + totalBlocks*consts.BlockBytes
	idx.Labels = data[labelOff:]

	idx.precomputeCentroids(int(h.K))

	return idx, nil
}

func (idx *Index) precomputeCentroids(k int) {
	idx.CentroidsF64 = make([]float64, k*consts.Dim)
	idx.CentroidNorms = make([]float64, k)
	for c := 0; c < k; c++ {
		base := c * consts.Dim
		v0 := float64(idx.Centroids[base+0])
		v1 := float64(idx.Centroids[base+1])
		v2 := float64(idx.Centroids[base+2])
		v3 := float64(idx.Centroids[base+3])
		v4 := float64(idx.Centroids[base+4])
		v5 := float64(idx.Centroids[base+5])
		v6 := float64(idx.Centroids[base+6])
		v7 := float64(idx.Centroids[base+7])
		v8 := float64(idx.Centroids[base+8])
		v9 := float64(idx.Centroids[base+9])
		v10 := float64(idx.Centroids[base+10])
		v11 := float64(idx.Centroids[base+11])
		v12 := float64(idx.Centroids[base+12])
		v13 := float64(idx.Centroids[base+13])

		idx.CentroidsF64[base+0] = v0
		idx.CentroidsF64[base+1] = v1
		idx.CentroidsF64[base+2] = v2
		idx.CentroidsF64[base+3] = v3
		idx.CentroidsF64[base+4] = v4
		idx.CentroidsF64[base+5] = v5
		idx.CentroidsF64[base+6] = v6
		idx.CentroidsF64[base+7] = v7
		idx.CentroidsF64[base+8] = v8
		idx.CentroidsF64[base+9] = v9
		idx.CentroidsF64[base+10] = v10
		idx.CentroidsF64[base+11] = v11
		idx.CentroidsF64[base+12] = v12
		idx.CentroidsF64[base+13] = v13

		idx.CentroidNorms[c] = v0*v0 + v1*v1 + v2*v2 + v3*v3 + v4*v4 + v5*v5 + v6*v6 +
			v7*v7 + v8*v8 + v9*v9 + v10*v10 + v11*v11 + v12*v12 + v13*v13
	}
}

func (idx *Index) PreTouch() {
	log.Printf("pre-touching %d MB index ...", len(idx.mmap)/1024/1024)
	for i := 0; i < len(idx.mmap); i += 4096 {
		_ = idx.mmap[i]
	}
	runtime.GC()
}

func (idx *Index) Close() error {
	return syscall.Munmap(idx.mmap)
}

func (idx *Index) K() int {
	return len(idx.CentroidNorms)
}

func (idx *Index) OffsetsData() []uint32 {
	return idx.Offsets
}

func (idx *Index) BlocksData() []int16 {
	return idx.Blocks
}

func (idx *Index) LabelsData() []byte {
	return idx.Labels
}

func (idx *Index) CentroidsF64Data() []float64 {
	return idx.CentroidsF64
}

func (idx *Index) CentroidNormsData() []float64 {
	return idx.CentroidNorms
}
