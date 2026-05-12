// Package quantize holds int16 quantization helpers used by the IVF index.
package quantize

import (
	"math"

	"rinha26/internal/vector"
)

const Scale int = 10000

func EncodeFloat(v float64) int16 {
	if v <= -0.999 {
		return int16(-Scale)
	}
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	q := int(math.Round(v * float64(Scale)))
	if q > Scale {
		q = Scale
	}
	if q < 0 {
		q = 0
	}
	return int16(q)
}

// EncodeVec quantizes a 14-D float64 vector into a 14-int16 array.
func EncodeVec(in *[vector.Dim]float64, out *[vector.Dim]int16) {
	for i := 0; i < vector.Dim; i++ {
		out[i] = EncodeFloat(in[i])
	}
}

/*
DistSqRaw computes squared L2 distance between a query int16 vec and a raw
int16 slice from mmap
*/
func DistSqRaw(query *[vector.Dim]int16, ref []int16) int64 {
	_ = ref[vector.Dim-1] // bounds check hint
	d0 := int64(query[0]) - int64(ref[0])
	d1 := int64(query[1]) - int64(ref[1])
	d2 := int64(query[2]) - int64(ref[2])
	d3 := int64(query[3]) - int64(ref[3])
	d4 := int64(query[4]) - int64(ref[4])
	d5 := int64(query[5]) - int64(ref[5])
	d6 := int64(query[6]) - int64(ref[6])
	d7 := int64(query[7]) - int64(ref[7])
	d8 := int64(query[8]) - int64(ref[8])
	d9 := int64(query[9]) - int64(ref[9])
	d10 := int64(query[10]) - int64(ref[10])
	d11 := int64(query[11]) - int64(ref[11])
	d12 := int64(query[12]) - int64(ref[12])
	d13 := int64(query[13]) - int64(ref[13])
	return d0*d0 + d1*d1 + d2*d2 + d3*d3 + d4*d4 + d5*d5 + d6*d6 +
		d7*d7 + d8*d8 + d9*d9 + d10*d10 + d11*d11 + d12*d12 + d13*d13
}
