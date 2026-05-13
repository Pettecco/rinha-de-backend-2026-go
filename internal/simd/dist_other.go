//go:build !amd64

package simd

// Block is the AoSoA layout for 8 reference vectors × 14 dimensions.
type Block [112]int16

// Distances holds 8 squared L2 distance results.
type Distances [8]int64

// DistBlock computes 8 squared L2 distances using scalar fallback.
func DistBlock(query *[16]int32, block *Block, out *Distances, threshold int64) {
	for v := 0; v < 8; v++ {
		var dist int64
		for d := 0; d < 14; d++ {
			diff := int64(query[d]) - int64(block[d*8+v])
			dist += diff * diff
		}
		out[v] = dist
	}
}
