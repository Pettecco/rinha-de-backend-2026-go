//go:build amd64

package simd

/*
#include <stdint.h>
#include <immintrin.h>

void dist_block(const int32_t *query, const int16_t *block, int64_t *out, int64_t threshold) {
	__m256i acc_lo = _mm256_setzero_si256();
	__m256i acc_hi = _mm256_setzero_si256();

	for (int d = 0; d < 14; d++) {
		__m128i refs16 = _mm_loadu_si128((const __m128i *)(block + d * 8));
		__m256i refs32 = _mm256_cvtepi16_epi32(refs16);
		__m256i q = _mm256_set1_epi32(query[d]);
		__m256i diff = _mm256_sub_epi32(refs32, q);
		__m256i sq32 = _mm256_mullo_epi32(diff, diff);
		__m256i sq_lo = _mm256_cvtepi32_epi64(_mm256_castsi256_si128(sq32));
		__m256i sq_hi = _mm256_cvtepi32_epi64(_mm256_extracti128_si256(sq32, 1));
		acc_lo = _mm256_add_epi64(acc_lo, sq_lo);
		acc_hi = _mm256_add_epi64(acc_hi, sq_hi);
	}

	_mm256_storeu_si256((__m256i *)(out + 0), acc_lo);
	_mm256_storeu_si256((__m256i *)(out + 4), acc_hi);
}
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
