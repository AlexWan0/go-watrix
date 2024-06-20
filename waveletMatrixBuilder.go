package wavelettree

import (
	"os"

	rsdic "github.com/AlexWan0/rsdic-mmap"
)

type waveletMatrixBuilder struct {
	vals []uint64
}

func (wmb *waveletMatrixBuilder) PushBack(val uint64) {
	wmb.vals = append(wmb.vals, val)
}

func (wmb *waveletMatrixBuilder) Build(wmPath string) (WaveletTree, error) {
	err := os.Mkdir(wmPath, 0777)
	if err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}

	dim := getDim(wmb.vals)
	blen := getBinaryLen(dim)
	zeros := wmb.vals
	ones := make([]uint64, 0)
	layers := make([]rsdic.RSDic, blen)
	for depth := uint64(0); depth < blen; depth++ {
		nextZeros := make([]uint64, 0)
		nextOnes := make([]uint64, 0)
		rsdPath := getRsdicPath(wmPath, int(depth))
		rsd, err := rsdic.New(rsdPath)
		if err != nil {
			return nil, err
		}
		err = rsd.LoadWriter()
		if err != nil {
			return nil, err
		}
		defer rsd.CloseWriter()
		filter(zeros, blen-depth-1, &nextZeros, &nextOnes, rsd)
		filter(ones, blen-depth-1, &nextZeros, &nextOnes, rsd)
		zeros = nextZeros
		ones = nextOnes
		layers[depth] = *rsd
	}

	wm := &waveletMatrix{layers, wmPath, dim, uint64(len(wmb.vals)), blen}

	err = wm.LoadReaders()
	if err != nil {
		return nil, err
	}
	return wm, nil
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

func getDim(vals []uint64) uint64 {
	dim := uint64(0)
	for _, val := range vals {
		if val >= dim {
			dim = val + 1
		}
	}
	return dim
}

func getBinaryLen(val uint64) uint64 {
	blen := uint64(0)
	for val > 0 {
		val >>= 1
		blen++
	}
	return blen
}
