#ifndef DIST_H
#define DIST_H

#include <stdint.h>

void dist_block(const int32_t *query, const int16_t *block, int64_t *out, int64_t threshold);

#endif
