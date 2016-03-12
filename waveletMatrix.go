// Package watrix provides a wavelet matrix (wavelet tree)
// supporting many range-query problems, including rank/select,
// range min/max query, most frequent and percentile query for general array.
package watrix

import (
	"github.com/hillbig/rsdic"
	"github.com/ugorji/go/codec"
)

// Range represents a range [Bpos, Epos)
// only valid for Bpos <= Epos
type Range struct {
	Bpos uint64
	Epos uint64
}

const (
	// OpEqual is used in RangedRankOp()
	OpEqual = iota
	// OpLessThan is used in RangedRankOp()
	OpLessThan
	// OpMoreThan is used in RangedRankOp()
	OpMoreThan
	// OpMax is upper boundary for OpXXXX constants
	OpMax
)

// WaveletMatrix is the core of the library.
type WaveletMatrix struct {
	layers []rsdic.RSDic
	dim    uint64
	num    uint64
	blen   uint64 // =len(layers)
}

// Num return the number of values in T
func (wm *WaveletMatrix) Num() uint64 {
	return wm.num
}

// Dim returns (max. of T[0...Num) + 1)
func (wm *WaveletMatrix) Dim() uint64 {
	return wm.dim
}

// Lookup returns T[pos]
func (wm *WaveletMatrix) Lookup(pos uint64) uint64 {
	val := uint64(0)
	for depth := 0; depth < len(wm.layers); depth++ {
		val <<= 1
		rsd := wm.layers[depth]
		if !rsd.Bit(pos) {
			pos = rsd.Rank(pos, false)
		} else {
			val |= 1
			pos = rsd.ZeroNum() + rsd.Rank(pos, true)
		}
	}
	return val
}

// Rank returns the number of c (== val) in T[0...pos)
func (wm *WaveletMatrix) Rank(pos uint64, val uint64) uint64 {
	return wm.RangedRankOp(Range{0, pos}, val, OpEqual) // Works but disabled for now to keep test cov.
	// ranze := wm.RankRange(Range{0, pos}, val)
	// return ranze.Epos - ranze.Bpos
}

// RankLessThan returns the number of c (< val) in T[0...pos)
func (wm *WaveletMatrix) RankLessThan(pos uint64, val uint64) (rankLessThan uint64) {
	return wm.RangedRankOp(Range{0, pos}, val, OpLessThan)
}

// RankMoreThan returns the number of c (> val) in T[0...pos)
func (wm *WaveletMatrix) RankMoreThan(pos uint64, val uint64) (rankLessThan uint64) {
	return wm.RangedRankOp(Range{0, pos}, val, OpMoreThan)
}

// RangedRankOp returns the number of c that satisfies 'c op val'
// in T[ranze.Bpos, ranze.Epos).
// The op should be one of {OpEaual, OpLessThan, OpMoreThan}.
func (wm *WaveletMatrix) RangedRankOp(ranze Range, val uint64, op int) uint64 {
	rankLessThan := uint64(0)
	rankMoreThan := uint64(0)
	for depth := uint64(0); depth < wm.blen; depth++ {
		bit := getMSB(val, depth, wm.blen)
		rsd := wm.layers[depth]
		if bit {
			if op == OpLessThan {
				rankLessThan += rsd.Rank(ranze.Epos, false) - rsd.Rank(ranze.Bpos, false)
			}
			ranze.Bpos = rsd.ZeroNum() + rsd.Rank(ranze.Bpos, bit)
			ranze.Epos = rsd.ZeroNum() + rsd.Rank(ranze.Epos, bit)
		} else {
			if op == OpMoreThan {
				rankMoreThan += rsd.Rank(ranze.Epos, true) - rsd.Rank(ranze.Bpos, true)
			}
			ranze.Bpos = rsd.Rank(ranze.Bpos, bit)
			ranze.Epos = rsd.Rank(ranze.Epos, bit)
		}
	}
	switch op {
	case OpEqual:
		return ranze.Epos - ranze.Bpos
	case OpLessThan:
		return rankLessThan
	case OpMoreThan:
		return rankMoreThan
	default:
		return 0
	}
}

// RangedRankRange searches T[ranze.Bpos, ranze.Epos) and
// returns the number of c that falls within valueRange
// i.e. [valueRange.Bpos, valueRange.Epos).
func (wm *WaveletMatrix) RangedRankRange(ranze Range, valueRange Range) uint64 {
	end := wm.RangedRankOp(ranze, valueRange.Epos, OpLessThan)
	beg := wm.RangedRankOp(ranze, valueRange.Bpos, OpLessThan)
	return end - beg
}

