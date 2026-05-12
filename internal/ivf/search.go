package ivf

import (
	"sort"
	"sync"

	"rinha26/internal/consts"
	"rinha26/internal/quantize"
)

func (idx *Index) FraudScore(query [consts.Dim]float64, nProbeFast, nProbeFull int) int {
	var queryI16 [consts.Dim]int16
	for d := 0; d < consts.Dim; d++ {
		queryI16[d] = quantize.EncodeFloat(query[d])
	}

	dists := centroidDistPool.Get().([]float64)
	defer centroidDistPool.Put(dists)

	idx.computeCentroidDistances(query, dists)

	top := topNIndices(dists, nProbeFast)

	topK := &TopK{}
	idx.scanClusters(top, queryI16, topK)

	fraudCount := topK.FraudCount()
	if fraudCount == 2 || fraudCount == 3 {
		topFull := topNIndices(dists, nProbeFull)
		topK.Reset()
		idx.scanClusters(topFull, queryI16, topK)
		fraudCount = topK.FraudCount()
	}

	return fraudCount
}

func (idx *Index) computeCentroidDistances(query [consts.Dim]float64, dst []float64) {
	k := idx.K()
	norms := idx.CentroidNormsData()
	centroids := idx.CentroidsF64Data()
	for c := 0; c < k; c++ {
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
		dst[c] = norms[c] - 2*dot
	}
}

func (idx *Index) scanClusters(clusters []int, queryI16 [consts.Dim]int16, topK *TopK) {
	offsets := idx.OffsetsData()
	blocks := idx.BlocksData()
	labels := idx.LabelsData()

	for _, c := range clusters {
		start := offsets[c]
		end := offsets[c+1]
		for b := start; b < end; b++ {
			blockStart := int(b) * consts.BlockBytes / 2
			labelStart := int(b) * consts.BlockSize

			for v := 0; v < consts.BlockSize; v++ {
				d0 := int64(queryI16[0]) - int64(blocks[blockStart+0*consts.BlockSize+v])
				d1 := int64(queryI16[1]) - int64(blocks[blockStart+1*consts.BlockSize+v])
				d2 := int64(queryI16[2]) - int64(blocks[blockStart+2*consts.BlockSize+v])
				d3 := int64(queryI16[3]) - int64(blocks[blockStart+3*consts.BlockSize+v])
				d4 := int64(queryI16[4]) - int64(blocks[blockStart+4*consts.BlockSize+v])
				d5 := int64(queryI16[5]) - int64(blocks[blockStart+5*consts.BlockSize+v])
				d6 := int64(queryI16[6]) - int64(blocks[blockStart+6*consts.BlockSize+v])
				d7 := int64(queryI16[7]) - int64(blocks[blockStart+7*consts.BlockSize+v])
				d8 := int64(queryI16[8]) - int64(blocks[blockStart+8*consts.BlockSize+v])
				d9 := int64(queryI16[9]) - int64(blocks[blockStart+9*consts.BlockSize+v])
				d10 := int64(queryI16[10]) - int64(blocks[blockStart+10*consts.BlockSize+v])
				d11 := int64(queryI16[11]) - int64(blocks[blockStart+11*consts.BlockSize+v])
				d12 := int64(queryI16[12]) - int64(blocks[blockStart+12*consts.BlockSize+v])
				d13 := int64(queryI16[13]) - int64(blocks[blockStart+13*consts.BlockSize+v])
				dist := d0*d0 + d1*d1 + d2*d2 + d3*d3 + d4*d4 + d5*d5 + d6*d6 +
					d7*d7 + d8*d8 + d9*d9 + d10*d10 + d11*d11 + d12*d12 + d13*d13

				topK.Push(uint32(dist), labels[labelStart+v] == 1)
			}
		}
	}
}

var centroidDistPool = sync.Pool{
	New: func() interface{} {
		return make([]float64, consts.K)
	},
}

type centroidEntry struct {
	dist float64
	idx  int
}

var centroidEntriesPool = sync.Pool{
	New: func() interface{} {
		return make([]centroidEntry, consts.K)
	},
}

func topNIndices(dists []float64, n int) []int {
	entries := centroidEntriesPool.Get().([]centroidEntry)
	defer centroidEntriesPool.Put(entries)

	for i := range dists {
		entries[i].dist = dists[i]
		entries[i].idx = i
	}

	sort.Slice(entries[:n], func(i, j int) bool {
		return entries[i].dist < entries[j].dist
	})

	for i := n; i < len(dists); i++ {
		if entries[i].dist < entries[n-1].dist {
			pos := n - 1
			for pos > 0 && entries[pos-1].dist > entries[i].dist {
				pos--
			}
			for j := n - 1; j > pos; j-- {
				entries[j] = entries[j-1]
			}
			entries[pos] = entries[i]
		}
	}

	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = entries[i].idx
	}
	return result
}
