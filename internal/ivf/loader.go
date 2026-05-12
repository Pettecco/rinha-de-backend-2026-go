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
		var norm float64
		for d := 0; d < consts.Dim; d++ {
			v := float64(idx.Centroids[c*consts.Dim+d])
			idx.CentroidsF64[c*consts.Dim+d] = v
			norm += v * v
		}
		idx.CentroidNorms[c] = norm
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

func (idx *Index) CentroidData() []float32 {
	return idx.Centroids
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
