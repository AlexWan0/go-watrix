package wavelettree

import (
	"github.com/hillbig/rsdic"
)

// WaveletMatrixBuilder builds WaveletTree from intergaer array.
// A user calls PushBack()s followed by Build().
type WaveletMatrixBuilder struct {
	vals []uint64
	dim  uint64
}

// NewBuilder returns Builder
func NewBuilder() *WaveletMatrixBuilder {
	return &WaveletMatrixBuilder{
		vals: make([]uint64, 0),
	}
}

// PushBack append a value to the builder.
func (wmb *WaveletMatrixBuilder) PushBack(val uint64) {
	wmb.vals = append(wmb.vals, val)
	if val >= wmb.dim {
		wmb.dim = val + 1
	}
}

// Build constructs WaveletMatrix data structure
func (wmb *WaveletMatrixBuilder) Build() *WaveletMatrix {
	blen := getBinaryLen(wmb.dim)
	num := len(wmb.vals)
	zeros := wmb.vals
	ones := make([]uint64, 0, num)
	layers := make([]rsdic.RSDic, blen)
	for depth := uint64(0); depth < blen; depth++ {
		nextZeros := make([]uint64, 0, num)
		nextOnes := make([]uint64, 0, num)
		rsd := rsdic.New()
		filter(zeros, blen-depth-1, &nextZeros, &nextOnes, rsd)
		filter(ones, blen-depth-1, &nextZeros, &nextOnes, rsd)
		zeros = nextZeros
		ones = nextOnes
		layers[depth] = *rsd
	}
	return &WaveletMatrix{layers, wmb.dim, uint64(num), blen}
}

func filter(vals []uint64, depth uint64, nextZeros *[]uint64, nextOnes *[]uint64, rsd *rsdic.RSDic) {
	for _, val := range vals {
		bit := ((val >> depth) & 1) == 1
		rsd.PushBack(bit)
		if bit {
			*nextOnes = append(*nextOnes, val)
		} else {
			*nextZeros = append(*nextZeros, val)
		}
	}
}

// func getDim(vals []uint64) uint64 {
// 	dim := uint64(0)
// 	for _, val := range vals {
// 		if val >= dim {
// 			dim = val + 1
// 		}
// 	}
// 	return dim
// }

func getBinaryLen(val uint64) uint64 {
	blen := uint64(0)
	for val > 0 {
		val >>= 1
		blen++
	}
	return blen
}
