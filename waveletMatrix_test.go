package wavelettree

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"sort"
	"testing"
)

func generateRange(num uint64) Range {
	bpos := uint64(rand.Intn(int(num)))
	epos := bpos + uint64(rand.Intn(int(num-bpos)))
	return Range{bpos, epos}
}

type uint64Slice []uint64

func (wt uint64Slice) Len() int {
	return len(wt)
}

func (wt uint64Slice) Swap(i, j int) {
	wt[i], wt[j] = wt[j], wt[i]
}

func (wt uint64Slice) Less(i, j int) bool {
	return wt[i] < wt[j]
}

func origIntersect(orig []uint64, ranges []Range, k int) []uint64 {
	cand := make(map[uint64]int)
	for _, ranze := range ranges {
		set := make(map[uint64]struct{})
		for i := ranze.Bpos; i < ranze.Epos; i++ {
			set[orig[i]] = struct{}{}
		}
		for v, _ := range set {
			cand[v]++
		}
	}
	ret := make([]uint64, 0)
	for key, val := range cand {
		if val >= k {
			ret = append(ret, key)
		}
	}
	sort.Sort(uint64Slice(ret))
	return ret
}

func buildWaveletHelper(t *testing.T, num uint64, testNum uint64, dim uint64, orig []uint64, ranks, ranksLessThan, ranksMoreThan [][]uint64) WaveletTree {
	wmb := NewBuilder()
	for i := 0; i < len(ranks); i++ {
		ranks[i] = make([]uint64, num)
		ranksLessThan[i] = make([]uint64, num)
		ranksMoreThan[i] = make([]uint64, num)
	}
	freqs := make([]uint64, dim)
	for i := uint64(0); i < num; i++ {
		x := uint64(rand.Int31n(int32(dim)))
		orig[i] = x
		wmb.PushBack(x)
		for j := uint64(0); j < dim; j++ {
			ranks[j][i] = freqs[j]
			for k := uint64(0); k < j; k++ {
				ranksLessThan[j][i] += freqs[k]
			}
			ranksMoreThan[j][i] = i - ranks[j][i] - ranksLessThan[j][i]
		}
		freqs[x]++
	}
	return wmb.Build()
}

func testWaveletHelper(t *testing.T, wm WaveletTree, num uint64, testNum uint64, dim uint64, orig []uint64, ranks, ranksLessThan, ranksMoreThan [][]uint64) {
	So(wm.Num(), ShouldEqual, num)
	So(wm.Select(num, 0), ShouldEqual, num) // equals num: Not Found
	for i := uint64(0); i < testNum; i++ {
		ind := uint64(rand.Int31n(int32(num)))
		x := uint64(rand.Int31n(int32(dim)))

		So(wm.Lookup(ind), ShouldEqual, orig[ind])

		So(wm.Rank(ind, x), ShouldEqual, ranks[x][ind])
		So(wm.RangedRankRange(Range{0, ind}, Range{x, x + 1}), ShouldEqual, ranks[x][ind])

		So(wm.RankLessThan(ind, x), ShouldEqual, ranksLessThan[x][ind])
		So(wm.RangedRankRange(Range{0, ind}, Range{0, x}), ShouldEqual, ranksLessThan[x][ind])

		So(wm.RankMoreThan(ind, x), ShouldEqual, ranksMoreThan[x][ind])
		So(wm.RangedRankRange(Range{0, ind}, Range{x + 1, dim}), ShouldEqual, ranksMoreThan[x][ind])

		c, rank := wm.LookupAndRank(ind)
		So(c, ShouldEqual, orig[ind])
		So(rank, ShouldEqual, ranks[c][ind])
		So(wm.Select(rank, c), ShouldEqual, ind)

		ranges := make([]Range, 0)
		for j := 0; j < 4; j++ {
			ranges = append(ranges, generateRange(num))
		}
		So(wm.Intersect(ranges, 4), ShouldResemble, origIntersect(orig, ranges, 4))

		ranze := generateRange(num)
		k := uint64(rand.Int63()) % (ranze.Epos - ranze.Bpos)
		vs := make([]int, ranze.Epos-ranze.Bpos)
		for i := uint64(0); i < uint64(len(vs)); i++ {
			vs[i] = int(orig[i+ranze.Bpos])
		}
		sort.Ints(vs)
		So(wm.Quantile(ranze, k), ShouldEqual, vs[k])
	}
	Convey("when op is wrong", func() {
		So(wm.RangedRankOp(Range{0, num}, 0, OpMax), ShouldEqual, 0)
	})
	// Convey("when range is wrong", func() {
	// 	So(wm.RangedRankOp(Range{num, 0}, 0, OpEqual), ShouldEqual, 0) // NOT Supported
	// })
}

