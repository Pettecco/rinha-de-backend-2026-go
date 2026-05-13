package simd

type Block [112]int16

type Distances [8]int64

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
