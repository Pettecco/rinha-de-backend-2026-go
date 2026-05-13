package ivf

import (
	"sync"

	"rinha26/internal/consts"
	"rinha26/internal/quantize"
)

var distsBufPool = sync.Pool{
	New: func() any {
		buf := make([]float64, consts.K)
		return &buf
	},
}

func (idx *Index) FraudScore(query [consts.Dim]float64, nProbeFast, nProbeFull int) int {
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

	var queryI16 [consts.Dim]int16
	for d := 0; d < consts.Dim; d++ {
		queryI16[d] = quantize.EncodeFloat(query[d])
	}

	bufPtr := distsBufPool.Get().(*[]float64)
	dists := (*bufPtr)[:K]
	defer distsBufPool.Put(bufPtr)

	computeCentroidDistances(query, idx.CentroidsF64Data(), idx.CentroidNormsData(), K, dists)

	fastChosen := pickTopFromDists(dists, K, nProbeFast)
	fastCount := idx.scanClusters(&queryI16, fastChosen)

	if nProbeFull <= nProbeFast || (fastCount != 2 && fastCount != 3) {
		return fastCount
	}

	fullChosen := pickTopFromDists(dists, K, nProbeFull)
	return idx.scanClusters(&queryI16, fullChosen)
}

func computeCentroidDistances(query [consts.Dim]float64, centroids, normsSq []float64, K int, out []float64) {
	for c := 0; c < K; c++ {
		base := c * consts.Dim
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

func (idx *Index) scanClusters(query *[consts.Dim]int16, clusters []uint32) int {
	var topDistances [topK]int64
	var topLabels [topK]uint8
	for j := range topDistances {
		topDistances[j] = 1<<63 - 1
	}
	worstIdx := 0

	blocks := idx.BlocksData()
	labels := idx.LabelsData()
	offsets := idx.OffsetsData()

	for _, clusterID := range clusters {
		blockStart := offsets[clusterID]
		blockEnd := offsets[clusterID+1]
		for blockIdx := int(blockStart); blockIdx < int(blockEnd); blockIdx++ {
			blockOffset := blockIdx * consts.BlockBytes / 2
			labelOffset := blockIdx * consts.BlockSize

			for v := 0; v < consts.BlockSize; v++ {
				d0 := int64(query[0]) - int64(blocks[blockOffset+0*consts.BlockSize+v])
				d1 := int64(query[1]) - int64(blocks[blockOffset+1*consts.BlockSize+v])
				d2 := int64(query[2]) - int64(blocks[blockOffset+2*consts.BlockSize+v])
				d3 := int64(query[3]) - int64(blocks[blockOffset+3*consts.BlockSize+v])
				d4 := int64(query[4]) - int64(blocks[blockOffset+4*consts.BlockSize+v])
				d5 := int64(query[5]) - int64(blocks[blockOffset+5*consts.BlockSize+v])
				d6 := int64(query[6]) - int64(blocks[blockOffset+6*consts.BlockSize+v])
				d7 := int64(query[7]) - int64(blocks[blockOffset+7*consts.BlockSize+v])
				d8 := int64(query[8]) - int64(blocks[blockOffset+8*consts.BlockSize+v])
				d9 := int64(query[9]) - int64(blocks[blockOffset+9*consts.BlockSize+v])
				d10 := int64(query[10]) - int64(blocks[blockOffset+10*consts.BlockSize+v])
				d11 := int64(query[11]) - int64(blocks[blockOffset+11*consts.BlockSize+v])
				d12 := int64(query[12]) - int64(blocks[blockOffset+12*consts.BlockSize+v])
				d13 := int64(query[13]) - int64(blocks[blockOffset+13*consts.BlockSize+v])
				dist := d0*d0 + d1*d1 + d2*d2 + d3*d3 + d4*d4 + d5*d5 + d6*d6 +
					d7*d7 + d8*d8 + d9*d9 + d10*d10 + d11*d11 + d12*d12 + d13*d13

				worstIdx = updateTopK(&topDistances, &topLabels, worstIdx, dist, labels[labelOffset+v])
			}
		}
	}

	frauds := 0
	for j := 0; j < topK; j++ {
		if topLabels[j] == 1 {
			frauds++
		}
	}
	return frauds
}
