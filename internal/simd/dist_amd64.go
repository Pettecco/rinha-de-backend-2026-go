//go:build amd64

package simd

/*
#cgo CFLAGS: -mavx2 -O3
#include "dist.h"
*/
import "C"

import "unsafe"

type Block [112]int16

type Distances [8]int64

func DistBlock(query *[16]int32, block *Block, out *Distances, threshold int64) {
	C.dist_block(
		(*C.int32_t)(unsafe.Pointer(query)),
		(*C.int16_t)(unsafe.Pointer(block)),
		(*C.int64_t)(unsafe.Pointer(out)),
		(C.int64_t)(threshold),
	)
}
