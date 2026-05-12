package quantize

import (
	"math"
	"rinha26/internal/consts"
)

func EncodeFloat(v float64) int16 {
	return int16(math.Round(v * consts.Scale))
}

func EncodeVec(src [consts.Dim]float64, dst *[consts.Dim]int16) {
	for i := 0; i < consts.Dim; i++ {
		dst[i] = EncodeFloat(src[i])
	}
}

func DistSqRaw(a, b [consts.Dim]int16) int64 {
	var sum int64
	for i := 0; i < consts.Dim; i++ {
		d := int64(a[i]) - int64(b[i])
		sum += d * d
	}
	return sum
}

func DistSqF64(a, b [consts.Dim]float64) float64 {
	var sum float64
	for i := 0; i < consts.Dim; i++ {
		d := a[i] - b[i]
		sum += d * d
	}
	return sum
}
