package builder

import (
	"math"
	"runtime"
	"rinha26/internal/consts"
)

func distSqF64(a, b []float64) float64 {
	var sum float64
	for i := 0; i < consts.Dim; i++ {
		d := a[i] - b[i]
		sum += d * d
	}
	return sum
}

func kmeansPlusPlusInit(vectors [][]float64, k int, rng *LCG) [][]float64 {
	n := len(vectors)
	centroids := make([][]float64, k)

	idx := rng.Intn(n)
	centroids[0] = append([]float64(nil), vectors[idx]...)

	dists := make([]float64, n)
	for i := 0; i < n; i++ {
		dists[i] = distSqF64(vectors[i], centroids[0])
	}

	for c := 1; c < k; c++ {
		total := 0.0
		for _, d := range dists {
			total += d
		}
		if total == 0 {
			idx := rng.Intn(n)
			centroids[c] = append([]float64(nil), vectors[idx]...)
			continue
		}

		r := rng.Float64() * total
		acc := 0.0
		for i := 0; i < n; i++ {
			acc += dists[i]
			if acc >= r {
				idx = i
				break
			}
		}
		centroids[c] = append([]float64(nil), vectors[idx]...)

		for i := 0; i < n; i++ {
			d := distSqF64(vectors[i], centroids[c])
			if d < dists[i] {
				dists[i] = d
			}
		}
	}

	return centroids
}

func assignAll(vectors [][]float64, centroids [][]float64) []int {
	n := len(vectors)
	k := len(centroids)
	labels := make([]int, n)
	workers := runtime.NumCPU()
	chunkSize := (n + workers - 1) / workers

	done := make(chan struct{}, workers)
	for w := 0; w < workers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > n {
			end = n
		}
		if start >= n {
			done <- struct{}{}
			continue
		}
		go func(s, e int) {
			for i := s; i < e; i++ {
				best := 0
				bestDist := math.MaxFloat64
				for c := 0; c < k; c++ {
					d := distSqF64(vectors[i], centroids[c])
					if d < bestDist {
						bestDist = d
						best = c
					}
				}
				labels[i] = best
			}
			done <- struct{}{}
		}(start, end)
	}
	for w := 0; w < workers; w++ {
		<-done
	}
	return labels
}

func updateCentroids(vectors [][]float64, labels []int, k int) [][]float64 {
	sums := make([][]float64, k)
	counts := make([]int, k)
	for c := 0; c < k; c++ {
		sums[c] = make([]float64, consts.Dim)
	}

	for i, v := range vectors {
		c := labels[i]
		counts[c]++
		for d := 0; d < consts.Dim; d++ {
			sums[c][d] += v[d]
		}
	}

	centroids := make([][]float64, k)
	for c := 0; c < k; c++ {
		centroids[c] = make([]float64, consts.Dim)
		if counts[c] > 0 {
			inv := 1.0 / float64(counts[c])
			for d := 0; d < consts.Dim; d++ {
				centroids[c][d] = sums[c][d] * inv
			}
		}
	}
	return centroids
}

func TrainKMeans(vectors [][]float64, k, maxIter int, earlyStop float64, rng *LCG) [][]float64 {
	centroids := kmeansPlusPlusInit(vectors, k, rng)

	n := len(vectors)
	for iter := 0; iter < maxIter; iter++ {
		labels := assignAll(vectors, centroids)
		newCentroids := updateCentroids(vectors, labels, k)

		changed := 0
		for c := 0; c < k; c++ {
			if distSqF64(centroids[c], newCentroids[c]) > 1e-10 {
				changed++
			}
		}

		centroids = newCentroids

		if float64(changed)/float64(k) < earlyStop {
			break
		}
		_ = n
	}

	return centroids
}
