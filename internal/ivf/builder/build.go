package builder

import (
	"math"
	"rinha26/internal/consts"
)

func groupByCluster(labels []int, k int) [][]int {
	clusters := make([][]int, k)
	for i, c := range labels {
		clusters[c] = append(clusters[c], i)
	}
	return clusters
}

type ClusterBlocks struct {
	Offsets []uint32
	Blocks  []int16
	Labels  []byte
}

func BuildIndex(vectors [][]float64, labels []byte, k int, rng *LCG) (*ClusterBlocks, [][]float64, error) {
	n := len(vectors)
	sampleSize := consts.TrainSamples
	if sampleSize > n {
		sampleSize = n
	}

	sample := make([][]float64, sampleSize)
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = i
	}
	rng.Shuffle(n, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})
	for i := 0; i < sampleSize; i++ {
		sample[i] = vectors[indices[i]]
	}

	centroids := TrainKMeans(sample, k, consts.MaxIter, consts.EarlyStop, rng)

	clusterLabels := assignAll(vectors, centroids)
	clusters := groupByCluster(clusterLabels, k)

	totalBlocks := 0
	blockOffsets := make([]uint32, k+1)
	blockOffsets[0] = 0
	for c := 0; c < k; c++ {
		nBlocks := (len(clusters[c]) + consts.BlockSize - 1) / consts.BlockSize
		blockOffsets[c+1] = blockOffsets[c] + uint32(nBlocks)
		totalBlocks += nBlocks
	}

	totalBlockBytes := totalBlocks * consts.BlockBytes
	blocks := make([]int16, totalBlockBytes/2)
	blockLabels := make([]byte, totalBlocks*consts.BlockSize)

	for c := 0; c < k; c++ {
		members := clusters[c]
		blockStart := int(blockOffsets[c])
		for b := 0; b < (len(members)+consts.BlockSize-1)/consts.BlockSize; b++ {
			blockIdx := blockStart + b
			vecStart := b * consts.BlockSize
			labelStart := blockIdx * consts.BlockSize

			for d := 0; d < consts.Dim; d++ {
				for v := 0; v < consts.BlockSize; v++ {
					idx := vecStart + v
					pos := blockIdx*consts.BlockBytes/2 + d*consts.BlockSize + v
					if idx < len(members) {
						mi := members[idx]
						val := vectors[mi][d]
						blocks[pos] = int16(math.Round(val * consts.Scale))
						blockLabels[labelStart+v] = labels[mi]
					} else {
						blocks[pos] = math.MaxInt16
						blockLabels[labelStart+v] = 0
					}
				}
			}
		}
	}

	return &ClusterBlocks{
		Offsets: blockOffsets,
		Blocks:  blocks,
		Labels:  blockLabels,
	}, centroids, nil
}