func (wm *WaveletMatrix) rangedRankIgnoreLSBsHelper(ranze Range, val uint64, ignoreBits uint64) Range {
	for depth := uint64(0); depth+ignoreBits < wm.blen; depth++ {
		bit := getMSB(val, depth, wm.blen)
		rsd := wm.layers[depth]
		if bit {
			ranze.Bpos = rsd.ZeroNum() + rsd.Rank(ranze.Bpos, bit)
			ranze.Epos = rsd.ZeroNum() + rsd.Rank(ranze.Epos, bit)
		} else {
			ranze.Bpos = rsd.Rank(ranze.Bpos, bit)
			ranze.Epos = rsd.Rank(ranze.Epos, bit)
		}
	}
	return ranze
}

// RangedRankIgnoreLSBs searches T[ranze.Bpos, ranze.Epos) and
// returns the number of c that matches the val.
//
// If ignoreBits > 0, ignoreBits-bit portion from LSB are not considered
// for match.
// This behavior is useful for IP address prefix search such as 192.168.10.0/24
// (ignoreBits in this case, is 8).
func (wm *WaveletMatrix) RangedRankIgnoreLSBs(ranze Range, val, ignoreBits uint64) (rank uint64) {
	r := wm.rangedRankIgnoreLSBsHelper(ranze, val, ignoreBits)
	return r.Epos - r.Bpos
}

func (wm *WaveletMatrix) rangedSelectIgnoreLSBsHelper(pos, val, ignoreBits uint64) uint64 {
	for depth := ignoreBits; depth < wm.blen; depth++ {
		bit := getLSB(val, depth)
		rsd := wm.layers[wm.blen-depth-1]
		if bit {
			pos = rsd.Select(pos-rsd.ZeroNum(), bit)
		} else {
			pos = rsd.Select(pos, bit)
		}
	}
	return pos
}

// RangedSelectIgnoreLSBs searches T[ranze.Bpos, ranze.Epos) and
// returns the position of (rank+1)'th c that matches the val.
//
// If ignoreBits > 0, ignoreBits-bit portion from LSB are not considered
// for match.
// This behavior is useful for IP address prefix search such as 192.168.10.0/24
// (ignoreBits in this case, is 8).
func (wm *WaveletMatrix) RangedSelectIgnoreLSBs(ranze Range, rank, val, ignoreBits uint64) uint64 {
	r := wm.rangedRankIgnoreLSBsHelper(ranze, val, ignoreBits)
	pos := r.Bpos + rank
	if r.Epos <= pos {
		return ranze.Epos
	}
	return wm.rangedSelectIgnoreLSBsHelper(pos, val, ignoreBits)
}

// Select returns the position of (rank+1)-th val in T.
// If not found, returns Num().
func (wm *WaveletMatrix) Select(rank uint64, val uint64) uint64 {
	return wm.selectHelper(rank, val, 0, 0)
	// return wm.RangedSelectIgnoreLSBs(Range{0, wm.Num()}, rank, val, 0)
}

func (wm *WaveletMatrix) selectHelper(rank uint64, val uint64, pos uint64, depth uint64) uint64 {
	if depth == wm.blen {
		return pos + rank
	}
	bit := getMSB(val, depth, wm.blen)
	rsd := wm.layers[depth]
	if !bit {
		pos = rsd.Rank(pos, bit)
		rank = wm.selectHelper(rank, val, pos, depth+1)
	} else {
		pos = rsd.ZeroNum() + rsd.Rank(pos, bit)
		rank = wm.selectHelper(rank, val, pos, depth+1) - rsd.ZeroNum()
	}
	return rsd.Select(rank, bit)
}

// RangedSelect is a experimental query
func (wm *WaveletMatrix) RangedSelect(ranze Range, rank uint64, val uint64) uint64 {
	return wm.RangedSelectIgnoreLSBs(ranze, rank, val, 0)
	// pos := wm.Select(rank+wm.Rank(ranze.Bpos, val), val)
	// if pos < ranze.Epos {
	// 	return pos // Found
	// } else {
	// 	return ranze.Epos // Not Found
	// }
}

// LookupAndRank returns T[pos] and Rank(pos, T[pos])
// Faster than Lookup and Rank
func (wm *WaveletMatrix) LookupAndRank(pos uint64) (uint64, uint64) {
	val := uint64(0)
	bpos := uint64(0)
	epos := uint64(pos)
	for depth := uint64(0); depth < wm.blen; depth++ {
		rsd := wm.layers[depth]
		bit := rsd.Bit(epos)
		bpos = rsd.Rank(bpos, bit)
		epos = rsd.Rank(epos, bit)
		val <<= 1
		if bit {
			bpos += rsd.ZeroNum()
			epos += rsd.ZeroNum()
			val |= 1
		}
	}
	return val, epos - bpos
}

