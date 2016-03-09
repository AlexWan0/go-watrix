// Package wavelettree provides a wavelet tree
// supporting many range-query problems, including rank/select, range min/max query, most frequent and percentile query for general array.
package wavelettree

// type RankResult struct {
// 	val  uint64
// 	freq uint64
// }

// // WaveletTree supports several range queries.
// type WaveletTree interface {
// 	Num() uint64
//
// 	Dim() uint64
//
// 	Lookup(pos uint64) uint64
//
// 	Rank(pos uint64, val uint64) uint64
//
// 	RankLessThan(pos uint64, val uint64) uint64
//
// 	RankMoreThan(pos uint64, val uint64) uint64
//
// 	RangedRankOp(ranze Range, val uint64, op int) uint64
//
// 	RangedRankRange(ranze Range, valueRange Range) uint64
//
// 	Select(rank uint64, val uint64) uint64
//
// 	LookupAndRank(pos uint64) (uint64, uint64)
//
// 	Quantile(ranze Range, k uint64) uint64
//
// 	Intersect(ranges []Range, k int) []uint64
//
// 	MarshalBinary() ([]byte, error)
//
// 	UnmarshalBinary([]byte) error
// }

// // Builder builds WaveletTree from intergaer array.
// // A user calls PushBack()s followed by Build().
// type Builder interface {
// 	PushBack(val uint64)
// 	Build() WaveletTree
// }