func TestWaveletMatrix(t *testing.T) {
	Convey("When a vector is empty", t, func() {
		b := NewBuilder()
		wm := b.Build()
		Convey("The num should be 0", func() {
			So(wm.Num(), ShouldEqual, 0)
			So(wm.Dim(), ShouldEqual, 0)
			So(wm.Rank(0, 0), ShouldEqual, 0)
			So(wm.RankLessThan(0, 0), ShouldEqual, 0)
			So(wm.RankMoreThan(0, 0), ShouldEqual, 0)
			So(wm.RangedRankOp(Range{0, 0}, 0, OpEqual), ShouldEqual, 0)
			So(wm.RangedRankRange(Range{0, 0}, Range{0, 0}), ShouldEqual, 0)
			So(wm.Select(0, 0), ShouldEqual, 0) // equals num: Not Found
		})
	})
	Convey("When a random bit vector is generated", t, func() {
		num := uint64(14000)
		dim := uint64(100)
		testNum := uint64(10)
		orig := make([]uint64, num)
		ranks := make([][]uint64, dim)
		ranksLessThan := make([][]uint64, dim)
		ranksMoreThan := make([][]uint64, dim)

		wm := buildWaveletHelper(t, num, testNum, dim, orig, ranks, ranksLessThan, ranksMoreThan)
		testWaveletHelper(t, wm, num, testNum, dim, orig, ranks, ranksLessThan, ranksMoreThan)
	})
	Convey("When a random bit vector is marshaled", t, func() {
		num := uint64(14000)
		dim := uint64(5)
		testNum := uint64(10)
		orig := make([]uint64, num)
		ranks := make([][]uint64, dim)
		ranksLessThan := make([][]uint64, dim)
		ranksMoreThan := make([][]uint64, dim)

		wmbefore := buildWaveletHelper(t, num, testNum, dim, orig, ranks, ranksLessThan, ranksMoreThan)

		out, err := wmbefore.MarshalBinary()
		So(err, ShouldBeNil)
		wm := New()
		err = wm.UnmarshalBinary(out)
		So(err, ShouldBeNil)

		testWaveletHelper(t, wm, num, testNum, dim, orig, ranks, ranksLessThan, ranksMoreThan)
	})
}

const (
	// N = 10000000 // 10M 10^7
	N = 1000000 // 1M 10^6
	// N = 1 << 20 // 1 Mi * 64 bit = 8 MiB
)

type benchFixture struct {
	builder Builder
	wt      WaveletTree
	counter map[uint64]uint64
	vals    []uint64
}

var bf *benchFixture = nil

func initBenchFixture(b *testing.B) {
	bf = &benchFixture{
		builder: NewBuilder(),
		wt:      nil, // nil at this time
		counter: make(map[uint64]uint64),
		vals:    make([]uint64, 0),
	}

	for i := uint64(0); i < N; i++ {
		x := uint64(rand.Int63())
		bf.counter[x]++
		bf.builder.PushBack(x)
		bf.vals = append(bf.vals, x)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.wt = bf.builder.Build()
	}
	fmt.Printf("{N = %v is used in the tests below}\n\t\t\t\t", N)
}

func BenchmarkWTBuild(b *testing.B) {
	initBenchFixture(b)
}

func BenchmarkWTLookup(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		bf.wt.Lookup(ind)
	}
}

func BenchmarkWTRank(b *testing.B) {
	dim := bf.wt.Dim()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		x := uint64(rand.Int63()) % dim
		bf.wt.Rank(ind, x)
	}
}

func BenchmarkWTRankLessThan(b *testing.B) {
	dim := bf.wt.Dim()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		x := uint64(rand.Int63()) % dim
		bf.wt.RankLessThan(ind, x)
	}
}

func BenchmarkWTSelect(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := bf.vals[uint64(rand.Int63())%uint64(len(bf.vals))]
		rank := uint64(rand.Int63()) % bf.counter[x]
		bf.wt.Select(rank, x)
	}
}

func BenchmarkWTQuantile(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ranze := generateRange(N)
		if ranze.Epos-ranze.Bpos == 0 {
			continue
		}
		k := uint64(rand.Int()) % (ranze.Epos - ranze.Bpos)
		bf.wt.Quantile(ranze, k)
	}
}

func BenchmarkRawLookup(b *testing.B) {
	b.ResetTimer()
	dummy := uint64(0)
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		dummy += bf.vals[ind]
	}
}

func BenchmarkRawRank(b *testing.B) {
	vs := make([]uint64, N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		x := uint64(rand.Int63())
		count := 0
		for j := uint64(0); j < ind; j++ {
			if vs[j] == x {
				count++
			}
		}
	}
}

func BenchmarkRawSelect(b *testing.B) {
	// vs := make([]uint64, N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rank := uint64(rand.Int63() % N)
		count := uint64(0)
		for j := uint64(0); j < N; j++ {
			if bf.vals[j] == 0 {
				count++
				if count == rank {
					break
				}
			}
		}
	}
}

func BenchmarkRawQuantile(b *testing.B) {
	vs := make([]int, N)
	b.ResetTimer()
	dummy := 0
	for i := 0; i < b.N; i++ {
		ranze := generateRange(N)
		k := uint64(rand.Int()) % (ranze.Epos - ranze.Bpos)
		target := vs[ranze.Bpos:ranze.Epos]
		sort.Ints(target)
		dummy += target[k]
	}
}
