package ivf

import (
	"sync"
	"unsafe"

	"rinha26/internal/consts"
	"rinha26/internal/quantize"
	"rinha26/internal/simd"
	"rinha26/internal/vector"
)

var distsBufPool = sync.Pool{
	New: func() any {
		buf := make([]float64, consts.K)
		return &buf
	},
}

func (idx *Index) FraudScore(query [vector.Dim]float64, nProbeFast, nProbeFull int) int {
	K := idx.K()
	if nProbeFast <= 0 {
		nProbeFast = 1
	}
	if nProbeFast > K {
		nProbeFast = K
	}
	if nProbeFull > K {
		nProbeFull = K
	}

	var queryI32 [16]int32
	for d := 0; d < vector.Dim; d++ {
		queryI32[d] = int32(quantize.EncodeFloat(query[d]))
	}

	bufPtr := distsBufPool.Get().(*[]float64)
	dists := (*bufPtr)[:K]
	defer distsBufPool.Put(bufPtr)

	computeCentroidDistances(query, idx.CentroidsF64Data(), idx.CentroidNormsData(), K, dists)

	fastChosen := pickTopFromDists(dists, K, nProbeFast)
	fastCount := idx.scanClusters(&queryI32, fastChosen)

	if nProbeFull <= nProbeFast || (fastCount != 2 && fastCount != 3) {
		return fastCount
	}

	fullChosen := pickTopFromDists(dists, K, nProbeFull)
	return idx.scanClusters(&queryI32, fullChosen)
}

func computeCentroidDistances(query [vector.Dim]float64, centroids, normsSq []float64, K int, out []float64) {
	for c := 0; c < K; c++ {
		base := c * vector.Dim
		dot := centroids[base]*query[0] +
			centroids[base+1]*query[1] +
			centroids[base+2]*query[2] +
			centroids[base+3]*query[3] +
			centroids[base+4]*query[4] +
			centroids[base+5]*query[5] +
			centroids[base+6]*query[6] +
			centroids[base+7]*query[7] +
			centroids[base+8]*query[8] +
			centroids[base+9]*query[9] +
			centroids[base+10]*query[10] +
			centroids[base+11]*query[11] +
			centroids[base+12]*query[12] +
			centroids[base+13]*query[13]
		out[c] = normsSq[c] - 2.0*dot
	}
}

func (idx *Index) scanClusters(query *[16]int32, clusters []uint32) int {
	var topDistances [vector.TopK]int64
	var topLabels [vector.TopK]uint8
	for j := range topDistances {
		topDistances[j] = 1<<63 - 1
	}
	worstIdx := 0

	blocks := idx.BlocksData()
	labels := idx.LabelsData()
	offsets := idx.OffsetsData()

	var blockDist simd.Distances

	for _, clusterID := range clusters {
		blockStart := offsets[clusterID]
		blockEnd := offsets[clusterID+1]
		for blockIdx := int(blockStart); blockIdx < int(blockEnd); blockIdx++ {
			blockOffset := blockIdx * consts.BlockBytes / 2
			labelOffset := blockIdx * consts.BlockSize

			simd.DistBlock(query,
				(*simd.Block)(unsafe.Pointer(&blocks[blockOffset])),
				&blockDist,
				topDistances[worstIdx])

			for v := 0; v < consts.BlockSize; v++ {
				worstIdx = updateTopK(&topDistances, &topLabels, worstIdx, blockDist[v], labels[labelOffset+v])
			}
		}
	}

	frauds := 0
	for j := 0; j < vector.TopK; j++ {
		if topLabels[j] == 1 {
			frauds++
		}
	}
	return frauds
}
