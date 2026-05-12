package ivf

import (
	"math"
	"sort"
	"sync"

	"rinha26/internal/consts"
)

func (idx *Index) FraudScore(query [consts.Dim]float64, nProbeFast, nProbeFull int) int {
	var queryI16 [consts.Dim]int32
	for d := 0; d < consts.Dim; d++ {
		queryI16[d] = int32(math.Round(query[d] * consts.Scale))
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
		var dot float64
		base := c * consts.Dim
		for d := 0; d < consts.Dim; d++ {
			dot += centroids[base+d] * query[d]
		}
		dst[c] = norms[c] - 2*dot
	}
}

func (idx *Index) scanClusters(clusters []int, queryI16 [consts.Dim]int32, topK *TopK) {
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
				var dist int64
				for d := 0; d < consts.Dim; d++ {
					val := int64(blocks[blockStart+d*consts.BlockSize+v])
					diff := int64(queryI16[d]) - val
					dist += diff * diff
				}
				isFraud := labels[labelStart+v] == 1
				topK.Push(uint32(dist), isFraud)
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