// Quantile returns (k+1)th smallest value in T[ranze.Bpos, ranze.Epos]
func (wm *WaveletMatrix) Quantile(ranze Range, k uint64) uint64 {
	val := uint64(0)
	bpos, epos := ranze.Bpos, ranze.Epos
	for depth := 0; depth < len(wm.layers); depth++ {
		val <<= 1
		rsd := wm.layers[depth]
		nzBpos := rsd.Rank(bpos, false)
		nzEpos := rsd.Rank(epos, false)
		nz := nzEpos - nzBpos
		if k < nz {
			bpos = nzBpos
			epos = nzEpos
		} else {
			k -= nz
			val |= 1
			bpos = rsd.ZeroNum() + bpos - nzBpos
			epos = rsd.ZeroNum() + epos - nzEpos
		}
	}
	return val
}

// Intersect returns values that occure at least k ranges
func (wm *WaveletMatrix) Intersect(ranges []Range, k int) []uint64 {
	return wm.intersectHelper(ranges, k, 0, 0)
}

func (wm *WaveletMatrix) intersectHelper(ranges []Range, k int, depth uint64, prefix uint64) []uint64 {
	if depth == wm.blen {
		ret := make([]uint64, 1)
		ret[0] = prefix
		return ret
	}
	rsd := wm.layers[depth]
	zeroRanges := make([]Range, 0)
	oneRanges := make([]Range, 0)
	for _, ranze := range ranges {
		bpos, epos := ranze.Bpos, ranze.Epos
		nzBpos := rsd.Rank(bpos, false)
		nzEpos := rsd.Rank(epos, false)
		noBpos := bpos - nzBpos + rsd.ZeroNum()
		noEpos := epos - nzEpos + rsd.ZeroNum()
		if nzEpos-nzBpos > 0 {
			zeroRanges = append(zeroRanges, Range{nzBpos, nzEpos})
		}
		if noEpos-noBpos > 0 {
			oneRanges = append(oneRanges, Range{noBpos, noEpos})
		}
	}
	ret := make([]uint64, 0)
	if len(zeroRanges) >= k {
		ret = append(ret, wm.intersectHelper(zeroRanges, k, depth+1, prefix<<1)...)
	}
	if len(oneRanges) >= k {
		ret = append(ret, wm.intersectHelper(oneRanges, k, depth+1, (prefix<<1)|1)...)
	}
	return ret
}

// MarshalBinary encodes WaveletTree into a binary form and returns the result.
func (wm *WaveletMatrix) MarshalBinary() (out []byte, err error) {
	var bh codec.MsgpackHandle
	enc := codec.NewEncoderBytes(&out, &bh)
	err = enc.Encode(len(wm.layers))
	if err != nil {
		return
	}
	for i := 0; i < len(wm.layers); i++ {
		err = enc.Encode(wm.layers[i])
		if err != nil {
			return
		}
	}
	err = enc.Encode(wm.dim)
	if err != nil {
		return
	}
	err = enc.Encode(wm.num)
	if err != nil {
		return
	}
	err = enc.Encode(wm.blen)
	if err != nil {
		return
	}
	return
}

// UnmarshalBinary decodes WaveletTree from a binary form generated MarshalBinary
func (wm *WaveletMatrix) UnmarshalBinary(in []byte) (err error) {
	var bh codec.MsgpackHandle
	dec := codec.NewDecoderBytes(in, &bh)
	layerNum := 0
	err = dec.Decode(&layerNum)
	if err != nil {
		return
	}
	wm.layers = make([]rsdic.RSDic, layerNum)
	for i := 0; i < layerNum; i++ {
		wm.layers[i] = *rsdic.New()
		err = dec.Decode(&wm.layers[i])
		if err != nil {
			return
		}
	}
	err = dec.Decode(&wm.dim)
	if err != nil {
		return
	}
	err = dec.Decode(&wm.num)
	if err != nil {
		return
	}
	err = dec.Decode(&wm.blen)
	if err != nil {
		return
	}
	return
}

func getMSB(x uint64, pos uint64, blen uint64) bool {
	return ((x >> (blen - pos - 1)) & 1) == 1
}

func getLSB(val, depth uint64) bool {
	return (val & (1 << depth)) != 0
}
