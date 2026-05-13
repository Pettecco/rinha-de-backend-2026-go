package ivf

import (
	"math"

	"rinha26/internal/vector"
)

func pickTopFromDists(distances []float64, K, nProbe int) []uint32 {
	var chosen [28]uint32
	var chosenDistances [28]float64
	worst := math.MaxFloat64
	worstIdx := 0
	count := 0

	for c := 0; c < K; c++ {
		d := distances[c]
		if count < nProbe {
			chosen[count] = uint32(c)
			chosenDistances[count] = d
			count++
			if count == nProbe {
				worstIdx = indexOfMax(chosenDistances[:nProbe])
				worst = chosenDistances[worstIdx]
			}
			continue
		}
		if d < worst {
			chosen[worstIdx] = uint32(c)
			chosenDistances[worstIdx] = d
			worstIdx = indexOfMax(chosenDistances[:nProbe])
			worst = chosenDistances[worstIdx]
		}
	}
	return chosen[:count]
}

func indexOfMax(xs []float64) int {
	idx := 0
	for i := 1; i < len(xs); i++ {
		if xs[i] > xs[idx] {
			idx = i
		}
	}
	return idx
}

func updateTopK(topDistances *[vector.TopK]int64, topLabels *[vector.TopK]uint8, worstIdx int, candidateDist int64, candidateLabel uint8) int {
	if candidateDist >= topDistances[worstIdx] {
		return worstIdx
	}
	topDistances[worstIdx] = candidateDist
	topLabels[worstIdx] = candidateLabel
	newWorst := 0
	for k := 1; k < vector.TopK; k++ {
		if topDistances[k] > topDistances[newWorst] {
			newWorst = k
		}
	}
	return newWorst
}
